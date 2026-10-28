// Package agent implements the local Linux log reader and durable counters.
package agent

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Rule struct {
	ID      string      `json:"id"`
	Version uint64      `json:"version"`
	Mode    string      `json:"mode"`
	Pattern string      `json:"pattern"`
	Exclude []Exclusion `json:"exclude,omitempty"`
}
type Exclusion struct {
	Mode    string `json:"mode"`
	Pattern string `json:"pattern"`
}
type Target struct {
	WorkloadType string `json:"workloadType"`
	ID           string `json:"id"`
	Namespace    string `json:"namespace"`
	PodUID       string `json:"podUid"`
	PodName      string `json:"podName"`
	InstanceID   string `json:"instanceId"`
	Container    string `json:"container"`
	WorkloadName string `json:"workloadName"`
	Directory    string `json:"directory"`
	Glob         string `json:"glob"`
	Start        string `json:"start"`
	Rules        []Rule `json:"rules"`
	Format       string `json:"format,omitempty"`
}
type Config struct {
	Version uint64   `json:"version"`
	Targets []Target `json:"targets"`
}
type cursor struct {
	Offset    int64
	Counts    map[string]uint64
	LastMatch int64
}
type checkpoint struct {
	Files           map[string]cursor
	Paths           map[string]bool
	Totals          map[string]cursor
	Config          Config
	Samples         map[string]logSample `json:",omitempty"`
	SampleBudget    logSampleBudget      `json:",omitempty"`
	SampleLimitHits uint64               `json:",omitempty"`
}
type Agent struct {
	mu                 sync.Mutex
	root, state        string
	lock               *os.File
	config             Config
	saved              checkpoint
	healthy            map[string]bool
	missing            map[string]bool
	readErrors         map[string]uint64
	lastRead           map[string]int64
	appMetrics         map[string]appMetric
	stateAvailable     bool
	stateResets        uint64
	statePersistErrors uint64
	lastScanIssueLog   time.Time
}

func New(root, state string) (*Agent, error) {
	if !filepath.IsAbs(root) || !filepath.IsAbs(state) {
		return nil, errors.New("absolute directories required")
	}
	a := &Agent{root: root, state: state, saved: checkpoint{Files: map[string]cursor{}, Paths: map[string]bool{}, Totals: map[string]cursor{}}, healthy: map[string]bool{}, missing: map[string]bool{}, readErrors: map[string]uint64{}, lastRead: map[string]int64{}, appMetrics: map[string]appMetric{}}
	lockErr := a.acquireLock()
	if errors.Is(lockErr, errCheckpointOwned) {
		return nil, lockErr
	}
	raw, err := os.ReadFile(filepath.Join(state, "checkpoint.json"))
	if lockErr != nil {
		err = lockErr
	}
	if err == nil && strings.TrimSpace(string(raw)) == "null" {
		err = errors.New("invalid checkpoint")
	}
	if err == nil {
		err = json.Unmarshal(raw, &a.saved)
	} else if os.IsNotExist(err) {
		err = nil
	}
	if err != nil || !validCheckpoint(a.saved) {
		a.saved = checkpoint{Files: map[string]cursor{}, Paths: map[string]bool{}, Totals: map[string]cursor{}}
		a.stateResets++
		log.Printf("checkpoint state is unavailable or invalid; starting with in-memory state")
	} else if !a.stateAvailable {
		a.stateResets++
	}
	if a.saved.Paths == nil {
		a.saved.Paths = map[string]bool{}
	}
	if a.saved.Totals == nil {
		a.saved.Totals = map[string]cursor{}
	}
	pruneLogSamples(&a.saved, time.Now().Unix())
	return a, nil
}

func validCheckpoint(c checkpoint) bool {
	if c.Files == nil || c.Paths == nil {
		return false
	}
	for _, saved := range c.Files {
		if saved.Offset < 0 || saved.Counts == nil || saved.LastMatch < 0 {
			return false
		}
	}
	for _, saved := range c.Totals {
		if saved.Counts == nil || saved.LastMatch < 0 {
			return false
		}
	}
	return true
}

var errCheckpointOwned = errors.New("checkpoint already owned")

