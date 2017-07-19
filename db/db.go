package db

import (
	"time"
)

// CronJob is the normalized format interfacing with API handlers
type CronJob struct {
	ID       string
	IsActive bool
	Function string
	Workload string
	CronTime string
	TimeZone string
	Backend  string
	Created  time.Time
}

// DB is required type when registering API handlers
type DB interface {
	GetDistinctActiveFunctions() ([]string, error)
	GetJobs(function string) ([]CronJob, error)
	UpdateJob(cronJob CronJob) error
	AddJob(job CronJob) error
	DeleteJob(jobID string) error
}
