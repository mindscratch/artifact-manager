package main

import (
	"os"

	"github.com/mindscratch/artifact-manager/core"
)

func main() {
	//TODO: add metrics, maybe record trace events
	// default < command line < env var

	//TODO REST API:
	// src, dst -- needed for symlink
	// restart apps - true (default - with 'force')
	//
	// check file when uploaded to verify it has a directory inside

	config := core.NewConfig("AM_")
	err := config.Parse()
	if err != nil {
		core.Log("problem with application configuration. %v", err)
		os.Exit(1)
	}

	core.Log("this is a test abc=%d", 123)
	core.Log("config: %#v", config)
}
