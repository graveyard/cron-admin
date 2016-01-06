package db

import (
	"log"
	"net/url"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	trialCreateTime = time.Date(2014, time.November, 12, 11, 45, 26, 371000000, time.UTC)
	trialTimeZone   = "America/Los_Angeles"
)

var cron1 = CronJob{
	ID:       "",
	IsActive: true,
	Function: "test1",
	Workload: "",
	CronTime: "* * * * * *",
	TimeZone: trialTimeZone,
	Created:  trialCreateTime,
}

var cron2 = CronJob{
	ID:       "",
	IsActive: true,
	Function: "test1",
	Workload: "",
	CronTime: "1 2 3 4 5 6",
	TimeZone: trialTimeZone,
	Created:  trialCreateTime,
}

var cron3 = CronJob{
	ID:       "",
	IsActive: false,
	Function: "test2",
	Workload: "",
	CronTime: "6 5 4 3 2 1",
	TimeZone: trialTimeZone,
	Created:  trialCreateTime,
}

var cron4 = CronJob{
	ID:       "",
	IsActive: true,
	Function: "test3",
	Workload: "",
	CronTime: "6 5 4 3 2 1",
	TimeZone: trialTimeZone,
	Created:  trialCreateTime,
}

var cron5 = CronJob{
	ID:       bson.NewObjectId().Hex(),
	IsActive: true,
	Function: "test4",
	Workload: "",
	CronTime: "6 5 4 3 2 1",
	TimeZone: trialTimeZone,
	Created:  trialCreateTime,
}

var cron6 = CronJob{
	ID:       bson.NewObjectId().Hex(),
	IsActive: true,
	Function: "test4",
	Workload: "{\"foo\":\"bar\"}",
	CronTime: "6 5 4 3 2 1",
	TimeZone: trialTimeZone,
	Created:  trialCreateTime,
}

func TestParseWorkload(t *testing.T) {
	w1 := "--task"
	assert.Equal(t, w1, parseWorkload(w1))

	w2 := "{\"foo\": \"bar\"}"
	expected := bson.M(map[string]interface{}{"foo": "bar"})
	assert.Equal(t, expected, parseWorkload(w2))
}

func TestToCronJob(t *testing.T) {
	mc1 := mongoCronJob{
		ID:       bson.ObjectIdHex(cron5.ID),
		IsActive: cron5.IsActive,
		Function: cron5.Function,
		Workload: "",
		CronTime: cron5.CronTime,
		TimeZone: cron5.TimeZone,
		Created:  cron5.Created,
	}

	assert.Equal(t, mc1.toCronJob(), cron5)

	mc2 := mongoCronJob{
		ID:       bson.ObjectIdHex(cron6.ID),
		IsActive: cron6.IsActive,
		Function: cron6.Function,
		Workload: bson.M(map[string]interface{}{"foo": "bar"}),
		CronTime: cron6.CronTime,
		TimeZone: cron6.TimeZone,
		Created:  cron6.Created,
	}
	assert.Equal(t, mc2.toCronJob(), cron6)

}

func TestToMongoCron(t *testing.T) {
	assert.Equal(t, cron1.ID, "")
	mc1 := cron1.toMongoCronJob()
	var id bson.ObjectId
	expected := mongoCronJob{
		ID:       id,
		IsActive: true,
		Function: "test1",
		Workload: "",
		CronTime: "* * * * * *",
		TimeZone: trialTimeZone,
		Created:  trialCreateTime,
	}
	assert.Equal(t, mc1, expected)

	assert.NotEqual(t, cron5.ID, "")
	mc2 := cron5.toMongoCronJob()
	expected = mongoCronJob{
		ID:       bson.ObjectId(mc2.ID),
		IsActive: true,
		Function: "test4",
		Workload: "",
		CronTime: "6 5 4 3 2 1",
		TimeZone: trialTimeZone,
		Created:  trialCreateTime,
	}
	assert.Equal(t, mc2, expected)
}

func TestAddJob(t *testing.T) {
	dbName := "testAddJob"
	testURL := mongoTestURL(dbName)
	database, dialErr := NewMongoDB(testURL, dbName)
	assert.NoError(t, dialErr)
	defer database.session.Close()
	defer dropDatabase(testURL)

	addErr := database.AddJob(cron1)
	assert.NoError(t, addErr)

	var mc mongoCronJob
	cronCollection := database.getCronCollection(nil)
	queryErr := cronCollection.Find(bson.M{"function": cron1.Function}).One(&mc)
	assert.NoError(t, queryErr)

	cron1.ID = mc.ID.Hex()
	cron1.Created = mc.Created
	assert.Equal(t, cron1, mc.toCronJob())
	cron1.ID = ""
}

