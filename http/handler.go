package http

import (
	"fmt"
	"log"
	gohttp "net/http"
	"path"

	"apex/artifact-manager/core"
)

// Handler represents a type that handles HTTP requests.
type Handler struct {
	// application configuraiton
	config *core.Config
	// logger for logging debug messages
	debug *log.Logger
	// the location of the uploaded file (symlink destination or actual file)
	// will be placed onto the channel
	requestQueue chan<- string
	// the max size of the request queue
	maxQueueSize int
}

// NewHandler creates a new Handler.
func NewHandler(config *core.Config, requestQueue chan<- string, maxQueueSize int, debug *log.Logger) *Handler {
	h := Handler{
		config:       config,
		debug:        debug,
		requestQueue: requestQueue,
		maxQueueSize: maxQueueSize,
	}
	return &h
}

// UploadHandler handles file upload requests
func (h *Handler) UploadHandler(w gohttp.ResponseWriter, r *gohttp.Request) {
	core.Log("received %s request to %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	// only accept POST
	if r.Method != gohttp.MethodPost {
		w.WriteHeader(gohttp.StatusMethodNotAllowed)
		fmt.Fprintf(w, "only %s is allowed", gohttp.MethodPost)
		return
	}
	// check to ensure content is being provided
	if r.ContentLength <= 0 {
		core.Log("invalid request, no content provided")
		w.WriteHeader(gohttp.StatusBadRequest)
		fmt.Fprintf(w, "invalid request, no content provided")
		return
	}
	// check URL parameters
	queryParams := r.URL.Query()
	name := queryParams.Get("name")
	src := queryParams.Get("src")
	dst := queryParams.Get("dst")
	if name == "" {
		core.Log("invalid request, name parameter must be provided in the URL")
		w.WriteHeader(gohttp.StatusBadRequest)
		fmt.Fprintf(w, "name parameter must be provided in the URL")
		return
	}

	// check if the queue is full, if so reject the request
	if len(h.requestQueue) >= h.maxQueueSize {
		core.Log("server has too many requests (%d) to fulfill", len(h.requestQueue))
		w.WriteHeader(gohttp.StatusServiceUnavailable)
		fmt.Fprintf(w, "server has too many requests (%d) to fulfill", len(h.requestQueue))
		return
	}

	defer r.Body.Close()

	// in case the name included a path, ensure we jut have the name of the file, and
	// add the directory to it
	name = path.Base(name)
	name = path.Join(h.config.Dir, name)

	// src and dst are optional, if they're provided a symlink we'll be created
	var err error
	createSymlink := false
	if src != "" && dst != "" {
		createSymlink = true
		src = path.Join(h.config.Dir, src)
		dst = path.Join(h.config.Dir, dst)
	}

	// save the file
	err = core.SaveFile(name, r.Body, r.ContentLength)
	if err != nil {
		core.Log("problem saving file to %s: %v", name, err)
		w.WriteHeader(gohttp.StatusInternalServerError)
		fmt.Fprintf(w, "problem saving file to %s: %v\n", name, err)
		return
	}

	// the message to put onto the 'requestQueue' is the name of file
	// or the "dst" of the symlink (as the 'dst' woud match the 'hostPath' of the marathon app)
	requestMsg := name
	if createSymlink {
		requestMsg = dst

		// if the 'src' already exists and is not the same as 'name', move it
		if src != name {
			h.debug.Printf("Given src %s might exist, renaming if necessary", src)
			err = core.RenameWithTimestamp(src)
			if err != nil {
				core.Log("problem renaming existing source path %s: %v", src, err)
				w.WriteHeader(gohttp.StatusInternalServerError)
				fmt.Fprintf(w, "problem renaming existing source path %s: %v", src, err)
				return
			}
		}

		// extract the file (if its an archive, otherwise this won't do anything)
		h.debug.Printf("Extracting %s (if it's an archive) into %s", name, h.config.Dir)
		err = core.ExtractFile(name, h.config.Dir)
		if err != nil {
			core.Log("problem extracting file %s into %s: %v\n", name, h.config.Dir, err)
			w.WriteHeader(gohttp.StatusInternalServerError)
			fmt.Fprintf(w, "problem extracting file %s into %s: %v\n", name, h.config.Dir, err)
			return
		}

		// create symlink
		h.debug.Printf("Creating symlink from %s to %s", src, dst)
		err = core.Symlink(src, dst)
		if err != nil {
			core.Log("problem creating symlink from %s to %s: %v", src, dst, err)
			w.WriteHeader(gohttp.StatusInternalServerError)
			fmt.Fprintf(w, "problem creating symlink from %s to %s: %v", src, dst, err)
			return
		}
	}

	// check if the queue is full, if so reject the request
	if len(h.requestQueue) >= h.maxQueueSize {
		w.WriteHeader(gohttp.StatusServiceUnavailable)
		fmt.Fprintf(w, "server has too many requests (%d) to fulfill", len(h.requestQueue))
		return
	}

	h.debug.Printf("Adding %s to request queue", requestMsg)
	h.requestQueue <- requestMsg

	w.WriteHeader(gohttp.StatusCreated)
}

// ListenAndServe starts the server
func (h *Handler) ListenAndServe() error {
	gohttp.HandleFunc("/", h.UploadHandler)

	return gohttp.ListenAndServe(h.config.ServeAddr(), nil)
}
