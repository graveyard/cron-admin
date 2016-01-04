package db

import (
	"time"
)

type CronJob struct {
	ID       string
	IsActive bool
	Function string
	Workload string
	CronTime string
	TimeZone string
	Created  time.Time
}

type DB interface {
	GetDistinctActiveJobs() ([]string, error)
	GetJobDetails(job string) ([]CronJob, error)
	UpdateJob(jobID string, fieldMap map[string]interface{}) error
	AddJob(job CronJob) error
}
