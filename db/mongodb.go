package db

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var (
	cronCollection = "cronjobs"
)

type ErrUnknownField struct {
	Field string
}

type ErrUnsupportedWorkloadType struct {
	Type string
}

func (e ErrUnknownField) Error() string {
	return fmt.Sprintf("Unknown cron field: %s", e.Field)
}

func (e ErrUnsupportedWorkloadType) Error() string {
	return fmt.Sprintf("Unsupported workload type: %s", e.Type)
}

type MongoCronJob struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	IsActive bool          `bson:"active"`
	Function string        `bson:"function"`
	Workload interface{}   `bson:"workload"`
	CronTime string        `bson:"time"`
	TimeZone string        `bson:"tz"`
	Created  time.Time     `bson:"created"`
}

func (c CronJob) ToMongoCronJob() MongoCronJob {
	return MongoCronJob{
		ID:       bson.ObjectIdHex(c.ID),
		IsActive: c.IsActive,
		Function: c.Function,
		Workload: parseWorkload(c.Workload),
		CronTime: c.CronTime,
		TimeZone: c.TimeZone,
		Created:  c.Created,
	}
}

func (mc *MongoCronJob) ToCronJob() CronJob {
	var workload string
	switch t := mc.Workload.(type) {
	case string:
		workload = t
	case bson.M:
		workloadByte, _ := json.Marshal(t)
		workload = string(workloadByte)
	}

	return CronJob{
		ID:       mc.ID.Hex(),
		IsActive: mc.IsActive,
		Function: mc.Function,
		Workload: workload,
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

func (db *MongoDB) GetDistinctActiveFunctions() ([]string, error) {
	var activeFunctions []string

	session := db.SessionClone()
	defer session.Close()
	collection := db.GetCronCollection(session)

	query := collection.Find(bson.M{"active": true})
	if err := query.Distinct("function", &activeFunctions); err != nil {
		return activeFunctions, err
	}
	return activeFunctions, nil
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

func (db *MongoDB) UpdateJob(cronJob CronJob) error {
	session := db.SessionClone()
	defer session.Close()
	collection := db.GetCronCollection(session)

	mongoCron := cronJob.ToMongoCronJob()
	query := bson.M{"_id": mongoCron.ID}
	// "_id" can't be non-null when updating a mongo document
	mongoCron.ID = bson.ObjectId("")

	change := bson.M{"$set": mongoCron}
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
		Workload: parseWorkload(job.Workload),
		CronTime: job.CronTime,
		TimeZone: job.TimeZone,
		Created:  time.Now(),
	}

	if insertErr := collection.Insert(&insertJob); insertErr != nil {
		return insertErr
	}

	return nil
}

func (db *MongoDB) DeleteJob(jobID string) error {
	session := db.SessionClone()
	defer session.Close()
	collection := db.GetCronCollection(session)
	return collection.RemoveId(bson.ObjectIdHex(jobID))
}

func parseWorkload(workloadString string) interface{} {
	var jsonWorkload map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(workloadString), &jsonWorkload); jsonErr == nil {
		return bson.M(jsonWorkload)
	} else {
		return workloadString
	}
}
