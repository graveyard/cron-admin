package mongodb

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	cronCollection = "cronjobs"
)

type DB struct {
	DB      *mgo.Database
	Session *mgo.Session
}

type CronJob struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	IsActive bool          `bson:"active"`
	Function string        `bson:"function"`
	Workload interface{}   `bson:"workload"`
	CronTime string        `bson:"time"`
	TimeZone string        `bson:"tz"`
	Created  time.Time     `bson:"created"`
}

func New(mongoURL string, dbName string) (*DB, error) {
	session, dialErr := mgo.Dial(mongoURL)
	if dialErr != nil {
		return nil, dialErr
	}
	cleverDb := session.DB(dbName)
	return &DB{DB: cleverDb, Session: session}, nil
}

func (db *DB) GetDistinctActiveJobs() ([]string, error) {
	var activeJobs []string
	collection := db.DB.C(cronCollection)
	query := collection.Find(bson.M{"active": true})
	if err := query.Distinct("function", &activeJobs); err != nil {
		return activeJobs, err
	}
	return activeJobs, nil
}

func (db *DB) GetJobDetails(job string) ([]CronJob, error) {
	var jobDetails []CronJob
	collection := db.DB.C(cronCollection)
	query := collection.Find(bson.M{"function": job})
	if err := query.All(&jobDetails); err != nil {
		return jobDetails, err
	}
	return jobDetails, nil
}

func (db *DB) UpdateJobActivity(jobID string, isActive bool) error {
	collection := db.DB.C(cronCollection)
	query := bson.M{"_id": bson.ObjectIdHex(jobID)}
	change := bson.M{"$set": bson.M{"active": isActive}}
	if err := collection.Update(query, change); err != nil {
		return err
	}
	return nil
}

func (db *DB) AddJob(function, crontime string, workload interface{}) error {
	collection := db.DB.C(cronCollection)
	var workloadResult interface{}
	// This will help determine how to submit the workload
	switch w := workload.(type) {
	default:
		workloadResult = w
	case map[string]interface{}:
		workloadResult = bson.M(w)
	}

	insertJob := CronJob{
		Function: function,
		IsActive: true,
		Workload: workloadResult,
		CronTime: crontime,
		TimeZone: "America/Los_Angeles",
		Created:  time.Now(),
	}

	if insertErr := collection.Insert(&insertJob); insertErr != nil {
		return insertErr
	}

	return nil
}
