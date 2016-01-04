package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"

	"github.com/Clever/cron-admin/db"
	"github.com/robfig/cron"
)

var (
	supportedUpdateFields = map[string]bool{
		"Workload": true,
		"IsActive": true,
		"CronTime": true,
	}
	ErrEmptyFunctionInput = errors.New("Error: Must include non-empty function")
)

func byteHandler(handler func(*http.Request) ([]byte, error)) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		data, err := handler(req)
		if err != nil {
			fmt.Printf("Received handler err: %s\n", err)
			rw.WriteHeader(500)
			rw.Write([]byte(err.Error()))
		} else {
			rw.Write(data)
		}
	}
}

func jsonHandler(handler func(*http.Request) (interface{}, error)) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		wrappedHandler := func(req *http.Request) ([]byte, error) {
			resp, err := handler(req)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal(resp)
			if err != nil {
				return nil, err
			}
			rw.Header().Add("Content-Type", "application/json; charset=utf-8")
			return data, nil
		}
		byteHandler(wrappedHandler)(rw, req)
	}
}

func SetupHandlers(r *mux.Router, database db.DB) {
	r.HandleFunc("/active-jobs", jsonHandler(func(req *http.Request) (interface{}, error) {
		defer req.Body.Close()
		activeJobs, getErr := database.GetDistinctActiveJobs()
		if getErr != nil {
			fmt.Println(getErr)
			return nil, getErr
		}
		return activeJobs, nil
	})).Methods("GET")

	r.HandleFunc("/jobs", jsonHandler(func(req *http.Request) (interface{}, error) {
		defer req.Body.Close()
		job := req.URL.Query().Get("Function")
		jobDetails, getErr := database.GetJobDetails(job)
		if getErr != nil {
			fmt.Println(getErr)
			return nil, getErr
		}
		return jobDetails, nil
	})).Methods("GET")

	r.HandleFunc("/jobs/{job_id}", jsonHandler(func(req *http.Request) (interface{}, error) {
		defer req.Body.Close()
		if parseErr := req.ParseForm(); parseErr != nil {
			fmt.Printf("Got error parsing form: %s", parseErr)
		}
		jobID := mux.Vars(req)["job_id"]

		// Filter out post body to include only supported update fields
		fieldMap := make(map[string]interface{})
		for k, _ := range req.PostForm {
			if _, ok := supportedUpdateFields[k]; !ok {
				continue
			}

			v := req.PostForm.Get(k)
			if k == "IsActive" {
				// Input will be in string form, but IsActive should be boolean
				boolVal, convErr := strconv.ParseBool(v)
				if convErr != nil {
					return nil, convErr
				}
				fieldMap[k] = boolVal
			} else {
				fieldMap[k] = v
			}
		}

		updateErr := database.UpdateJob(jobID, fieldMap)
		if updateErr != nil {
			fmt.Println(updateErr)
			return nil, updateErr
		}
		return nil, nil
	})).Methods("PUT")

	r.HandleFunc("/jobs/{job_id}", jsonHandler(func(req *http.Request) (interface{}, error) {
		defer req.Body.Close()
		jobID := mux.Vars(req)["job_id"]
		if removeErr := database.DeleteJob(jobID); removeErr != nil {
			return nil, removeErr
		}
		return nil, nil
	})).Methods("DELETE")

	r.HandleFunc("/jobs", jsonHandler(func(req *http.Request) (interface{}, error) {
		defer req.Body.Close()
		if parseErr := req.ParseForm(); parseErr != nil {
			fmt.Printf("Got error parsing form: %s", parseErr)
		}
		function := req.PostForm.Get("Function")
		if function == "" {
			return nil, ErrEmptyFunctionInput
		}
		cronTime := req.PostForm.Get("CronTime")
		if _, cronErr := cron.Parse(cronTime); cronErr != nil {
			return nil, cronErr
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
			return nil, insertErr
		}
		return nil, nil
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
