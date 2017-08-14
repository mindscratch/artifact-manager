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

// Start the service, polling for artifacts after each interval.
func (as *ArtifactsService) Start(interval time.Duration) {
	err := as.FetchArtifacts()
	if err != nil {
		log.Printf("problem fetching marathon artifacts: %v", err)
	}
	stop := false

	for {
		if stop {
			break
		}
		select {
		case <-as.stopCh:
			stop = true
		case <-time.After(interval):
			log.Println("fetching artifacts...")
			err = as.FetchArtifacts()
			log.Println("DONE fetching artifacts...")
			if err != nil {
				log.Printf("problem fetching marathon artifacts: %v", err)
			}
		}
	}
	log.Println("ARTIFACT SERVICE HAS STOPPED")
}

// Stop the service, passing true to block until it stops.
func (as *ArtifactsService) Stop(block bool) {
	log.Println("STOPPING")
	if block {
		//as.stopCh <- struct{}{}
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

func (as *ArtifactsService) log(format string, values ...interface{}) {
	as.debug.Write([]byte(fmt.Sprintf(format, values...)))
}
