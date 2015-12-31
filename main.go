package main

import (
	"github.com/Clever/cron-admin/server"
	"github.com/segmentio/go-env"
	"os"
)

// init panics if required environmental variables are not set before running
func init() {
	requiredEnv := []string{"MONGO_URL"}
	for _, envVar := range requiredEnv {
		env.MustGet(envVar)
	}
}

func main() {
	mongoURL := env.MustGet("MONGO_URL")
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "80"
	}
	serverErr := server.Serve(serverPort, mongoURL)
	if serverErr != nil {
		panic(serverErr)
	}
}
