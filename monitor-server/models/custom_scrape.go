package models

import "time"

const (
	PrometheusYmlPath        = "/app/monitor/prometheus/prometheus.yml"
	PromtoolPath             = "/app/monitor/prometheus/promtool"
	CustomScrapeDir          = "/app/monitor/prometheus/custom_scrape"
	CustomScrapeGlob         = "/app/monitor/prometheus/custom_scrape/*.yml"
	CustomScrapeYamlMaxBytes = 64 * 1024
	CustomScrapeConfigMarker = "/app/monitor/prometheus/custom_scrape/*.yml"
)

type CustomScrapeConfigTable struct {
	Guid        string    `json:"guid" xorm:"guid"`
	Name        string    `json:"name" xorm:"name"`
	JobName     string    `json:"job_name" xorm:"job_name"`
	YamlContent string    `json:"yaml_content" xorm:"yaml_content"`
	Enabled     int       `json:"enabled" xorm:"enabled"`
	CreateUser  string    `json:"create_user" xorm:"create_user"`
	UpdateUser  string    `json:"update_user" xorm:"update_user"`
	CreateAt    time.Time `json:"create_at" xorm:"create_at"`
	UpdateAt    time.Time `json:"update_at" xorm:"update_at"`
	CreateTime  string    `json:"create_time" xorm:"-"`
	UpdateTime  string    `json:"update_time" xorm:"-"`
}

type CustomScrapeConfigParam struct {
	Guid        string `json:"guid"`
	Name        string `json:"name"`
	YamlContent string `json:"yaml_content"`
	Enabled     *int   `json:"enabled"`
}

type CustomScrapeCheckResult struct {
	JobNames []string `json:"job_names"`
}
