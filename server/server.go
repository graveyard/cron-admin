package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Clever/cron-admin/db"
	"github.com/robfig/cron"
)

var (
	cronUpdateFields = []string{
		"IsActive",
		"Function",
		"Workload",
		"CronTime",
		"TimeZone",
		"Created",
	}
	ErrEmptyFunctionInput  = fmt.Errorf("Error: Must include non-empty function")
	ErrMissingUpdateFields = fmt.Errorf("Error: Updates require you to include the following:  %s", strings.Join(cronUpdateFields, ", "))
)

func byteHandler(handler func(*http.Request) ([]byte, int, error)) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		data, statusCode, err := handler(req)
		if err != nil {
			fmt.Printf("Received handler err: %s\n", err)
			rw.WriteHeader(statusCode)
			rw.Write([]byte(err.Error()))
		} else {
			rw.WriteHeader(statusCode)
			rw.Write(data)
		}
	}
}

func jsonHandler(handler func(*http.Request) (interface{}, int, error)) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		wrappedHandler := func(req *http.Request) ([]byte, int, error) {
			resp, statusCode, err := handler(req)
			if err != nil {
				return nil, statusCode, err
			}
			data, err := json.Marshal(resp)
			if err != nil {
				return nil, statusCode, err
			}
			rw.Header().Add("Content-Type", "application/json; charset=utf-8")
			return data, statusCode, nil
		}
		byteHandler(wrappedHandler)(rw, req)
	}
}

func SetupHandlers(r *mux.Router, database db.DB) {
	r.HandleFunc("/active-functions", jsonHandler(func(req *http.Request) (interface{}, int, error) {
		defer req.Body.Close()
		activeJobs, getErr := database.GetDistinctActiveFunctions()
		if getErr != nil {
			fmt.Println(getErr)
			return nil, 500, getErr
		}
		return activeJobs, 200, nil
	})).Methods("GET")

	r.HandleFunc("/jobs", jsonHandler(func(req *http.Request) (interface{}, int, error) {
		defer req.Body.Close()
		function := req.URL.Query().Get("Function")
		if function == "" {
			return nil, 400, ErrEmptyFunctionInput
		}
		jobDetails, getErr := database.GetJobDetails(function)
		if getErr != nil {
			fmt.Println(getErr)
			return nil, 500, getErr
		}
		return jobDetails, 200, nil
	})).Methods("GET")

	r.HandleFunc("/jobs/{job_id}", jsonHandler(func(req *http.Request) (interface{}, int, error) {
		defer req.Body.Close()
		if parseErr := req.ParseForm(); parseErr != nil {
			fmt.Printf("Got error parsing form: %s", parseErr)
		}
		jobID := mux.Vars(req)["job_id"]

		// Check all required fields are included in request
		for _, field := range cronUpdateFields {
			if _, ok := req.PostForm[field]; !ok {
				return nil, 400, ErrMissingUpdateFields
			}
		}

		function := req.PostForm.Get("Function")
		if function == "" {
			return nil, 400, ErrEmptyFunctionInput
		}

		isActive, convErr := strconv.ParseBool(req.PostForm.Get("IsActive"))
		if convErr != nil {
			return nil, 400, convErr
		}

		created, timeErr := time.Parse(time.RFC3339, req.PostForm.Get("Created"))
		if timeErr != nil {
			return nil, 400, timeErr
		}

		cronJob := db.CronJob{
			ID:       jobID,
			IsActive: isActive,
			Function: function,
			Workload: req.PostForm.Get("Workload"),
			CronTime: req.PostForm.Get("CronTime"),
			TimeZone: req.PostForm.Get("TimeZone"),
			Created:  created,
		}

		updateErr := database.UpdateJob(cronJob)
		if updateErr != nil {
			fmt.Println(updateErr)
			return nil, 500, updateErr
		}
		return nil, 200, nil
	})).Methods("PUT")

	r.HandleFunc("/jobs/{job_id}", jsonHandler(func(req *http.Request) (interface{}, int, error) {
		defer req.Body.Close()
		jobID := mux.Vars(req)["job_id"]
		if removeErr := database.DeleteJob(jobID); removeErr != nil {
			return nil, 500, removeErr
		}
		return nil, 200, nil
	})).Methods("DELETE")

	r.HandleFunc("/jobs", jsonHandler(func(req *http.Request) (interface{}, int, error) {
		defer req.Body.Close()
		if parseErr := req.ParseForm(); parseErr != nil {
			fmt.Printf("Got error parsing form: %s", parseErr)
		}
		function := req.PostForm.Get("Function")
		if function == "" {
			return nil, 400, ErrEmptyFunctionInput
		}
		cronTime := req.PostForm.Get("CronTime")
		if _, cronErr := cron.Parse(cronTime); cronErr != nil {
			return nil, 400, cronErr
		}
		workload := req.PostForm.Get("Workload")
		timeZone := req.PostForm.Get("TimeZone")
		if timeZone == "" {
			timeZone = "America/Los_Angeles"
		}

		cronJob := db.CronJob{
			Function: function,
			CronTime: cronTime,
			Workload: workload,
			IsActive: true,
			TimeZone: timeZone,
			Created:  time.Now(),
		}

		if insertErr := database.AddJob(cronJob); insertErr != nil {
			fmt.Printf("Error inserting job: %s", insertErr)
			return nil, 500, insertErr
		}
		return nil, 200, nil
	})).Methods("POST")
}

// Serve starts up a server. This call won't return unless there's an error.
func Serve(serverPort string, mongoURL string) error {
	r := mux.NewRouter()

	database, err := db.NewMongoDB(mongoURL, "clever")
	if err != nil {
		fmt.Println(err)
	}
	SetupHandlers(r, database)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		http.ServeFile(w, r, "./static/index.html")
	}).Methods("GET")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
	http.Handle("/", r)
	fmt.Printf("Starting cron-admin on port %s\n", serverPort)
	return http.ListenAndServe(":"+serverPort, nil)
}
