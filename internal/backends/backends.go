package backends

import (
	"fmt"
	"gamenode/internal/backends/file"
	"gamenode/internal/backends/joy"

	"github.com/maxb-odessa/sconf"
)

type BackendHandler interface {
	Name() string
	Consumer() chan interface{}
	Producer() chan interface{}
}

type BackendType string
type BackendName string

// loaded and started backends
var loadedBackends map[BackendType]map[BackendName]BackendHandler

type BackendRunFunc func(string) (interface{}, error)

// registered backends that we support
var registeredBackends = map[BackendType]BackendRunFunc{
	"joy":  joy.Run,
	"file": file.Run,
}

func Start() error {

	loadedBackends = make(map[BackendType]map[BackendName]BackendHandler)

	// start all configured backends

	for _, confScope := range sconf.Scopes() {

		// which backend type to run
		bkt, err := sconf.ValAsStr(confScope, "backend")
		if err != nil {
			// it's ok, not all configured scopes are backends
			continue
		}
		bkType := BackendType(bkt)

		// is this backend exists/registered?
		bkRunFunc, ok := registeredBackends[bkType]
		if !ok {
			return fmt.Errorf("backend '%s' is not supported", bkType)
		}

		// get the name of a backend to run (mandatory)
		bkn, err := sconf.ValAsStr(confScope, "name")
		if err != nil {
			return fmt.Errorf("scope '%s' has no backend name set", confScope)
		}
		bkName := BackendName(bkn)

		// run selected backend and get its handler
		bkHandler, err := bkRunFunc(confScope)
		if err != nil {
			return fmt.Errorf("failed to run backend '%s': %s", bkName, err)
		}

		// save started backend handler
		if _, ok := loadedBackends[bkType]; !ok {
			loadedBackends[bkType] = make(map[BackendName]BackendHandler)
		}
		loadedBackends[bkType][bkName] = bkHandler.(BackendHandler)

	}

	return nil
}
