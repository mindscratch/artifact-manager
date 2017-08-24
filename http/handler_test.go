package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	gohttp "net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"apex/artifact-manager/core"
)

// TestUploadHandler_WithoutUrlParams tests the behavior when a file is uploaded
// with no URL parameters provided.
func TestUploadHandler_WithoutUrlParams(t *testing.T) {
	req, err := createRequest("../Makefile", "http://localhost/")
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// recorder satisfies the http response interface
	rec := httptest.NewRecorder()

	// handler is some http handler function we wrote that we want to test
	requestQueue := make(chan string, 10)
	h := NewHandler(core.NewConfig("AM_TEST_"), requestQueue, 10, log.New(ioutil.Discard, log.Prefix(), log.Flags()))
	h.UploadHandler(rec, req)

	if rec.Code != gohttp.StatusBadRequest {
		msg := fmt.Sprintf("expected status %d; got %d", gohttp.StatusBadRequest, rec.Code)
		resp, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			msg = fmt.Sprintf("%s: could not read http response body: %v", msg, err)
		} else {
			msg = fmt.Sprintf("%s: response=%s", msg, string(resp))
		}
		t.Errorf(msg)
	}
	if len(requestQueue) != 0 {
		t.Errorf("expected requestQueue channel to be empty; got %d", len(requestQueue))
	}
}

// TestUploadHandler_WithNameUrlParam tests the behavior when a file is
// uploaded and the name URL parameter is provided.
func TestUploadHandler_WithNameUrlParam(t *testing.T) {
	req, err := createRequest("../Makefile", "http://localhost/?name=Makefile")
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// recorder satisfies the http response interface
	rec := httptest.NewRecorder()

	// handler is some http handler function we wrote that we want to test
	requestQueue := make(chan string, 10)
	h := NewHandler(core.NewConfig("AM_TEST_"), requestQueue, 10, log.New(ioutil.Discard, log.Prefix(), log.Flags()))
	pathToFile := path.Join(h.config.Dir, "Makefile")
	defer os.Remove(pathToFile)
	h.UploadHandler(rec, req)

	if rec.Code != gohttp.StatusCreated {
		msg := fmt.Sprintf("expected status %d; got %d", gohttp.StatusCreated, rec.Code)
		resp, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			msg = fmt.Sprintf("%s: could not read http response body: %v", msg, err)
		} else {
			msg = fmt.Sprintf("%s: response=%s", msg, string(resp))
		}
		t.Errorf(msg)
	} else {
		// ensure the file was created
		if _, err = os.Stat(pathToFile); os.IsNotExist(err) {
			t.Errorf("file should have been created %s, but it does not exist", pathToFile)
		}
	}

	if len(requestQueue) != 1 {
		t.Errorf("expected requestQueue channel to have 1 message; got %d", len(requestQueue))
	} else {
		result := <-requestQueue
		if result != pathToFile {
			t.Errorf("expected message on requestQueue to be %s; got %v", pathToFile, result)
		}
	}
}

// TestUploadHandler_WithNameSrcDstUrlParams tests the behavior when a file
// is uploaded and the name, src and dst URL parameters are provided.
func TestUploadHandler_WithNameSrcDstUrlParams(t *testing.T) {
	req, err := createRequest("../Makefile", "http://localhost/?name=Makefile&src=Makefile&dst=myfile")
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// recorder satisfies the http response interface
	rec := httptest.NewRecorder()

	// handler is some http handler function we wrote that we want to test
	requestQueue := make(chan string, 10)
	h := NewHandler(core.NewConfig("AM_TEST_"), requestQueue, 10, log.New(ioutil.Discard, log.Prefix(), log.Flags()))
	pathToFile := path.Join(h.config.Dir, "Makefile")
	symlinkSrc := path.Join(h.config.Dir, "Makefile")
	symlinkDst := path.Join(h.config.Dir, "myfile")
	defer func() {
		os.RemoveAll(symlinkDst)
		os.RemoveAll(symlinkSrc)
		os.Remove(pathToFile)
	}()
	h.UploadHandler(rec, req)

	if rec.Code != gohttp.StatusCreated {
		msg := fmt.Sprintf("expected status %d; got %d", gohttp.StatusCreated, rec.Code)
		resp, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			msg = fmt.Sprintf("%s: could not read http response body: %v", msg, err)
		} else {
			msg = fmt.Sprintf("%s: response=%s", msg, string(resp))
		}
		t.Errorf(msg)
		return
	}

	// ensure the file was created
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Errorf("file should have been created %s, but it does not exist", pathToFile)
		return
	}

	// ensure a symlink was made
	stat, err := os.Stat(symlinkDst)
	if os.IsNotExist(err) {
		t.Errorf("symlink should have been created from %s to %s, but it does not exist", symlinkSrc, symlinkDst)
		return
	}
	if stat.Mode()&os.ModeSymlink != 0 {
		t.Errorf("expected %s to be a symlink but it's not", symlinkDst)
		return
	}

	if len(requestQueue) != 1 {
		t.Errorf("expected requestQueue channel to have 1 message; got %d", len(requestQueue))
	} else {
		result := <-requestQueue
		if result != symlinkDst {
			t.Errorf("expected message on requestQueue to be %s; got %v", pathToFile, symlinkDst)
		}
	}
}

