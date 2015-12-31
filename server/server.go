package server

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"

	"github.com/Clever/cron-admin/db"
)

func byteHandler(handler func(*http.Request) ([]byte, error)) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		data, err := handler(req)
		if err != nil {
			fmt.Printf("Received handler err: %s\n", err)
		}
		rw.Write(data)
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
		activeJobs, err := database.GetDistinctActiveJobs()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		return activeJobs, nil
	})).Methods("GET")

	r.HandleFunc("/job-details", jsonHandler(func(req *http.Request) (interface{}, error) {
		defer req.Body.Close()
		job := req.URL.Query().Get("job")
		jobDetails, err := database.GetJobDetails(job)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		return jobDetails, nil
	})).Methods("GET")

	r.HandleFunc("/modify-job/{job_id}", jsonHandler(func(req *http.Request) (interface{}, error) {
		defer req.Body.Close()
		jobID := mux.Vars(req)["job_id"]
		if parseErr := req.ParseForm(); parseErr != nil {
			fmt.Printf("Got error parsing form: %s", parseErr)
		}
		isActive := req.PostForm.Get("active")
		// Active comes as a string, we need to submit a boolean
		setActive := true
		if isActive == "false" {
			setActive = false
		}

		err := database.UpdateJobActivationStatus(jobID, setActive)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		return nil, nil
	})).Methods("POST")

	r.HandleFunc("/add-job", jsonHandler(func(req *http.Request) (interface{}, error) {
		defer req.Body.Close()
		if parseErr := req.ParseForm(); parseErr != nil {
			fmt.Printf("Got error parsing form: %s", parseErr)
		}
		function := req.PostForm.Get("job")
		crontime := req.PostForm.Get("crontime")
		workloadString := req.PostForm.Get("workload")

		var workload interface{}
		if jsonErr := json.Unmarshal([]byte(workloadString), &workload); jsonErr != nil {
			workload = workloadString
		}

		cronJob := db.CronJob{
			Function: function,
			CronTime: crontime,
			Workload: workload,
			IsActive: true,
			TimeZone: "America/Los_Angeles",
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
	//setup handlers
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
