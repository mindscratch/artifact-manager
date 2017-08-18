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
	volumes Volumes
	client  marathon.Marathon
	debug   io.Writer
	mutex   *sync.Mutex
	stopCh  chan struct{}
	stopped bool
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

// FetchVolumes fetches volumes in use by Marathon applications
// and stores them.
func (as *ArtifactsService) FetchVolumes() (int, error) {
	applications, err := as.client.Applications(url.Values{})
	if err != nil {
		return 0, fmt.Errorf("failed to list applications: %v", err)
	}

	newVolumes := Volumes{}

	as.log("Found %d applications running\n", len(applications.Apps))
	for _, application := range applications.Apps {
		if application.Container != nil {
			volumes := application.Container.Volumes
			if volumes != nil && len(*volumes) > 0 {
				as.log("%s depends on %d volumes\n", application.ID, len(*volumes))
				for _, volume := range *volumes {
					if volume.HostPath != "" {
						as.log("Adding %s path for %s\n", volume.HostPath, application.ID)
						newVolumes.Add(application.ID, volume.HostPath)
					}
				}
			}
		}
	}

	count := len(newVolumes)

	as.mutex.Lock()
	as.volumes = newVolumes
	as.mutex.Unlock()

	return count, nil
}

// GetAppIds returns a list of Marathon application ids that
// rely on an artifact identified by `path`.
func (as *ArtifactsService) GetAppIds(path string) []string {
	as.mutex.Lock()
	appIds := as.volumes.Get(path)
	as.mutex.Unlock()
	return appIds
}

// HasArtifact returns true if it has the artifact
func (as *ArtifactsService) HasArtifact(path string) bool {
	as.mutex.Lock()
	result := as.volumes.Has(path)
	as.mutex.Unlock()
	return result
}

// StartFetching starts polling for artifacts after each interval.
func (as *ArtifactsService) StartFetching(interval time.Duration) {
	log.Println("fetching volumes depended on by applications for the first time...")
	numVolumes, err := as.FetchVolumes()
	log.Printf("DONE fetching volumes, found %d.\n", numVolumes)
	if err != nil {
		log.Printf("problem fetching volumes depended on by applications: %v", err)
	}

	for {
		if as.stopped {
			break
		}
		select {
		case <-as.stopCh:
			as.stopped = true
		case <-time.After(interval):
			log.Println("fetching volumes depended on by applications...")
			numVolumes, err = as.FetchVolumes()
			log.Printf("DONE fetching volumes, found %d.\n", numVolumes)
			if err != nil {
				log.Printf("problem fetching marathon artifacts: %v", err)
			}
		}
	}
	log.Println("VOLUME FETCHING SERVICE HAS STOPPED")
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
	paths := make([]string, 0)
	for {
		if as.stopped {
			break
		}
		select {
		case <-as.stopCh:
			as.stopped = true
		case path := <-requestQueue:
			paths = append(paths, path)
			if len(paths) >= count {
				as.restartApps(paths)
				paths = make([]string, 0)
			}
		case <-time.After(timeout):
			log.Printf("Timeout after %s, have %d paths that have been updated", timeout, len(paths))
			if len(paths) > 0 {
				as.restartApps(paths)
				paths = make([]string, 0)
			}
		}
	}
	log.Println("RESTART SERVICE HAS STOPPED")
}

func (as *ArtifactsService) restartApps(paths []string) {
	log.Printf("we have %d paths that were updated\n", len(paths))
	for _, path := range paths {
		appIds := as.volumes.Get(path)
		fmt.Printf("found %d app ids depending on %s\n", len(appIds), path)
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