func (a *Agent) acquireLock() error {
	if err := os.MkdirAll(a.state, 0700); err != nil {
		return err
	}
	if a.lock != nil {
		return nil
	}
	lock, err := os.OpenFile(filepath.Join(a.state, "lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		lock.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return errCheckpointOwned
		}
		return err
	}
	a.lock = lock
	a.stateAvailable = true
	return nil
}
func (a *Agent) Close() {
	if a.lock != nil {
		a.lock.Close()
	}
}
func (a *Agent) Version() uint64 { a.mu.Lock(); defer a.mu.Unlock(); return a.config.Version }
func validRelative(p string) bool {
	return p != "" && !filepath.IsAbs(p) && p != ".." && !strings.HasPrefix(filepath.Clean(p), "../")
}
func validDirectory(p string) bool {
	if validRelative(p) {
		return true
	}
	if p == "" || !filepath.IsAbs(p) || p == "/" {
		return false
	}
	for _, part := range strings.Split(filepath.Clean(p), string(os.PathSeparator)) {
		if part == ".." {
			return false
		}
	}
	return true
}
func (a *Agent) Apply(c Config) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c.Version == 0 || c.Version < a.config.Version || len(c.Targets) > 1000 {
		return errors.New("invalid configuration version or size")
	}
	ids := map[string]bool{}
	for _, t := range c.Targets {
		if t.ID == "" || ids[t.ID] || t.Namespace == "" || t.PodUID == "" || !validDirectory(t.Directory) || t.Glob == "" || strings.ContainsAny(t.Glob, "/\\") || len(t.Rules) > 100 || (t.Start != "end" && t.Start != "beginning") || (t.Format != "" && t.Format != "otlp_metric") || (t.Format == "otlp_metric" && len(t.Rules) > 0) {
			return errors.New("invalid target")
		}
		ids[t.ID] = true
		if _, err := filepath.Match(t.Glob, ""); err != nil {
			return errors.New("invalid glob")
		}
		rules := map[string]bool{}
		for _, r := range t.Rules {
			if r.ID == "" || rules[r.ID] || r.Version == 0 || len(r.Pattern) == 0 || len(r.Pattern) > 4096 {
				return errors.New("invalid rule")
			}
			rules[r.ID] = true
			if r.Mode == "regex" {
				if _, err := regexp.Compile(r.Pattern); err != nil {
					return errors.New("invalid regex")
				}
			} else if r.Mode != "literal" {
				return errors.New("invalid rule mode")
			}
			if len(r.Exclude) > 100 {
				return errors.New("too many rule exclusions")
			}
			for _, exclude := range r.Exclude {
				if exclude.Pattern == "" || len(exclude.Pattern) > 4096 || (exclude.Mode != "literal" && exclude.Mode != "regex") {
					return errors.New("invalid rule exclusion")
				}
				if exclude.Mode == "regex" {
					if _, err := regexp.Compile(exclude.Pattern); err != nil {
						return errors.New("invalid exclusion regex")
					}
				}
			}
		}
	}
	raw, _ := json.Marshal(c)
	var copy Config
	json.Unmarshal(raw, &copy)
	if c.Version == a.config.Version {
		old, _ := json.Marshal(a.config)
		if string(old) != string(raw) {
			return errors.New("configuration version conflict")
		}
	}
	if c.Version != a.config.Version {
		a.appMetrics = map[string]appMetric{}
		a.healthy = map[string]bool{}
	}
	// A new rule version starts a new counter but continues from the committed
	// byte offset, rather than replaying the file or skipping unread lines.
	previousConfig := a.config
	if previousConfig.Version == 0 {
		previousConfig = a.saved.Config
	}
	for _, target := range copy.Targets {
		for _, previousTarget := range previousConfig.Targets {
			if target.ID != previousTarget.ID || target.PodUID != previousTarget.PodUID || target.Directory != previousTarget.Directory || target.Glob != previousTarget.Glob {
				continue
			}
			for _, rule := range target.Rules {
				for _, previous := range previousTarget.Rules {
					if rule.ID != previous.ID || rule.Version == previous.Version {
						continue
					}
					oldPrefix := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00", target.ID, target.PodUID, rule.ID, previous.Version)
					newPrefix := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00", target.ID, target.PodUID, rule.ID, rule.Version)
					for key, saved := range a.saved.Files {
						if strings.HasPrefix(key, oldPrefix) {
							a.saved.Files[newPrefix+strings.TrimPrefix(key, oldPrefix)] = cursor{Offset: saved.Offset, Counts: map[string]uint64{}}
						}
					}
					for key, seen := range a.saved.Paths {
						if strings.HasPrefix(key, oldPrefix) {
							a.saved.Paths[newPrefix+strings.TrimPrefix(key, oldPrefix)] = seen
						}
					}
				}
			}
		}
	}
	a.config = copy
	allowed := map[string]bool{}
	allowedSamples := map[string]bool{}
	for _, t := range copy.Targets {
		for _, r := range t.Rules {
			prefix := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00", t.ID, t.PodUID, r.ID, r.Version)
			allowed[prefix] = true
			allowedSamples[logSampleKey(t, r)] = true
		}
	}
	for key := range a.saved.Samples {
		if !allowedSamples[key] {
			delete(a.saved.Samples, key)
		}
	}
	for key := range a.saved.Files {
		keep := false
		for prefix := range allowed {
			if strings.HasPrefix(key, prefix) {
				keep = true
				break
			}
		}
		if !keep {
			delete(a.saved.Files, key)
		}
	}
	for key := range a.saved.Paths {
		keep := false
		for prefix := range allowed {
			if strings.HasPrefix(key, prefix) {
				keep = true
				break
			}
		}
		if !keep {
			delete(a.saved.Paths, key)
		}
	}
	for key := range a.saved.Totals {
		if !allowed[key+"\x00"] {
			delete(a.saved.Totals, key)
		}
	}
	return nil
}

