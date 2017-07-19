/*
	In this file we provide a general methodology for testing interface methods specific to DB.

	The methodology has you define a DBWrapper function, handling database specific
	implementations such as connecting to and dropping dbs, and all interface methods
	are tested for free.

	While non-interface methods specific to the DB implementation may require extra tests,
	many should be implicitly tested by the success or failure of the provided interface tests.
*/

package db

import (
	"sort"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	defaultTimeZone = "America/Los_Angeles"
)

// MethodTest is a type asserting properties about the DB
type MethodTest func(*assert.Assertions, DB)

// DBWrapper will make database specific implementations while calling MethodTests. String input is used for labelling.
type DBWrapper func(*assert.Assertions, string, MethodTest)

// DBTest verifies assertions across all interface methods
func DBTest(assert *assert.Assertions, dbWrapper DBWrapper) {
	dbWrapper(assert, "testAddJob", testAddJob)
	dbWrapper(assert, "testUpdateJob", testUpdateJob)
	dbWrapper(assert, "testDeleteJob", testDeleteJob)
	dbWrapper(assert, "testGetDistinctActiveFunctions", testGetDistinctActiveFunctions)
}

// testAddJob also implicitly tests the GetJobs implementation.
func testAddJob(assert *assert.Assertions, db DB) {
	cronJob := CronJob{
		IsActive: true,
		Function: "test",
		Workload: "--foo bar",
		CronTime: "1 2 3 4 5 6",
		TimeZone: defaultTimeZone,
		Backend:  "gearman",
	}

	// Next three lines included to implicitly test the GetJobs method
	jobs, jobsErr := db.GetJobs(cronJob.Function)
	assert.NoError(jobsErr)
	assert.Len(jobs, 0)

	assert.NoError(db.AddJob(cronJob))

	jobs, jobsErr = db.GetJobs(cronJob.Function)
	assert.NoError(jobsErr)
	assert.Len(jobs, 1)

	compareCronJobs(assert, cronJob, jobs[0])
}

func testDeleteJob(assert *assert.Assertions, db DB) {
	cronJob := CronJob{
		IsActive: true,
		Function: "test",
		Workload: "--foo bar",
		CronTime: "1 2 3 4 5 6",
		TimeZone: defaultTimeZone,
		Backend:  "gearman",
	}

	assert.NoError(db.AddJob(cronJob))

	jobs, getErr := db.GetJobs(cronJob.Function)
	assert.NoError(getErr)
	assert.Len(jobs, 1)

	assert.NoError(db.DeleteJob(jobs[0].ID))

	jobs, getErr = db.GetJobs(cronJob.Function)
	assert.NoError(getErr)
	assert.Len(jobs, 0)
}

func testUpdateJob(assert *assert.Assertions, db DB) {
	cronJobBefore := CronJob{
		IsActive: true,
		Function: "test",
		Workload: "--foo bar",
		CronTime: "1 2 3 4 5 6",
		TimeZone: defaultTimeZone,
		Created:  time.Now(),
		Backend:  "gearman",
	}

	assert.NoError(db.AddJob(cronJobBefore))

	jobsBefore, getErr := db.GetJobs(cronJobBefore.Function)
	assert.NoError(getErr)
	assert.Len(jobsBefore, 1)

	cronJobAfter := CronJob{
		ID:       jobsBefore[0].ID, // An update still needs to match on ID
		IsActive: false,
		Function: "new",
		Workload: "--new car",
		CronTime: "6 5 4 3 2 1",
		TimeZone: defaultTimeZone,
		Created:  time.Now(),
		Backend:  "gearman",
	}

	assert.NoError(db.UpdateJob(cronJobAfter))

	jobsAfter, getErr := db.GetJobs(cronJobAfter.Function)
	assert.NoError(getErr)
	assert.Len(jobsAfter, 1)

	compareCronJobs(assert, jobsAfter[0], cronJobAfter)
	// We call UTC() to abstract away location information and choose RFC3339
	// since it checks times to the second, and we can't guarantee DBs store
	// timestamps much more precisely.
	assert.Equal(jobsAfter[0].Created.UTC().Format(time.RFC3339),
		cronJobAfter.Created.UTC().Format(time.RFC3339))
}

func testGetDistinctActiveFunctions(assert *assert.Assertions, db DB) {
	cronJob1 := CronJob{
		IsActive: true,
		Function: "test1",
		Workload: "--foo bar",
		CronTime: "1 2 3 4 5 6",
		TimeZone: defaultTimeZone,
		Backend:  "gearman",
	}
	cronJob2 := CronJob{
		IsActive: true,
		Function: "test1",
		Workload: "--new car",
		CronTime: "1 2 3 4 5 6",
		TimeZone: defaultTimeZone,
		Backend:  "gearman",
	}
	cronJob3 := CronJob{
		IsActive: false,
		Function: "inactiveFunction",
		Workload: "--foo bar",
		CronTime: "1 2 3 4 5 6",
		TimeZone: defaultTimeZone,
		Backend:  "gearman",
	}
	cronJob4 := CronJob{
		IsActive: true,
		Function: "test2",
		Workload: "--foo bar",
		CronTime: "1 2 3 4 5 6",
		TimeZone: defaultTimeZone,
		Backend:  "gearman",
	}

	assert.NoError(db.AddJob(cronJob1))
	assert.NoError(db.AddJob(cronJob2))
	assert.NoError(db.AddJob(cronJob3))
	assert.NoError(db.AddJob(cronJob4))

	expectedFunctions := []string{"test1", "test2"}
	activeFunctions, getErr := db.GetDistinctActiveFunctions()
	assert.NoError(getErr)
	sort.Strings(activeFunctions)
	assert.Equal(expectedFunctions, activeFunctions)
}

// compareCronJobs compares the meat of two CronJobs. We leave out "ID" and "Created"
// as some tests can't predict what the ID will be and "Created" is tricky because
// either 1) inputs don't have a created time to compare to or 2) some databases impose
// a location which is hard to compare against
func compareCronJobs(assert *assert.Assertions, cBefore, cAfter CronJob) {
	assert.Equal(cBefore.Function, cAfter.Function)
	assert.Equal(cBefore.IsActive, cAfter.IsActive)
	assert.Equal(cBefore.Workload, cAfter.Workload)
	assert.Equal(cBefore.TimeZone, cAfter.TimeZone)
	assert.Equal(cBefore.Backend, cAfter.Backend)
}
