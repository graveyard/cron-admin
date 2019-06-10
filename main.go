package main

import (
	"crypto/tls"
	"net"
	"os"
	"time"

	"github.com/Clever/cron-admin/server"
	"github.com/segmentio/go-env"
	"gopkg.in/mgo.v2"
)

func main() {
	dialInfo, err := mgo.ParseURL(env.MustGet("LEGACY_MONGO_URL"))
	if err != nil {
		panic(err)
	}
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return tls.Dial("tcp", addr.String(), &tls.Config{})
	}
	dialInfo.Username = env.MustGet("LEGACY_MONGO_USERNAME")
	dialInfo.Password = env.MustGet("LEGACY_MONGO_PASSWORD")
	dialInfo.Timeout = time.Second * 10

	legacyDB, err := mgo.DialWithInfo(dialInfo)

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "80"
	}

	if serverErr := server.Serve(serverPort, legacyDB); serverErr != nil {
		panic(serverErr)
	}
}
