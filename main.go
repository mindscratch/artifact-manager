package main

import (
	"os"

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

	artifactsService := artifacts.NewArtifactsService(marathonClient, nil)
	go func() {
		artifactsService.Start(config.MarathonQueryInterval)
	}()

	requestQueue := make(chan string, 100)

	// setup http server
	core.Log("Serving requests at %s", config.ServeAddr())

	handler := http.NewHandler(config, requestQueue, 100)
	err = handler.ListenAndServe()
	core.Log("server failed: %v", err)
}
