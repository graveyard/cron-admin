package db

import (
	"time"
)

type CronJob struct {
	ID       string
	IsActive bool
	Function string
	Workload interface{}
	CronTime string
	TimeZone string
	Created  time.Time
}

type DB interface {
	GetDistinctActiveJobs() ([]string, error)
	GetJobDetails(job string) ([]CronJob, error)
	UpdateJobActivationStatus(jobID string, isActive bool) error
	AddJob(job CronJob) error
}
