package db

import (
	"log"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func mongoTestWrapper(assert *assert.Assertions, dbName string, test MethodTest) {
	testURL := mongoTestURL(dbName)
	database, dialErr := NewMongoDB(testURL, dbName)
	assert.NoError(dialErr)
	defer database.session.Close()
	defer dropDatabase(testURL)
	test(assert, database)
}

// TestMongoInterface tests all MongoDB interface implementations
func TestMongoInterface(t *testing.T) {
	DBTest(assert.New(t), mongoTestWrapper)
}

// TestParseWorkload verifies that valid json inputs are converted to bson.M objects
func TestParseWorkload(t *testing.T) {
	for _, test := range []struct {
		Input  string
		Output interface{}
	}{
		{Input: "--task", Output: interface{}("--task")},
		{Input: "{\"foo\":\"bar\"}", Output: bson.M(map[string]interface{}{"foo": "bar"})},
		{Input: `["array", "of", "items"]`, Output: []interface{}{"array", "of", "items"}},
	} {
		assert.Equal(t, test.Output, parseWorkload(test.Input))
	}
}

func TestToCronJob(t *testing.T) {
	idHex := "12345678901234567890abcd"
	id := bson.ObjectIdHex(idHex)
	v := mongoCronJob{
		ID:       id,
		IsActive: true,
		Function: "echo",
		Workload: bson.M{"district_id": "abc123"},
		CronTime: "* * 3 * * *",
		TimeZone: "America/Los_Angeles",
		Created:  time.Date(2017, 5, 20, 0, 0, 0, 0, time.UTC),
		Backend:  "gearman",
	}
	actual := v.toCronJob()
	expected := CronJob{
		ID:       idHex,
		IsActive: true,
		Function: "echo",
		Workload: `{"district_id":"abc123"}`,
		CronTime: "* * 3 * * *",
		TimeZone: "America/Los_Angeles",
		Created:  time.Date(2017, 5, 20, 0, 0, 0, 0, time.UTC),
		Backend:  "gearman",
	}
	assert.Equal(t, expected, actual)
}

// Auxilliary wrapper functions

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
