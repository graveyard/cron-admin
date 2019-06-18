package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Clever/cron-admin/server"
	"gopkg.in/mgo.v2"
)

func getEnvOrFail(envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		log.Fatalf("environment variable '%s' must be set", envVar)
	}
	return value
}

func dialAtlas(envVarPrefix string) (*mgo.Session, error) {
	mongoURL := getEnvOrFail(fmt.Sprintf("%s_MONGO_URL", envVarPrefix))
	username := getEnvOrFail(fmt.Sprintf("%s_MONGO_USERNAME", envVarPrefix))
	password := getEnvOrFail(fmt.Sprintf("%s_MONGO_PASSWORD", envVarPrefix))

	dialInfo, err := mgo.ParseURL(mongoURL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing mongo url - error: '%s', url='%s'", err, mongoURL)
	}
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return tls.Dial("tcp", addr.String(), &tls.Config{})
	}
	dialInfo.Username = username
	dialInfo.Password = password

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to mongo - error: '%s', url='%s'", err, mongoURL)
	}

	return session, nil
}

func main() {
	legacyDB, err := dialAtlas("LEGACY")
	if err != nil {
		log.Fatalf("failed to connect to Legacy DB: %s", err)
	}

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "80"
	}

	if serverErr := server.Serve(serverPort, legacyDB); serverErr != nil {
		panic(serverErr)
	}
}
