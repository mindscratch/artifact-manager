package http

import (
	"fmt"
	"io"
	"io/ioutil"
	gohttp "net/http"

	"github.com/mindscratch/artifact-manager/core"
)

type Handler struct {
	config *core.Config
}

func NewHandler(config *core.Config) *Handler {
	h := Handler{
		config: config,
	}
	return &h
}

// UploadHandler handles file upload requests
func (h *Handler) UploadHandler(w gohttp.ResponseWriter, r *gohttp.Request) {
	// only accept POST
	if r.Method != gohttp.MethodPost {
		w.WriteHeader(gohttp.StatusMethodNotAllowed)
		fmt.Fprintf(w, "only %s is allowed", gohttp.MethodPost)
		return
	}
	// check to ensure content is being provided
	if r.ContentLength <= 0 {
		w.WriteHeader(gohttp.StatusBadRequest)
		fmt.Fprintf(w, "invalid request, no content provided")
		return
	}

	// create a temporary file to copy the request contents into
	f, err := ioutil.TempFile(h.config.Dir, h.config.EnvVarPrefix)
	if err != nil {
		w.WriteHeader(gohttp.StatusInternalServerError)
		fmt.Fprintf(w, "unable to create file to write request body into: %v", err)
		return
	}
	defer f.Close()

	// copy the request body into the file
	written, err := io.Copy(f, r.Body)
	if err != nil {
		w.WriteHeader(gohttp.StatusInternalServerError)
		fmt.Fprintf(w, "failed to write request body to %s: %v", f.Name(), err)
		return
	}
	if written != r.ContentLength {
		w.WriteHeader(gohttp.StatusInternalServerError)
		fmt.Fprintf(w, "failed to write entire request body to %s, wrote=%d bytes, expected=%d bytes", f.Name(), written, r.ContentLength)
		return
	}

	fmt.Fprintf(w, "wrote file to %s\n", f.Name())
}

// ListenAndServe starts the server
func (h *Handler) ListenAndServe() error {
	gohttp.HandleFunc("/", h.UploadHandler)

	return gohttp.ListenAndServe(h.config.ServeAddr(), nil)
}
