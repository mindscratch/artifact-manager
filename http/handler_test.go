package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	gohttp "net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/mindscratch/artifact-manager/core"
)

func TestUploadHandler_WithoutUrlParams(t *testing.T) {
	req, err := createRequest("../Makefile", "http://localhost/")
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// recorder satisfies the http response interface
	rec := httptest.NewRecorder()

	// handler is some http handler function we wrote that we want to test
	h := NewHandler(core.NewConfig("AM_TEST_"))
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
}

func TestUploadHandler_WithNameUrlParam(t *testing.T) {
	req, err := createRequest("../Makefile", "http://localhost/?name=Makefile")
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// recorder satisfies the http response interface
	rec := httptest.NewRecorder()

	// handler is some http handler function we wrote that we want to test
	h := NewHandler(core.NewConfig("AM_TEST_"))
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
	}

	// ensure the file was created
	if _, err = os.Stat(pathToFile); os.IsNotExist(err) {
		t.Errorf("file should have been created %s, but it does not exist", pathToFile)
	}
}

func TestUploadHandler_WithNameSrcDstUrlParams(t *testing.T) {
	req, err := createRequest("../Makefile", "http://localhost/?name=Makefile&src=Makefile&dst=myfile")
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// recorder satisfies the http response interface
	rec := httptest.NewRecorder()

	// handler is some http handler function we wrote that we want to test
	h := NewHandler(core.NewConfig("AM_TEST_"))
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
	}

	// ensure the file was created
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		t.Errorf("file should have been created %s, but it does not exist", pathToFile)
	}

	// ensure a symlink was made
	stat, err := os.Stat(symlinkDst)
	if os.IsNotExist(err) {
		t.Errorf("symlink should have been created from %s to %s, but it does not exist", symlinkSrc, symlinkDst)
	}
	if stat.Mode()&os.ModeSymlink != 0 {
		t.Errorf("expected %s to be a symlink but it's not", symlinkDst)
	}
}

func TestUploadHandler_WithArchiveNoSymlink(t *testing.T) {
	req, err := createRequest("../_samples/x.tgz", "http://localhost/?name=x.tgz")
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// recorder satisfies the http response interface
	rec := httptest.NewRecorder()

	// handler is some http handler function we wrote that we want to test
	h := NewHandler(core.NewConfig("AM_TEST_"))
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
}

func TestUploadHandler_WithArchiveAndSymlink(t *testing.T) {
	req, err := createRequest("../_samples/x.tgz", "http://localhost/?name=x.tgz&src=sample&dst=x-latest")
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	// recorder satisfies the http response interface
	rec := httptest.NewRecorder()

	// handler is some http handler function we wrote that we want to test
	h := NewHandler(core.NewConfig("AM_TEST_"))
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