// TestUploadHandler_WithArchiveNoSymlink tests the behavior when an archive
// file is uploaded without the URL parameters that would cause a symlink
// to get created.
func TestUploadHandler_WithArchiveNoSymlink(t *testing.T) {
	req, err := createRequest("../_samples/x.tgz", "http://localhost/?name=x.tgz")
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// recorder satisfies the http response interface
	rec := httptest.NewRecorder()

	// handler is some http handler function we wrote that we want to test
	requestQueue := make(chan string, 10)
	h := NewHandler(core.NewConfig("AM_TEST_"), requestQueue, 10, log.New(ioutil.Discard, log.Prefix(), log.Flags()))
	pathToFile := path.Join(h.config.Dir, "x.tgz")
	defer func() {
		os.Remove(pathToFile)
	}()
	h.UploadHandler(rec, req)

	if rec.Code != gohttp.StatusCreated {
		msg := fmt.Sprintf("expected status %d; got %d", gohttp.StatusCreated, rec.Code)
		resp, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			msg = fmt.Sprintf("%s: could not read http response body: %v", msg, err)
		} else {
			msg = fmt.Sprintf("%s: response=%s", msg, string(resp))
		}
		t.Errorf(msg)
	}

	// ensure the file was created
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Errorf("file should have been created %s, but it does not exist", pathToFile)
	}

	if len(requestQueue) != 1 {
		t.Errorf("expected requestQueue channel to have 1 message; got %d", len(requestQueue))
	} else {
		result := <-requestQueue
		if result != pathToFile {
			t.Errorf("expected message on requestQueue to be %s; got %v", pathToFile, pathToFile)
		}
	}
}

// TestUploadHandler_WithArchiveAndSymlink tests the behavior when an archive
// file is uploaded with the URL parameters required to create a symlink to
// the extracted content.
func TestUploadHandler_WithArchiveAndSymlink(t *testing.T) {
	req, err := createRequest("../_samples/x.tgz", "http://localhost/?name=x.tgz&src=sample&dst=x-latest")
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// recorder satisfies the http response interface
	rec := httptest.NewRecorder()

	// handler is some http handler function we wrote that we want to test
	requestQueue := make(chan string, 10)
	h := NewHandler(core.NewConfig("AM_TEST_"), requestQueue, 10, log.New(ioutil.Discard, log.Prefix(), log.Flags()))
	pathToFile := path.Join(h.config.Dir, "x.tgz")
	symlinkSrc := path.Join(h.config.Dir, "sample")
	symlinkDst := path.Join(h.config.Dir, "x-latest")
	defer func() {
		os.RemoveAll(symlinkDst)
		os.RemoveAll(symlinkSrc)
		os.Remove(pathToFile)
	}()
	h.UploadHandler(rec, req)

	if rec.Code != gohttp.StatusCreated {
		msg := fmt.Sprintf("expected status %d; got %d", gohttp.StatusCreated, rec.Code)
		resp, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			msg = fmt.Sprintf("%s: could not read http response body: %v", msg, err)
		} else {
			msg = fmt.Sprintf("%s: response=%s", msg, string(resp))
		}
		t.Errorf(msg)
	}

	// ensure the file was created
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Errorf("file should have been created %s, but it does not exist", pathToFile)
	}

	// ensure it was extracted
	stat, err := os.Stat(symlinkSrc)
	if os.IsNotExist(err) {
		t.Errorf("file %s should have been extracted to %s, but it does not exist", pathToFile, symlinkSrc)
	} else if err != nil {
		t.Errorf("failed to check on %s: %v", symlinkSrc, err)
	} else if !stat.IsDir() {
		t.Errorf("extracted file %s was expected to contain a directory, but %s is not a directory", pathToFile, symlinkSrc)
	} else {
		// ensure a symlink was made
		stat, err = os.Stat(symlinkDst)
		if os.IsNotExist(err) {
			t.Errorf("symlink should have been created from %s to %s, but it does not exist", symlinkSrc, symlinkDst)
		} else if stat.Mode()&os.ModeSymlink != 0 {
			t.Errorf("expected %s to be a symlink but it's not", symlinkDst)
		}
	}

	if len(requestQueue) != 1 {
		t.Errorf("expected requestQueue channel to have 1 message; got %d", len(requestQueue))
	} else {
		result := <-requestQueue
		if result != symlinkDst {
			t.Errorf("expected message on requestQueue to be %s; got %v", pathToFile, symlinkDst)
		}
	}
}

// TestUploadHandler_ExceedRequestLImit tests the behavior of a file being
// uploaded when the requestQueue is full.
func TestUploadHandler_ExceedRequestLImit(t *testing.T) {
	req, err := createRequest("../Makefile", "http://localhost/?name=Makefile")
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// recorder satisfies the http response interface
	rec := httptest.NewRecorder()

	// handler is some http handler function we wrote that we want to test
	requestQueue := make(chan string, 1)
	// simulate a request having already been put onto the queue
	requestQueue <- "this is a test"

	h := NewHandler(core.NewConfig("AM_TEST_"), requestQueue, 1, log.New(ioutil.Discard, log.Prefix(), log.Flags()))
	pathToFile := path.Join(h.config.Dir, "Makefile")
	defer os.Remove(pathToFile)
	h.UploadHandler(rec, req)

	if rec.Code != gohttp.StatusServiceUnavailable {
		msg := fmt.Sprintf("expected status %d; got %d", gohttp.StatusServiceUnavailable, rec.Code)
		resp, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			msg = fmt.Sprintf("%s: could not read http response body: %v", msg, err)
		} else {
			msg = fmt.Sprintf("%s: response=%s", msg, string(resp))
		}
		t.Errorf(msg)
	}

	// ensure the file was not created
	if _, err = os.Stat(pathToFile); os.IsExist(err) {
		t.Errorf("file should not have been created %s, but it does exist", pathToFile)
	}
}

func createRequest(file, url string) (*gohttp.Request, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file (%s) for use in testing: %v", file, err)
	}

	reader := bytes.NewReader(data)
	return gohttp.NewRequest(
		gohttp.MethodPost,
		url,
		reader,
	)
}
