package db

import (
	"encoding/json"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var (
	cronCollection = "cronjobs"
)

type mongoCronJob struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	IsActive bool          `bson:"active"`
	Function string        `bson:"function"`
	Workload interface{}   `bson:"workload"`
	CronTime string        `bson:"time"`
	TimeZone string        `bson:"tz"`
	Created  time.Time     `bson:"created"`
}

func (c CronJob) toMongoCronJob() mongoCronJob {
	var id bson.ObjectId
	if c.ID != "" {
		id = bson.ObjectIdHex(c.ID)
	}

	return mongoCronJob{
		ID:       id,
		IsActive: c.IsActive,
		Function: c.Function,
		Workload: parseWorkload(c.Workload),
		CronTime: c.CronTime,
		TimeZone: c.TimeZone,
		Created:  c.Created,
	}
}

func (mc *mongoCronJob) toCronJob() CronJob {
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

// MongoDB is the visible DB interface used in server
type MongoDB struct {
	session *mgo.Session
	dbName  string
}

func (db *MongoDB) sessionClone() *mgo.Session {
	return db.session.Clone()
}

func (db *MongoDB) getCronCollection(session *mgo.Session) *mgo.Collection {
	if session == nil {
		session = db.session
	}
	return session.DB(db.dbName).C(cronCollection)
}

// NewMongoDB attempts to connect with DB and errs if problems are found
func NewMongoDB(mongoURL string, dbName string) (*MongoDB, error) {
	session, dialErr := mgo.Dial(mongoURL)
	if dialErr != nil {
		return nil, dialErr
	}
	return &MongoDB{session: session, dbName: dbName}, nil
}

// GetDistinctActiveFunctions returns a list of functions with active jobs
func (db *MongoDB) GetDistinctActiveFunctions() ([]string, error) {
	var activeFunctions []string

	session := db.sessionClone()
	defer session.Close()
	collection := db.getCronCollection(session)

	query := collection.Find(bson.M{"active": true})
	if err := query.Distinct("function", &activeFunctions); err != nil {
		return activeFunctions, err
	}
	return activeFunctions, nil
}

// GetJobs returns all jobs associated with a function
func (db *MongoDB) GetJobs(function string) ([]CronJob, error) {
	var mongoJobDetails []mongoCronJob
	var jobDetails []CronJob

	session := db.sessionClone()
	defer session.Close()
	collection := db.getCronCollection(session)

	query := collection.Find(bson.M{"function": function})
	if err := query.All(&mongoJobDetails); err != nil {
		return jobDetails, err
	}

	for _, mongoJob := range mongoJobDetails {
		jobDetails = append(jobDetails, mongoJob.toCronJob())
	}

	return jobDetails, nil
}

// UpdateJob updates existing document as determined by CronJob input
func (db *MongoDB) UpdateJob(cronJob CronJob) error {
	session := db.sessionClone()
	defer session.Close()
	collection := db.getCronCollection(session)

	mongoCron := cronJob.toMongoCronJob()
	query := bson.M{"_id": mongoCron.ID}
	// "_id" can't be non-null when updating a mongo document
	mongoCron.ID = bson.ObjectId("")

	change := bson.M{"$set": mongoCron}
	if err := collection.Update(query, change); err != nil {
		return err
	}
	return nil
}

// AddJob parses CronJob input and inserts into DB
func (db *MongoDB) AddJob(job CronJob) error {
	session := db.sessionClone()
	defer session.Close()
	collection := db.getCronCollection(session)

	insertJob := mongoCronJob{
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

// DeleteJob removes jobID from DB
func (db *MongoDB) DeleteJob(jobID string) error {
	session := db.sessionClone()
	defer session.Close()
	collection := db.getCronCollection(session)
	return collection.RemoveId(bson.ObjectIdHex(jobID))
}

func parseWorkload(workloadString string) interface{} {
	var jsonWorkload map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(workloadString), &jsonWorkload); jsonErr == nil {
		return bson.M(jsonWorkload)
	}
	return workloadString
}
