package funcs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// dateRollingWriter writes logs to files named with the current date.
// It rolls over at midnight by closing the old file and opening a new one.
type dateRollingWriter struct {
	lock    sync.Mutex
	dir     string
	prefix  string
	curDate string
	file    *os.File
	// retentionDays controls how many days of logs to keep.
	retentionDays int
}

func newDateRollingWriter(dir, prefix string) (*dateRollingWriter, error) {
	w := &dateRollingWriter{dir: dir, prefix: prefix, retentionDays: getRetentionDaysFromEnv()}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	if err := w.rotateIfNeeded(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *dateRollingWriter) Write(p []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	if err := w.rotateIfNeeded(); err != nil {
		return 0, err
	}
	return w.file.Write(p)
}

func (w *dateRollingWriter) rotateIfNeeded() error {
	today := time.Now().Format("20060102")
	if w.file != nil && w.curDate == today {
		return nil
	}
	if w.file != nil {
		_ = w.file.Close()
	}
	filename := filepath.Join(w.dir, fmt.Sprintf("%s_%s.log", w.prefix, today))
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file = f
	w.curDate = today
	// After successful rotation, perform cleanup of old files.
	w.cleanupOldFiles()
	return nil
}

// cleanupOldFiles deletes log files older than retentionDays.
func (w *dateRollingWriter) cleanupOldFiles() {
	if w.retentionDays <= 0 {
		return
	}
	cutoff := time.Now().Add(-time.Duration(w.retentionDays) * 24 * time.Hour)
	entries, err := ioutil.ReadDir(w.dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		// expect pattern: <prefix>_YYYYMMDD.log
		if !strings.HasPrefix(name, w.prefix+"_") || !strings.HasSuffix(name, ".log") {
			continue
		}
		datePart := strings.TrimSuffix(strings.TrimPrefix(name, w.prefix+"_"), ".log")
		if len(datePart) != 8 {
			continue
		}
		t, err := time.Parse("20060102", datePart)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			_ = os.Remove(filepath.Join(w.dir, name))
		}
	}
}

// InitLogger configures the global standard logger to write into daily log files.
// Example files: logs/app_20250922.log
func InitLogger() {
	// Default location relative to the working directory of the process.
	dir := "logs"
	prefix := "app"
	w, err := newDateRollingWriter(dir, prefix)
	if err != nil {
		// fallback to stderr if file init fails
		log.Printf("InitLogger fallback, open log file fail: %v", err)
		return
	}
	log.SetOutput(w)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// Background scheduler: sleep until next midnight, then rotate & cleanup.
	go func() {
		for {
			now := time.Now()
			next := nextMidnight(now)
			time.Sleep(time.Until(next))
			_ = w.rotateIfNeeded()
		}
	}()
}

// getRetentionDaysFromEnv reads LOG_RETENTION_DAYS from env, default 7.
func getRetentionDaysFromEnv() int {
	v := os.Getenv("LOG_RETENTION_DAYS")
	if v == "" {
		return 7
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 7
	}
	return n
}

// nextMidnight returns the next local midnight time after t.
func nextMidnight(t time.Time) time.Time {
	y, m, d := t.Date()
	loc := t.Location()
	return time.Date(y, m, d+1, 0, 0, 0, 0, loc)
}
