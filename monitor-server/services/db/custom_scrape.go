package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/WeBankPartners/go-common-lib/guid"
	"github.com/WeBankPartners/open-monitor/monitor-server/middleware/log"
	"github.com/WeBankPartners/open-monitor/monitor-server/models"
	"github.com/WeBankPartners/open-monitor/monitor-server/services/prom"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	customScrapeForbiddenRootKeys = map[string]bool{
		"global":              true,
		"scrape_config_files": true,
		"remote_write":        true,
		"remote_read":         true,
		"alerting":            true,
		"rule_files":          true,
		"storage":             true,
		"tracing":             true,
		"otlp":                true,
	}
	customScrapeReservedJobNames = map[string]bool{
		"prometheus":                 true,
		"transgateway":               true,
		"ping_exporter":              true,
		"db_monitor_exporter":        true,
		"metric_comparison_exporter": true,
	}
	customScrapeReservedJobPrefixes = []string{"sd_file_", "k8s-kubelet-", "k8s-cadvisor-"}
	customScrapeGlobLine            = regexp.MustCompile(`(?m)^(\s*-\s+)/app/monitor/prometheus/custom_scrape/\*\.yml\s*$`)
	customScrapeJobNameLine         = regexp.MustCompile(`\{\{|\}\}`)
)

func CustomScrapeConfigList() (result []*models.CustomScrapeConfigTable, err error) {
	result = []*models.CustomScrapeConfigTable{}
	err = x.SQL("select * from custom_scrape_config order by create_at desc").Find(&result)
	if err != nil {
		return
	}
	for _, row := range result {
		if !row.CreateAt.IsZero() {
			row.CreateTime = row.CreateAt.Format(models.DatetimeFormat)
		}
		if !row.UpdateAt.IsZero() {
			row.UpdateTime = row.UpdateAt.Format(models.DatetimeFormat)
		}
	}
	return
}

func CustomScrapeConfigCheck(param models.CustomScrapeConfigParam) (result models.CustomScrapeCheckResult, err error) {
	fileContent, jobNames, err := prepareCustomScrapeContent(param)
	if err != nil {
		return
	}
	err = checkCustomScrapeWithPromtool(param.Guid, fileContent)
	if err != nil {
		return
	}
	result.JobNames = jobNames
	return
}

func CustomScrapeConfigCreate(param models.CustomScrapeConfigParam, operator string) error {
	fileContent, jobNames, err := prepareCustomScrapeContent(param)
	if err != nil {
		return err
	}
	if err = checkCustomScrapeWithPromtool("", fileContent); err != nil {
		return err
	}
	enabled := 1
	if param.Enabled != nil {
		enabled = *param.Enabled
		if enabled != 0 {
			enabled = 1
		}
	}
	row := models.CustomScrapeConfigTable{
		Guid:        guid.CreateGuid(),
		Name:        strings.TrimSpace(param.Name),
		JobName:     strings.Join(jobNames, ","),
		YamlContent: strings.TrimSpace(param.YamlContent),
		Enabled:     enabled,
		CreateUser:  operator,
		UpdateUser:  operator,
	}
	if err = applyCustomScrapeFiles(row.Guid, fileContent, enabled == 1); err != nil {
		return err
	}
	_, err = x.Exec("insert into custom_scrape_config(guid,name,job_name,yaml_content,enabled,create_user,update_user,create_at) value (?,?,?,?,?,?,?,?)",
		row.Guid, row.Name, row.JobName, row.YamlContent, row.Enabled, row.CreateUser, row.UpdateUser, time.Now())
	if err != nil {
		_ = applyCustomScrapeFiles(row.Guid, "", false)
		return fmt.Errorf("Insert database fail,%s ", err.Error())
	}
	return nil
}

func CustomScrapeConfigUpdate(param models.CustomScrapeConfigParam, operator string) error {
	if strings.TrimSpace(param.Guid) == "" {
		return fmt.Errorf("guid is empty")
	}
	exist, err := getCustomScrapeConfig(param.Guid)
	if err != nil {
		return err
	}
	fileContent, jobNames, err := prepareCustomScrapeContent(param)
	if err != nil {
		return err
	}
	if err = checkCustomScrapeWithPromtool(param.Guid, fileContent); err != nil {
		return err
	}
	enabled := exist.Enabled
	if param.Enabled != nil {
		enabled = *param.Enabled
		if enabled != 0 {
			enabled = 1
		}
	}
	oldContent, _ := ioutil.ReadFile(customScrapeFilePath(param.Guid))
	if err = applyCustomScrapeFiles(param.Guid, fileContent, enabled == 1); err != nil {
		return err
	}
	_, err = x.Exec("update custom_scrape_config set name=?,job_name=?,yaml_content=?,enabled=?,update_user=? where guid=?",
		strings.TrimSpace(param.Name), strings.Join(jobNames, ","), strings.TrimSpace(param.YamlContent), enabled, operator, param.Guid)
	if err != nil {
		_ = rollbackCustomScrapeFile(param.Guid, oldContent, exist.Enabled == 1)
		return fmt.Errorf("Update database fail,%s ", err.Error())
	}
	return nil
}

