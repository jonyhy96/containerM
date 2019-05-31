package main

import (
	"containerM/selfhandler"
	"containerM/server"
	"fmt"
	"log"
	"net/http"
	"os"

	sh "containerM/selfhandler"
)

func handle(h sh.Handler, r *http.Request, ec chan error) {
	h.Handler(r, ec)
}

var (
	logger      *log.Logger
	defaultPort = 8888
)

func init() {
	os.Setenv("DOCKER_API_VERSION", "1.37")
	if os.Getenv("TOKEN") == "" {
		os.Setenv("TOKEN", "VE9LRU4=")
	}
	if os.Getenv("SECRECT") == "" {
		os.Setenv("SECRECT", "eyJ1c2VybmFtZSI6ICJhZG1pbiIsICJwYXNzd29yZCI6ICJhZG1pbiJ9")
	}
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	ec := make(chan error)
	defer func() {
		if r := recover(); r != nil {
			logger.Println("Recovered err:", r)
		}
	}()
	mux := http.NewServeMux()
	var gh = selfhandler.NewGoHandler(logger)
	gh.SetupRoute(mux, ec)
	srv := server.NewServer(mux, fmt.Sprintf(":%d", defaultPort))
	logger.Printf("Server started at port:%d\n", defaultPort)
	err := srv.ListenAndServe()
	if err != nil {
		logger.Printf("err:%+v\n", err)
	}
}