// Open each component relative to an already-open directory. O_NOFOLLOW prevents
// symlink swaps between validation and open; only regular files are accepted.
func (a *Agent) open(path string) (*os.File, error) {
	if !validDirectory(path) {
		return nil, errors.New("invalid log directory")
	}
	base := a.root
	relative := path
	if filepath.IsAbs(path) {
		cleanRoot := filepath.Clean(a.root)
		cleanPath := filepath.Clean(path)
		if cleanRoot != "/" {
			relativePath, err := filepath.Rel(cleanRoot, cleanPath)
			if err != nil || relativePath == ".." || strings.HasPrefix(relativePath, "../") {
				return nil, errors.New("log path is outside agent root")
			}
			relative = relativePath
		} else {
			relative = strings.TrimPrefix(cleanPath, string(os.PathSeparator))
		}
		if relative == "." {
			relative = ""
		}
		base = "/"
		if cleanRoot != "/" {
			base = cleanRoot
		}
	}
	fd, err := syscall.Open(base, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(filepath.Clean(relative), string(os.PathSeparator))
	for i, p := range parts {
		flags := syscall.O_RDONLY | syscall.O_NOFOLLOW | syscall.O_NONBLOCK
		if i < len(parts)-1 {
			flags |= syscall.O_DIRECTORY
		}
		next, e := syscall.Openat(fd, p, flags, 0)
		syscall.Close(fd)
		if e != nil {
			return nil, e
		}
		fd = next
	}
	return os.NewFile(uintptr(fd), relative), nil
}
func (a *Agent) Scan() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	raw, _ := json.Marshal(a.saved)
	var next checkpoint
	json.Unmarshal(raw, &next)
	if next.Files == nil {
		next.Files = map[string]cursor{}
	}
	if next.Paths == nil {
		next.Paths = map[string]bool{}
	}
	if next.Totals == nil {
		next.Totals = map[string]cursor{}
	}
	next.Config = a.config
	pruneLogSamples(&next, time.Now().Unix())
	a.appMetrics = map[string]appMetric{}
	seen := map[string]bool{}
	matched := map[string]bool{}
	activeFiles := map[string]map[string]bool{}
	activePaths := map[string]map[string]bool{}
	pruneTarget := map[string]bool{}
	missingTargets, readErrorTargets := 0, 0
	for _, t := range a.config.Targets {
		pruneTarget[t.ID+"\x00"+t.PodUID] = true
		activeFiles[t.ID+"\x00"+t.PodUID] = map[string]bool{}
		activePaths[t.ID+"\x00"+t.PodUID] = map[string]bool{}
		a.healthy[t.ID] = false
		errorsBefore := a.readErrors[t.ID]
		dir, err := a.open(t.Directory)
		if err != nil {
			a.missing[t.ID] = true
			missingTargets++
			pruneTarget[t.ID+"\x00"+t.PodUID] = false
			continue
		}
		entries, err := dir.Readdirnames(10001)
		dir.Close()
		if err != nil && err != io.EOF {
			a.readErrors[t.ID]++
			readErrorTargets++
			pruneTarget[t.ID+"\x00"+t.PodUID] = false
			continue
		}
		if len(entries) > 10000 {
			a.readErrors[t.ID]++
			readErrorTargets++
			pruneTarget[t.ID+"\x00"+t.PodUID] = false
			continue
		}
		for _, name := range entries {
			match, _ := filepath.Match(t.Glob, name)
			if !match || (t.Format == "" && strings.HasSuffix(name, "_metric.log")) {
				continue
			}
			matched[t.ID] = true
			f, err := a.open(filepath.Join(t.Directory, name))
			if err != nil {
				a.readErrors[t.ID]++
				readErrorTargets++
				pruneTarget[t.ID+"\x00"+t.PodUID] = false
				continue
			}
			stat, statErr := f.Stat()
			if statErr != nil {
				a.readErrors[t.ID]++
				readErrorTargets++
				pruneTarget[t.ID+"\x00"+t.PodUID] = false
				f.Close()
				continue
			}
			sys := stat.Sys().(*syscall.Stat_t)
			for _, rule := range t.Rules {
				prefix := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00", t.ID, t.PodUID, rule.ID, rule.Version)
				activeFiles[t.ID+"\x00"+t.PodUID][fmt.Sprintf("%s%d\x00%d", prefix, sys.Dev, sys.Ino)] = true
				activePaths[t.ID+"\x00"+t.PodUID][fmt.Sprintf("%s%s", prefix, f.Name())] = true
			}
			seen[t.ID] = true
			err = a.read(f, t, &next)
			f.Close()
			if err != nil {
				a.readErrors[t.ID]++
				readErrorTargets++
				pruneTarget[t.ID+"\x00"+t.PodUID] = false
				continue
			}
		}
		a.healthy[t.ID] = seen[t.ID] && errorsBefore == a.readErrors[t.ID]
		if !matched[t.ID] {
			a.missing[t.ID] = true
		} else {
			a.missing[t.ID] = false
		}
	}
	if (missingTargets > 0 || readErrorTargets > 0) && time.Since(a.lastScanIssueLog) >= time.Minute {
		log.Printf("log scan summary: target_count=%d missing_target_count=%d read_error_target_count=%d", len(a.config.Targets), missingTargets, readErrorTargets)
		a.lastScanIssueLog = time.Now()
	}
	for key := range next.Files {
		parts := strings.SplitN(key, "\x00", 3)
		if len(parts) < 3 {
			continue
		}
		targetKey := parts[0] + "\x00" + parts[1]
		if !pruneTarget[targetKey] || activeFiles[targetKey][key] {
			continue
		}
		prefixParts := strings.SplitN(key, "\x00", 5)
		if len(prefixParts) < 5 {
			continue
		}
		totalKey := strings.Join(prefixParts[:4], "\x00")
		total := next.Totals[totalKey]
		if total.Counts == nil {
			total.Counts = map[string]uint64{}
		}
		for rule, count := range next.Files[key].Counts {
			total.Counts[rule] += count
		}
		if next.Files[key].LastMatch > total.LastMatch {
			total.LastMatch = next.Files[key].LastMatch
		}
		next.Totals[totalKey] = total
		delete(next.Files, key)
	}
	for key := range next.Paths {
		parts := strings.SplitN(key, "\x00", 3)
		if len(parts) < 3 {
			continue
		}
		targetKey := parts[0] + "\x00" + parts[1]
		if !pruneTarget[targetKey] || activePaths[targetKey][key] {
			continue
		}
		delete(next.Paths, key)
	}
	a.saved = next
	if err := a.persist(); err != nil {
		a.statePersistErrors++
		log.Printf("checkpoint persistence failed; continuing in memory: %v", err)
		a.stateAvailable = false
		return nil
	}
	return nil
}

