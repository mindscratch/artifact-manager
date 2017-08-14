package artifacts

import (
	"fmt"
	"io/ioutil"
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

	// create the artifacts service and fetch the artifacts
	svc := NewArtifactsService(marathonClient, nil)
	err = svc.FetchArtifacts()
	if err != nil {
		t.Errorf("failed to fetch artifacts: %v", err)
	}

	// ensure it has the correct artifacts
	artifactName := "data.txt"
	if !svc.HasArtifact(artifactName) {
		t.Errorf("expected to have applications associated with artifact named %s", artifactName)
		return
	}

	appIds := svc.GetAppIds(artifactName)
	if len(appIds) != 2 {
		t.Errorf("expected two applications to depend on %s, got %d", artifactName, len(appIds))
		return
	}

	appID := "/myapp-txt"
	if appIds[0] != appID {
		t.Errorf("expected the first application to be %s, got %s", appID, appIds[0])
		return
	}
	appID = "/backened/another-txt"
	if appIds[1] != appID {
		t.Errorf("expected the second application to be %s, got %s", appID, appIds[1])
	}
}
