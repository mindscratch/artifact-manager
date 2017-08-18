package artifacts

import (
	"fmt"
	"io/ioutil"
	"log"
	gohttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestArtifactsService_FetchArtifacts(t *testing.T) {
	// create the "mock" marathon server
	s := httptest.NewServer(gohttp.HandlerFunc(func(w gohttp.ResponseWriter, r *gohttp.Request) {
		file := "../_samples/apps.json"
		data, err := ioutil.ReadFile(file)
		if err != nil {
			w.WriteHeader(gohttp.StatusInternalServerError)
			fmt.Fprintf(w, "could not read file (%s) for use in testing: %v", file, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(gohttp.StatusOK)
		fmt.Fprintf(w, string(data))
	}))
	defer s.Close()

	// create a marathon client
	u, err := url.Parse(s.URL)
	if err != nil {
		t.Errorf("unable to parse mock server url %s: %v", s.URL, err)
		return
	}
	marathonClient, err := NewMarathonClient(nil, nil, u.Host)
	if err != nil {
		t.Errorf("unable to create marathon client: %v", err)
		return
	}

	// create the artifacts service and fetch the volumes
	svc := NewArtifactsService(marathonClient, log.New(ioutil.Discard, log.Prefix(), log.Flags()))
	count, err := svc.FetchVolumes()
	if err != nil {
		t.Errorf("failed to fetch volumes: %v", err)
		return
	}

	if count != 1 {
		t.Errorf("expected there to be %d volume, got %d", 1, count)
		return
	}

	// ensure it has the correct volumes
	volumeHostPath := "/data/models/mymodel-latest"
	if !svc.HasArtifact(volumeHostPath) {
		t.Errorf("expected to have applications associated with path named %s", volumeHostPath)
		return
	}

	appIds := svc.GetAppIds(volumeHostPath)
	if len(appIds) != 1 {
		t.Errorf("expected one application to depend on %s, got %d", volumeHostPath, len(appIds))
		return
	}

	appID := "/myapp"
	if appIds[0] != appID {
		t.Errorf("expected the first application to be %s, got %s", appID, appIds[0])
		return
	}
}