func TestDeleteJob(t *testing.T) {
	dbName := "testDeleteJob"
	testURL := mongoTestURL(dbName)
	database, dialErr := NewMongoDB(testURL, dbName)
	assert.NoError(t, dialErr)
	defer database.session.Close()
	defer dropDatabase(testURL)

	addErr := database.AddJob(cron1)
	assert.NoError(t, addErr)

	var mc mongoCronJob
	cronCollection := database.getCronCollection(nil)
	queryErr := cronCollection.Find(bson.M{"function": cron1.Function}).One(&mc)
	assert.NoError(t, queryErr)

	deleteErr := database.DeleteJob(mc.ID.Hex())
	assert.NoError(t, deleteErr)

	count, countErr := cronCollection.Count()
	assert.NoError(t, countErr)
	assert.Equal(t, count, 0)
}

func TestUpdateJob(t *testing.T) {
	dbName := "testUpdateJob"
	testURL := mongoTestURL(dbName)
	database, dialErr := NewMongoDB(testURL, dbName)
	assert.NoError(t, dialErr)
	defer database.session.Close()
	defer dropDatabase(testURL)

	addErr := database.AddJob(cron1)
	assert.NoError(t, addErr)

	var mcBefore mongoCronJob
	cronCollection := database.getCronCollection(nil)
	queryErr := cronCollection.Find(bson.M{"function": cron1.Function}).One(&mcBefore)
	assert.NoError(t, queryErr)

	modJob := cron2
	modJob.ID = mcBefore.ID.Hex()

	updateErr := database.UpdateJob(modJob)
	assert.NoError(t, updateErr)

	var mcAfter mongoCronJob
	queryErr = cronCollection.FindId(bson.ObjectIdHex(modJob.ID)).One(&mcAfter)
	assert.NoError(t, queryErr)
	modJob.Created = mcAfter.Created
	assert.Equal(t, mcAfter.toCronJob(), modJob)
}

func TestGetJobs(t *testing.T) {
	dbName := "testGetJobs"
	testURL := mongoTestURL(dbName)
	database, dialErr := NewMongoDB(testURL, dbName)
	assert.NoError(t, dialErr)
	defer database.session.Close()
	defer dropDatabase(testURL)

	assert.Equal(t, cron1.Function, cron2.Function)

	// Setup assumptions in our data
	addErr := database.AddJob(cron1)
	assert.NoError(t, addErr)
	addErr = database.AddJob(cron2)
	assert.NoError(t, addErr)

	jobs, jobsErr := database.GetJobs(cron1.Function)
	assert.NoError(t, jobsErr)
	assert.Equal(t, len(jobs), 2)

	// Even if there are no jobs for a function there should be no error
	missingFunction := "missing"
	assert.NotEqual(t, missingFunction, cron1.Function)

	_, jobsErr = database.GetJobs(missingFunction)
	assert.NoError(t, jobsErr)
}

func TestGetActiveFunctions(t *testing.T) {
	dbName := "testGetActiveFunctions"
	testURL := mongoTestURL(dbName)
	database, dialErr := NewMongoDB(testURL, dbName)
	assert.NoError(t, dialErr)
	defer database.session.Close()
	defer dropDatabase(testURL)

	// Setup assumptions in our data
	assert.Equal(t, cron1.IsActive, true)
	assert.Equal(t, cron2.IsActive, true)
	assert.Equal(t, cron3.IsActive, false)
	assert.Equal(t, cron4.IsActive, true)
	assert.Equal(t, cron1.Function, cron2.Function)
	assert.NotEqual(t, cron1.Function, cron4.Function)

	addErr := database.AddJob(cron1)
	assert.NoError(t, addErr)
	addErr = database.AddJob(cron2)
	assert.NoError(t, addErr)
	addErr = database.AddJob(cron3)
	assert.NoError(t, addErr)
	addErr = database.AddJob(cron4)
	assert.NoError(t, addErr)

	expectedFunctions := []string{cron1.Function, cron4.Function}
	sort.Strings(expectedFunctions)

	activeFunctions, getErr := database.GetDistinctActiveFunctions()
	assert.NoError(t, getErr)
	sort.Strings(activeFunctions)

	assert.Equal(t, expectedFunctions, activeFunctions)
}

func mongoTestURL(dbName string) string {
	testURL, err := url.Parse(os.Getenv("MONGO_TEST_DB"))
	if err != nil {
		log.Fatal(err.Error())
	}
	if testURL.Host == "" {
		testURL.Host = "localhost"
	}
	testURL.Scheme = "mongodb"
	testURL.Path = dbName
	return testURL.String()
}

func dropDatabase(mongoURL string) error {
	session, err := mgo.Dial(mongoURL)
	if err != nil {
		return err
	}
	return session.DB("").DropDatabase()
}