func (a *Agent) persist() error {
	if err := a.acquireLock(); err != nil {
		return err
	}
	raw, err := json.Marshal(a.saved)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(a.state, ".checkpoint-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(raw); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = os.Rename(f.Name(), filepath.Join(a.state, "checkpoint.json")); err != nil {
		return err
	}
	d, err := os.Open(a.state)
	if err != nil {
		return err
	}
	err = d.Sync()
	d.Close()
	if err != nil {
		return err
	}
	a.stateAvailable = true
	return nil
}
func (a *Agent) read(f *os.File, t Target, next *checkpoint) error {
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	if !stat.Mode().IsRegular() {
		return errors.New("log is not a regular file")
	}
	sys := stat.Sys().(*syscall.Stat_t)
	a.lastRead[t.ID] = time.Now().Unix()
	if t.Format == "otlp_metric" {
		return a.readAppMetrics(f, t)
	}
	for _, r := range t.Rules {
		key := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%d\x00%d", t.ID, t.PodUID, r.ID, r.Version, sys.Dev, sys.Ino)
		c, exists := next.Files[key]
		if !exists {
			c = cursor{Counts: map[string]uint64{}}
			pathKey := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s", t.ID, t.PodUID, r.ID, r.Version, f.Name())
			if t.Start == "end" && !a.saved.Paths[pathKey] {
				c.Offset = stat.Size()
			}
			next.Paths[pathKey] = true
		}
		if stat.Size() < c.Offset {
			c.Offset = 0
		}
		if _, err = f.Seek(c.Offset, io.SeekStart); err != nil {
			return err
		}
		reader := bufio.NewReaderSize(f, 64*1024)
		var sample *logSample
		sampleChecked := false
		var pattern *regexp.Regexp
		if r.Mode == "regex" {
			pattern = regexp.MustCompile(r.Pattern)
		}
		exclusions := make([]*regexp.Regexp, len(r.Exclude))
		for i, exclude := range r.Exclude {
			if exclude.Mode == "regex" {
				exclusions[i] = regexp.MustCompile(exclude.Pattern)
			}
		}
		count := c.Counts[r.ID]
		for n := 0; n < 10000; n++ {
			line, e := reader.ReadSlice('\n')
			if e == io.EOF {
				break
			}
			if e != nil {
				return errors.New("line too long or read failed")
			}
			match := (pattern != nil && pattern.Match(line)) || (pattern == nil && strings.Contains(string(line), r.Pattern))
			if match {
				for i, exclude := range r.Exclude {
					if (exclusions[i] != nil && exclusions[i].Match(line)) || (exclusions[i] == nil && strings.Contains(string(line), exclude.Pattern)) {
						match = false
						break
					}
				}
			}
			if match {
				count++
				c.LastMatch = time.Now().Unix()
				if !sampleChecked {
					sample = newLogSample(next, logSampleKey(t, r), f.Name(), fmt.Sprintf("%d:%d", sys.Dev, sys.Ino), c.Offset, line, c.LastMatch)
					sampleChecked = true
				}
			}
			c.Offset += int64(len(line))
		}
		c.Counts[r.ID] = count
		next.Files[key] = c
		if sample != nil {
			commitLogSample(next, logSampleKey(t, r), *sample)
		}
	}
	return nil
}
func appMetricKey(target string, sample appMetric) string {
	raw, _ := json.Marshal(sample.Labels)
	return target + "\x00" + sample.Name + "\x00" + string(raw)
}
func (a *Agent) Count(target, rule string) uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	var total uint64
	for _, t := range a.config.Targets {
		if t.ID != target {
			continue
		}
		for _, r := range t.Rules {
			if r.ID != rule {
				continue
			}
			prefix := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00", t.ID, t.PodUID, r.ID, r.Version)
			for k, c := range a.saved.Files {
				if strings.HasPrefix(k, prefix) {
					total += c.Counts[rule]
				}
			}
			if totalCursor, ok := a.saved.Totals[strings.TrimSuffix(prefix, "\x00")]; ok {
				total += totalCursor.Counts[rule]
			}
		}
	}
	return total
}
