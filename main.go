package main

import (
	"github.com/Clever/cron-admin/server"
	"github.com/segmentio/go-env"
	"os"
	"runtime"
)

// init panics if required environmental variables are not set before running
func init() {
	requiredEnv := []string{"MONGO_URL"}
	for _, envVar := range requiredEnv {
		// panic if env not set
		env.MustGet(envVar)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//start database connection
	mongoURL := env.MustGet("MONGO_URL")

	//get port info
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "80"
	}

	err := server.Serve(serverPort, mongoURL)
	if err != nil {
		panic(err)
	}
}
