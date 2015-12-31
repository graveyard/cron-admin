package db

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	cronCollection = "cronjobs"
)

type MongoCronJob struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	IsActive bool          `bson:"active"`
	Function string        `bson:"function"`
	Workload interface{}   `bson:"workload"`
	CronTime string        `bson:"time"`
	TimeZone string        `bson:"tz"`
	Created  time.Time     `bson:"created"`
}

func (mc *MongoCronJob) ToCronJob() CronJob {
	return CronJob{
		ID:       mc.ID.Hex(),
		IsActive: mc.IsActive,
		Function: mc.Function,
		Workload: mc.Workload,
		CronTime: mc.CronTime,
		TimeZone: mc.TimeZone,
		Created:  mc.Created,
	}
}

type MongoDB struct {
	Session *mgo.Session
	DBName  string
}

func (db *MongoDB) SessionClone() *mgo.Session {
	return db.Session.Clone()
}

func (db *MongoDB) GetCronCollection(session *mgo.Session) *mgo.Collection {
	return session.DB(db.DBName).C(cronCollection)
}

func NewMongoDB(mongoURL string, dbName string) (*MongoDB, error) {
	session, dialErr := mgo.Dial(mongoURL)
	if dialErr != nil {
		return nil, dialErr
	}
	return &MongoDB{Session: session, DBName: dbName}, nil
}

func (db *MongoDB) GetDistinctActiveJobs() ([]string, error) {
	var activeJobs []string

	session := db.SessionClone()
	defer session.Close()
	collection := db.GetCronCollection(session)

	query := collection.Find(bson.M{"active": true})
	if err := query.Distinct("function", &activeJobs); err != nil {
		return activeJobs, err
	}
	return activeJobs, nil
}

func (db *MongoDB) GetJobDetails(job string) ([]CronJob, error) {
	var mongoJobDetails []MongoCronJob
	var jobDetails []CronJob

	session := db.SessionClone()
	defer session.Close()
	collection := db.GetCronCollection(session)

	query := collection.Find(bson.M{"function": job})
	if err := query.All(&mongoJobDetails); err != nil {
		return jobDetails, err
	}

	for _, mongoJob := range mongoJobDetails {
		jobDetails = append(jobDetails, mongoJob.ToCronJob())
	}

	return jobDetails, nil
}

func (db *MongoDB) UpdateJobActivationStatus(jobID string, isActive bool) error {
	session := db.SessionClone()
	defer session.Close()
	collection := db.GetCronCollection(session)

	query := bson.M{"_id": bson.ObjectIdHex(jobID)}
	change := bson.M{"$set": bson.M{"active": isActive}}
	if err := collection.Update(query, change); err != nil {
		return err
	}
	return nil
}

func (db *MongoDB) AddJob(job CronJob) error {
	session := db.SessionClone()
	defer session.Close()
	collection := db.GetCronCollection(session)

	insertJob := MongoCronJob{
		Function: job.Function,
		IsActive: job.IsActive,
		Workload: job.Workload,
		CronTime: job.CronTime,
		TimeZone: job.TimeZone,
		Created:  time.Now(),
	}

	if insertErr := collection.Insert(&insertJob); insertErr != nil {
		return insertErr
	}

	return nil
}
