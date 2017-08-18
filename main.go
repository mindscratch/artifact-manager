package main

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/mindscratch/artifact-manager/artifacts"
	"github.com/mindscratch/artifact-manager/core"
	"github.com/mindscratch/artifact-manager/http"
)

func main() {
	//TODO: add metrics, maybe record trace events
	// default < command line < env var

	//TODO REST API:
	// src, dst -- needed for symlink
	// restart apps - true (default - with 'force')
	//
	// Create go routine worker pool

	config := core.NewConfig("AM_")
	err := config.Parse()
	if err != nil {
		core.Log("problem with application configuration. %v", err)
		os.Exit(1)
	}

	// create marathon client and artifacts service
	marathonClient, err := artifacts.NewMarathonClient(nil, nil, config.MarathonHosts)
	if err != nil {
		core.Log("problem creating marathon client. %v", err)
		os.Exit(1)
	}

	debugWriter := ioutil.Discard
	if config.Debug {
		debugWriter = os.Stdout
	}
	artifactsService := artifacts.NewArtifactsService(marathonClient, debugWriter)
	go func() {
		artifactsService.StartFetching(config.MarathonQueryInterval)
	}()

	requestQueue := make(chan string, 100)

	go func() {
		artifactsService.StartApplicationRestartProcessing(requestQueue, 5, 5*time.Second)
	}()

	// setup http server
	core.Log("Serving requests at %s", config.ServeAddr())

	handler := http.NewHandler(config, requestQueue, 100)
	err = handler.ListenAndServe()
	core.Log("server failed: %v", err)
}