func CustomScrapeConfigDelete(guidStr string) error {
	exist, err := getCustomScrapeConfig(guidStr)
	if err != nil {
		return err
	}
	oldContent, _ := ioutil.ReadFile(customScrapeFilePath(guidStr))
	if err = applyCustomScrapeFiles(guidStr, "", false); err != nil {
		return err
	}
	_, err = x.Exec("delete from custom_scrape_config where guid=?", guidStr)
	if err != nil {
		_ = rollbackCustomScrapeFile(guidStr, oldContent, exist.Enabled == 1)
		return fmt.Errorf("Delete database fail,%s ", err.Error())
	}
	return nil
}

func SyncCustomScrapeConfig() error {
	rows, err := CustomScrapeConfigList()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(models.CustomScrapeDir, 0755); err != nil {
		return fmt.Errorf("Create custom scrape dir fail,%s ", err.Error())
	}
	keep := make(map[string]bool)
	for _, row := range rows {
		if row.Enabled != 1 {
			continue
		}
		fileContent, _, parseErr := normalizeCustomScrapeYaml(row.YamlContent)
		if parseErr != nil {
			log.Error(nil, log.LOGGER_APP, "Skip invalid custom scrape config", zap.String("guid", row.Guid), zap.Error(parseErr))
			continue
		}
		if writeErr := ioutil.WriteFile(customScrapeFilePath(row.Guid), []byte(fileContent), 0644); writeErr != nil {
			return fmt.Errorf("Write custom scrape file fail,%s ", writeErr.Error())
		}
		keep[row.Guid+".yml"] = true
	}
	entries, _ := ioutil.ReadDir(models.CustomScrapeDir)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}
		if !keep[entry.Name()] {
			_ = os.Remove(filepath.Join(models.CustomScrapeDir, entry.Name()))
		}
	}
	if err = ensureCustomScrapeConfigFiles(); err != nil {
		return err
	}
	if err = prom.ReloadConfigNow(); err != nil {
		return err
	}
	return nil
}

func getCustomScrapeConfig(guidStr string) (*models.CustomScrapeConfigTable, error) {
	var rows []*models.CustomScrapeConfigTable
	err := x.SQL("select * from custom_scrape_config where guid=?", guidStr).Find(&rows)
	if err != nil {
		return nil, fmt.Errorf("Query custom scrape config fail,%s ", err.Error())
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("custom scrape config %s not found", guidStr)
	}
	return rows[0], nil
}

func prepareCustomScrapeContent(param models.CustomScrapeConfigParam) (fileContent string, jobNames []string, err error) {
	name := strings.TrimSpace(param.Name)
	if name == "" {
		err = fmt.Errorf("name is required")
		return
	}
	if len(name) > 64 {
		err = fmt.Errorf("name is too long")
		return
	}
	yamlContent := strings.TrimSpace(param.YamlContent)
	if yamlContent == "" {
		err = fmt.Errorf("yaml_content is required")
		return
	}
	if len(yamlContent) > models.CustomScrapeYamlMaxBytes {
		err = fmt.Errorf("yaml_content larger than %d bytes", models.CustomScrapeYamlMaxBytes)
		return
	}
	if customScrapeJobNameLine.MatchString(yamlContent) {
		err = fmt.Errorf("yaml_content can not contain template placeholder {{ }}")
		return
	}
	fileContent, jobNames, err = normalizeCustomScrapeYaml(yamlContent)
	if err != nil {
		return
	}
	if err = checkCustomScrapeJobConflict(param.Guid, name, jobNames); err != nil {
		return
	}
	return
}

