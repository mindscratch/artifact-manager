package artifacts

import "path"

// Volumes keeps a mapping of Marathon application volume hostPath's
// to Marathon application ID's.
type Volumes map[string][]string

// Add adds the Marathon appId and path to its internal store.
func (v Volumes) Add(appID, path string) {
	if path == "" {
		return
	}

	appIds, ok := v[path]
	if ok {
		// check if the appId already exists, if it does
		// then exit the function
		for _, val := range appIds {
			if val == appID {
				return
			}
		}
		appIds = append(appIds, appID)
	} else {
		appIds = []string{appID}
	}
	v[path] = appIds
}

// Get returns the Marathon application IDs that depend on a path
// identified by the given `path`.
//
// This function is safe to call without first checking if the name
// exists using the `Has` function.
func (v Volumes) Get(path string) []string {
	appIds, ok := v[path]
	if !ok {
		return make([]string, 0)
	}
	return appIds
}

// Has returns true if it contains the given path.
func (v Volumes) Has(path string) bool {
	_, ok := v[path]
	return ok
}

//TODO remove 'artifacts' type

// Artifacts represents a mapping of Marathon application volume hostPath's (name only)
// to Marathon application ID's.
type Artifacts map[string][]string

// Add adds the Marathon appId and artifact name (taken from the given uri) to its internal store.
func (a Artifacts) Add(appID, uri string) {
	name := path.Base(uri)
	if name == "" {
		return
	}

	appIds, ok := a[name]
	if ok {
		// check if the appId already exists, if it does
		// then exit the function
		for _, val := range appIds {
			if val == appID {
				return
			}
		}
		appIds = append(appIds, appID)
	} else {
		appIds = []string{appID}
	}
	a[name] = appIds
}

// Get returns the Marathon application IDs that depend on an artifacted
// identified by the given `name`.
//
// This function is safe to call without first checking if the name
// exists using the `Has` function.
func (a Artifacts) Get(name string) []string {
	appIds, ok := a[name]
	if !ok {
		return make([]string, 0)
	}
	return appIds
}

// Has returns true if it contains the given name.
func (a Artifacts) Has(name string) bool {
	_, ok := a[name]
	return ok
}
