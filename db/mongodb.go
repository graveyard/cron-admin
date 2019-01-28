package db

import (
	"encoding/json"
	"time"

	"gopkg.in/Clever/kayvee-go.v6/logger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	cronCollection = "cronjobs"
	kv             = logger.New("cronAdmin")
)

type mongoCronJob struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	IsActive bool          `bson:"active"`
	Function string        `bson:"function"`
	Workload interface{}   `bson:"workload"`
	CronTime string        `bson:"time"`
	TimeZone string        `bson:"tz"`
	Created  time.Time     `bson:"created"`
	Backend  string        `bson:"backend"`
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
		Backend:  c.Backend,
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
	case []interface{}:
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
		Backend:  mc.Backend,
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
	var oldMongoJob mongoCronJob
	session := db.sessionClone()
	defer session.Close()
	collection := db.getCronCollection(session)

	mongoCron := cronJob.toMongoCronJob()
	if err := collection.Find(bson.M{"_id": mongoCron.ID}).One(&oldMongoJob); err != nil {
		kv.WarnD("Error finding job during update", logger.M{"err": err.Error()})
	}
	kv.InfoD("JobUpdated", logger.M{"oldJob": oldMongoJob, "newJob": mongoCron})

	query := bson.M{"_id": mongoCron.ID}
	// "_id" can't be non-null when updating a mongo document
	mongoCron.ID = bson.ObjectId("")

	change := bson.M{"$set": mongoCron}

	return collection.Update(query, change)
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
		Backend:  job.Backend,
	}

	kv.InfoD("JobCreated", logger.M{"newJob": insertJob})
	return collection.Insert(&insertJob)
}

// DeleteJob removes jobID from DB
func (db *MongoDB) DeleteJob(jobID string) error {
	var oldMongoJob mongoCronJob
	session := db.sessionClone()
	defer session.Close()
	collection := db.getCronCollection(session)
	if err := collection.Find(bson.M{"_id": bson.ObjectIdHex(jobID)}).One(&oldMongoJob); err != nil {
		kv.WarnD("Error finding job to delete", logger.M{"err": err.Error()})
	}
	kv.InfoD("JobDeleted", logger.M{"oldJob": oldMongoJob})
	return collection.RemoveId(bson.ObjectIdHex(jobID))
}

func parseWorkload(workloadString string) interface{} {
	var jsonWorkload map[string]interface{}
	if err := json.Unmarshal([]byte(workloadString), &jsonWorkload); err == nil {
		return bson.M(jsonWorkload)
	}

	var array []interface{}
	if err := json.Unmarshal([]byte(workloadString), &array); err == nil {
		return array
	}

	return workloadString
}
