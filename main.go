package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"apex/artifact-manager/artifacts"
	"apex/artifact-manager/core"
	"apex/artifact-manager/http"
)

func main() {
	// TODO: add metrics, maybe record trace events
	// TODO: Create go routine worker pool
	// note: default < command line < env var

	config := core.NewConfig("AM_")
	err := config.Parse()
	if err != nil {
		core.Log("problem with application configuration. %v", err)
		os.Exit(1)
	}

	// setup a logger for recording debug/verbose messages
	debugWriter := ioutil.Discard
	if config.Debug {
		debugWriter = os.Stdout
	}
	debugLogger := log.New(debugWriter, "DEBUG ", log.Flags())

	goMarathonDebugWriter := ioutil.Discard
	if config.MarathonDebug {
		goMarathonDebugWriter = os.Stdout
	}

	// create marathon client and artifacts service
	marathonClient, err := artifacts.NewMarathonClient(nil, goMarathonDebugWriter, config.MarathonHosts)
	if err != nil {
		core.Log("problem creating marathon client. %v", err)
		os.Exit(1)
	}

	artifactsService := artifacts.NewArtifactsService(marathonClient, debugLogger)
	go func() {
		artifactsService.StartFetching(config.MarathonQueryInterval)
	}()

	requestQueue := make(chan string, 100)

	go func() {
		artifactsService.StartApplicationRestartProcessing(requestQueue, 5, 5*time.Second)
	}()

	// setup http server
	core.Log("Serving requests at %s", config.ServeAddr())

	handler := http.NewHandler(config, requestQueue, 100, debugLogger)
	err = handler.ListenAndServe()
	core.Log("server failed: %v", err)
}