func normalizeCustomScrapeYaml(content string) (fileContent string, jobNames []string, err error) {
	var raw interface{}
	if err = yaml.Unmarshal([]byte(content), &raw); err != nil {
		err = fmt.Errorf("yaml parse fail,%s", err.Error())
		return
	}
	if raw == nil {
		err = fmt.Errorf("yaml_content is empty")
		return
	}
	var jobs []map[interface{}]interface{}
	switch typed := raw.(type) {
	case []interface{}:
		jobs, err = jobsFromYamlList(typed)
		if err != nil {
			return
		}
	case map[interface{}]interface{}:
		if scrapeRaw, hasScrape := lookupYamlMap(typed, "scrape_configs"); hasScrape {
			for key := range typed {
				keyStr := fmt.Sprintf("%v", key)
				if keyStr == "scrape_configs" {
					continue
				}
				if customScrapeForbiddenRootKeys[keyStr] {
					err = fmt.Errorf("please only paste scrape job list starting with - job_name, do not paste full prometheus.yml (found %s)", keyStr)
					return
				}
				err = fmt.Errorf("scrape file extra root key %s is not allowed", keyStr)
				return
			}
			scrapeList, ok := scrapeRaw.([]interface{})
			if !ok {
				err = fmt.Errorf("scrape_configs must be a job list")
				return
			}
			jobs, err = jobsFromYamlList(scrapeList)
			if err != nil {
				return
			}
		} else {
			for key := range typed {
				keyStr := fmt.Sprintf("%v", key)
				if customScrapeForbiddenRootKeys[keyStr] {
					err = fmt.Errorf("please only paste scrape job list starting with - job_name, do not paste full prometheus.yml (found %s)", keyStr)
					return
				}
			}
			jobs = []map[interface{}]interface{}{typed}
		}
	default:
		err = fmt.Errorf("yaml_content must be a scrape job list starting with - job_name")
		return
	}
	nameSet := make(map[string]bool)
	for i, job := range jobs {
		jobName := strings.TrimSpace(getYamlMapString(job, "job_name"))
		if jobName == "" {
			err = fmt.Errorf("scrape job[%d] job_name is required", i)
			return
		}
		if err = validateCustomScrapeJobName(jobName); err != nil {
			return
		}
		if nameSet[jobName] {
			err = fmt.Errorf("duplicate job_name %s in yaml", jobName)
			return
		}
		nameSet[jobName] = true
		jobNames = append(jobNames, jobName)
	}
	fileContent, err = marshalCustomScrapeFile(jobs)
	return
}

func jobsFromYamlList(typed []interface{}) ([]map[interface{}]interface{}, error) {
	if len(typed) == 0 {
		return nil, fmt.Errorf("yaml_content must contain at least one scrape job")
	}
	jobs := make([]map[interface{}]interface{}, 0, len(typed))
	for i, item := range typed {
		jobMap, ok := item.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("scrape job[%d] must be a mapping", i)
		}
		jobs = append(jobs, jobMap)
	}
	return jobs, nil
}

func marshalCustomScrapeFile(jobs []map[interface{}]interface{}) (string, error) {
	out, err := yaml.Marshal(map[string]interface{}{"scrape_configs": jobs})
	if err != nil {
		return "", fmt.Errorf("wrap scrape job fail,%s", err.Error())
	}
	return string(out), nil
}

func lookupYamlMap(input map[interface{}]interface{}, key string) (interface{}, bool) {
	if input == nil {
		return nil, false
	}
	if v, ok := input[key]; ok {
		return v, true
	}
	return nil, false
}

func validateCustomScrapeJobName(jobName string) error {
	if customScrapeReservedJobNames[jobName] {
		return fmt.Errorf("job_name %s is reserved", jobName)
	}
	for _, prefix := range customScrapeReservedJobPrefixes {
		if strings.HasPrefix(jobName, prefix) {
			return fmt.Errorf("job_name %s is reserved", jobName)
		}
	}
	return nil
}

func checkCustomScrapeJobConflict(currentGuid, name string, jobNames []string) error {
	rows, err := CustomScrapeConfigList()
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Guid == currentGuid {
			continue
		}
		if strings.EqualFold(row.Name, name) {
			return fmt.Errorf("name %s already exists", name)
		}
		existJobs := strings.Split(row.JobName, ",")
		existMap := make(map[string]bool)
		for _, existJob := range existJobs {
			existMap[strings.TrimSpace(existJob)] = true
		}
		for _, jobName := range jobNames {
			if existMap[jobName] {
				return fmt.Errorf("job_name %s already exists in %s", jobName, row.Name)
			}
		}
	}
	snmpList, _ := SnmpExporterList()
	for _, snmp := range snmpList {
		for _, jobName := range jobNames {
			if snmp.Id == jobName {
				return fmt.Errorf("job_name %s already used by snmp exporter", jobName)
			}
		}
	}
	return nil
}

