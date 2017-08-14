package artifacts

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	marathon "github.com/gambol99/go-marathon"
)

// NewMarathonClient configures and returns a Marathon client.
//
// httpClient    - a custom http client may be provided, otherwise pass nil
// debug         - The third-party marathon library being used allows debug
// 				   output to be written to a writer, pass nil if you want debug
//                 for that library to be disabled.
// marathonAddrs - one or more "host:port" addresses used to connect to Marathon
func NewMarathonClient(httpClient *http.Client, debug io.Writer, marathonAddrs ...string) (marathon.Marathon, error) {
	config := marathon.NewDefaultConfig()
	config.URL = fmt.Sprintf("http://%s", strings.Join(marathonAddrs, ","))
	if debug != nil {
		config.LogOutput = debug
	}
	if httpClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: (time.Duration(10) * time.Second),
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 10 * time.Second,
				}).Dial,
			},
		}
	}

	client, err := marathon.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("Failed to create a client for marathon, error: %v", err)
	}
	return client, nil
}

// ArtifactsService represents a service for managing Marathon application
// artifacts.
type ArtifactsService struct {
	artifacts Artifacts
	client    marathon.Marathon
	debug     io.Writer
	mutex     *sync.Mutex
	stopCh    chan struct{}
	stopped   bool
}

// NewArtifactsService configures, creates and returns a new ArtifactsService.
//
// The debug argument allows debug messages to be written to the provided writer,
// pass nil to disable.
func NewArtifactsService(client marathon.Marathon, debug io.Writer) *ArtifactsService {
	if debug == nil {
		debug = ioutil.Discard
	}
	as := ArtifactsService{
		client: client,
		debug:  debug,
		mutex:  &sync.Mutex{},
		stopCh: make(chan struct{}),
	}
	return &as
}

// FetchArtifacts fetches artifacts in use by Marathon applications
// and stores them.
func (as *ArtifactsService) FetchArtifacts() error {
	applications, err := as.client.Applications(url.Values{})
	if err != nil {
		return fmt.Errorf("failed to list applications: %v", err)
	}

	newArtifacts := Artifacts{}

	as.log("Found %d applications running\n", len(applications.Apps))
	var urisCount int
	var fetchCount int
	for _, application := range applications.Apps {
		if application.Uris != nil {
			urisCount = len(*application.Uris)
		} else {
			urisCount = 0
		}
		if application.Fetch != nil {
			fetchCount = len(*application.Fetch)
		} else {
			fetchCount = 0
		}

		as.log("%s has %d URIs and %d Fetch\n", application.ID, urisCount, fetchCount)
		for i := 0; i < urisCount; i++ {
			newArtifacts.Add(application.ID, (*application.Uris)[i])
		}
		for i := 0; i < fetchCount; i++ {
			newArtifacts.Add(application.ID, (*application.Fetch)[i].URI)
		}
	}

	as.mutex.Lock()
	as.artifacts = newArtifacts
	as.mutex.Unlock()

	return nil
}

// GetAppIds returns a list of Marathon application ids that
// rely on an artifact identified by `name`.
func (as *ArtifactsService) GetAppIds(name string) []string {
	as.mutex.Lock()
	appIds := as.artifacts.Get(name)
	as.mutex.Unlock()
	return appIds
}

// HasArtifact returns true if it has the artifact
func (as *ArtifactsService) HasArtifact(name string) bool {
	as.mutex.Lock()
	result := as.artifacts.Has(name)
	as.mutex.Unlock()
	return result
}

// StartFetching starts polling for artifacts after each interval.
func (as *ArtifactsService) StartFetching(interval time.Duration) {
	err := as.FetchArtifacts()
	if err != nil {
		log.Printf("problem fetching marathon artifacts: %v", err)
	}

	for {
		if as.stopped {
			break
		}
		select {
		case <-as.stopCh:
			as.stopped = true
		case <-time.After(interval):
			log.Println("fetching artifacts...")
			err = as.FetchArtifacts()
			log.Println("DONE fetching artifacts...")
			if err != nil {
				log.Printf("problem fetching marathon artifacts: %v", err)
			}
		}
	}
	log.Println("ARTIFACT FETCHING SERVICE HAS STOPPED")
}

// Stop stops the service, passing true to block until it stops.
func (as *ArtifactsService) Stop(block bool) {
	log.Println("STOPPING")
	if block {
		close(as.stopCh)
		log.Println("STOPPED")
	} else {
		go func() {
			close(as.stopCh)
			log.Println("STOPPED")
		}()
	}
	log.Println("exiting stop routine")
}

// StartApplicationRestartProcessing begins waiting for requests on the requestQueue.
// Once the queue has "count" items or "timeout" has occured, the marathon
// applications will be restarted.
func (as *ArtifactsService) StartApplicationRestartProcessing(requestQueue <-chan string, count int, timeout time.Duration) {
	artifactNames := make([]string, 0)
	for {
		if as.stopped {
			break
		}
		select {
		case <-as.stopCh:
			as.stopped = true
		case name := <-requestQueue:
			artifactNames = append(artifactNames, name)
			if len(artifactNames) >= count {
				as.restartApps(artifactNames)
				artifactNames = make([]string, 0)
			}
		case <-time.After(timeout):
			log.Println("Timeout", len(artifactNames))
			if len(artifactNames) >= 0 {
				as.restartApps(artifactNames)
				artifactNames = make([]string, 0)
			}
		}
	}
	log.Println("ARTIFACT RESTART SERVICE HAS STOPPED")
}

func (as *ArtifactsService) restartApps(artifactNames []string) {
	log.Printf("we have %d artifact names that were updated\n", len(artifactNames))
	for _, name := range artifactNames {
		appIds := as.artifacts.Get(name)
		fmt.Printf("found %d app ids associated with %s\n", len(appIds), name)
		for _, appID := range appIds {
			deploymentID, err := as.client.RestartApplication(appID, true)
			if err != nil {
				log.Printf("failed to restart %s: %v\n", appID, err)
				continue
			}
			log.Printf("restarted %s deploymentID=%s version=%s\n", appID, deploymentID.DeploymentID, deploymentID.Version)
		}
	}
}

func (as *ArtifactsService) log(format string, values ...interface{}) {
	as.debug.Write([]byte(fmt.Sprintf(format, values...)))
}
