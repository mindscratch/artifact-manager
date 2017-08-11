package main

import (
	"os"

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

	core.Log("Serving requests at %s", config.ServeAddr())

	handler := http.NewHandler(config)
	err = handler.ListenAndServe()
	core.Log("server failed: %v", err)
}