func checkCustomScrapeWithPromtool(currentGuid, fileContent string) error {
	if _, err := os.Stat(models.PromtoolPath); err != nil {
		log.Warn(nil, log.LOGGER_APP, "promtool not found, skip official prometheus yaml check", zap.Error(err))
		return nil
	}
	promBytes, err := ioutil.ReadFile(models.PrometheusYmlPath)
	if err != nil {
		return fmt.Errorf("Read prometheus.yml fail,%s ", err.Error())
	}
	tmpDir, err := ioutil.TempDir("/tmp", "prom-scrape-check-")
	if err != nil {
		return fmt.Errorf("Create temp dir fail,%s ", err.Error())
	}
	defer os.RemoveAll(tmpDir)
	customDir := filepath.Join(tmpDir, "custom_scrape")
	if err = os.MkdirAll(customDir, 0755); err != nil {
		return err
	}
	tmpGlob := filepath.Join(customDir, "*.yml")
	promStr := replaceCustomScrapeGlob(string(promBytes), tmpGlob)
	if err = ioutil.WriteFile(filepath.Join(tmpDir, "prometheus.yml"), []byte(promStr), 0644); err != nil {
		return err
	}
	rows, _ := CustomScrapeConfigList()
	for _, row := range rows {
		if row.Enabled != 1 || row.Guid == currentGuid {
			continue
		}
		normalized, _, parseErr := normalizeCustomScrapeYaml(row.YamlContent)
		if parseErr != nil {
			continue
		}
		if writeErr := ioutil.WriteFile(filepath.Join(customDir, row.Guid+".yml"), []byte(normalized), 0644); writeErr != nil {
			return writeErr
		}
	}
	checkName := currentGuid
	if checkName == "" {
		checkName = "new"
	}
	if err = ioutil.WriteFile(filepath.Join(customDir, checkName+".yml"), []byte(fileContent), 0644); err != nil {
		return err
	}
	cmd := exec.Command(models.PromtoolPath, "check", "config", filepath.Join(tmpDir, "prometheus.yml"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("prometheus yaml check fail,%s", msg)
	}
	return nil
}

func applyCustomScrapeFiles(guidStr, fileContent string, enabled bool) error {
	if err := os.MkdirAll(models.CustomScrapeDir, 0755); err != nil {
		return fmt.Errorf("Create custom scrape dir fail,%s ", err.Error())
	}
	if err := ensureCustomScrapeConfigFiles(); err != nil {
		return err
	}
	target := customScrapeFilePath(guidStr)
	oldContent, readErr := ioutil.ReadFile(target)
	oldExisted := readErr == nil
	if enabled {
		if err := ioutil.WriteFile(target, []byte(fileContent), 0644); err != nil {
			return fmt.Errorf("Write custom scrape file fail,%s ", err.Error())
		}
	} else if oldExisted {
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("Remove custom scrape file fail,%s ", err.Error())
		}
	}
	if err := prom.ReloadConfigNow(); err != nil {
		_ = rollbackCustomScrapeFile(guidStr, oldContent, oldExisted)
		return err
	}
	return nil
}

func rollbackCustomScrapeFile(guidStr string, oldContent []byte, existed bool) error {
	target := customScrapeFilePath(guidStr)
	if existed && len(oldContent) > 0 {
		_ = ioutil.WriteFile(target, oldContent, 0644)
	} else {
		_ = os.Remove(target)
	}
	return prom.ReloadConfigNow()
}

func customScrapeFilePath(guidStr string) string {
	return filepath.Join(models.CustomScrapeDir, guidStr+".yml")
}

func ensureCustomScrapeConfigFiles() error {
	prometheusConfigMutex.Lock()
	defer prometheusConfigMutex.Unlock()
	promBytes, err := ioutil.ReadFile(models.PrometheusYmlPath)
	if err != nil {
		return fmt.Errorf("Read prometheus.yml fail,%s ", err.Error())
	}
	promStr := string(promBytes)
	if strings.Contains(promStr, models.CustomScrapeConfigMarker) {
		return nil
	}
	backupName, backupErr := backupPrometheusConfig()
	if backupErr != nil {
		return backupErr
	}
	promStr = insertCustomScrapeConfigFiles(promStr, models.CustomScrapeGlob)
	if err = ioutil.WriteFile(models.PrometheusYmlPath, []byte(promStr), 0644); err != nil {
		recoverPrometheusConfig(backupName)
		return fmt.Errorf("Write prometheus.yml fail,%s ", err.Error())
	}
	return nil
}

func replaceCustomScrapeGlob(promStr, newGlob string) string {
	if customScrapeGlobLine.MatchString(promStr) {
		return customScrapeGlobLine.ReplaceAllString(promStr, "${1}"+newGlob)
	}
	return insertCustomScrapeConfigFiles(promStr, newGlob)
}

func insertCustomScrapeConfigFiles(promStr, glob string) string {
	block := fmt.Sprintf("\n# Custom scrape jobs live in isolated files. Comment scrape_config_files to disable them all.\nscrape_config_files:\n  - %s\n\n", glob)
	idx := strings.Index(promStr, "\nremote_write:")
	if idx < 0 {
		idx = strings.Index(promStr, "\n#Remote write configuration")
	}
	if idx >= 0 {
		return promStr[:idx] + block + promStr[idx+1:]
	}
	return promStr + block
}

func getYamlMapString(input map[interface{}]interface{}, key string) string {
	if input == nil {
		return ""
	}
	if v, ok := input[key]; ok {
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
	return ""
}
