package http

import (
	"fmt"
	gohttp "net/http"
	"path"

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
	// check URL parameters
	queryParams := r.URL.Query()
	name := queryParams.Get("name")
	src := queryParams.Get("src")
	dst := queryParams.Get("dst")
	if name == "" {
		w.WriteHeader(gohttp.StatusBadRequest)
		fmt.Fprintf(w, "name parameter must be provided in the URL")
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
		w.WriteHeader(gohttp.StatusInternalServerError)
		fmt.Fprintf(w, "problem saving file to %s: %v\n", name, err)
		return
	}

	if createSymlink {
		// extract the file
		err = core.ExtractFile(name, h.config.Dir)
		if err != nil {
			w.WriteHeader(gohttp.StatusInternalServerError)
			fmt.Fprintf(w, "problem extracting file %s into %s: %v\n", name, h.config.Dir, err)
			return
		}

		// create symlink
		err = core.Symlink(src, dst)
		if err != nil {
			w.WriteHeader(gohttp.StatusInternalServerError)
			fmt.Fprintf(w, "problem creating symlink from %s to %s: %v", src, dst, err)
			return
		}
	}

	w.WriteHeader(gohttp.StatusCreated)
}

// ListenAndServe starts the server
func (h *Handler) ListenAndServe() error {
	gohttp.HandleFunc("/", h.UploadHandler)

	return gohttp.ListenAndServe(h.config.ServeAddr(), nil)
}
