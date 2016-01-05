package db

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var (
	cronCollection     = "cronjobs"
	mongoCollectionMap = map[string]string{
		"id":       "_id",
		"IsActive": "active",
		"Function": "function",
		"Workload": "workload",
		"CronTime": "time",
		"TimeZone": "tz",
	}
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

func (db *MongoDB) UpdateJob(jobID string, fieldMap map[string]interface{}) error {
	session := db.SessionClone()
	defer session.Close()
	collection := db.GetCronCollection(session)

	updateMap := make(map[string]interface{})
	for k, v := range fieldMap {
		mongoKey, ok := mongoCollectionMap[k]
		if !ok {
			return ErrUnknownField{Field: k}
		}

		if k == "Workload" {
			switch val := v.(type) {
			case string:
				updateMap[mongoKey] = parseWorkload(val)
			default:
				return ErrUnsupportedWorkloadType{Type: fmt.Sprintf("%T", val)}
			}
		} else {
			updateMap[mongoKey] = v
		}
	}

	query := bson.M{"_id": bson.ObjectIdHex(jobID)}
	change := bson.M{"$set": bson.M(updateMap)}
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
