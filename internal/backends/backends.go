package backends

import (
	"fmt"

	"gamenode/internal/backends/file"
	"gamenode/internal/backends/kbd"

	"gamenode/internal/pubsub"
	pb "gamenode/pkg/gamenodepb"

	"github.com/maxb-odessa/sconf"
)

type backendHandler interface {
	Run(*pubsub.Pubsub) error
}

// loaded and started backends
var loadedBackends map[pb.Backend_Type]map[string]backendHandler

type backendInitFunc func(string) (interface{}, error)

// registered backends that we support
var registeredBackends = map[pb.Backend_Type]backendInitFunc{
	pb.Backend_FILE: file.Init,
	//pb.Backend_JOY:  joy.Init,
	pb.Backend_KBD: kbd.Init,
	//pb.Backend_SND:  snd.Init,
}

func Run(broker *pubsub.Pubsub) error {

	loadedBackends = make(map[pb.Backend_Type]map[string]backendHandler)

	// start all configured backends
	for _, confScope := range sconf.Scopes() {

		// which backend type to run
		bkt, err := sconf.Str(confScope, "backend")
		if err != nil {
			// it's ok, not all configured scopes are backends
			continue
		}

		// is this backend exists/registered?
		var bkType pb.Backend_Type
		if b, ok := pb.Backend_Type_value[bkt]; !ok {
			return fmt.Errorf("backend type '%s' is not supported", bkt)
		} else {
			bkType = pb.Backend_Type(b)
		}

		bkName := confScope

		// init selected backend and get its handler
		bkInitFunc, ok := registeredBackends[bkType]
		if !ok {
			return fmt.Errorf("backend type '%s' is not compiled", bkType)
		}

		bkHandler, err := bkInitFunc(confScope)
		if err != nil {
			return fmt.Errorf("failed to init backend '%s': %s", bkName, err)
		}

		// save started backend handler
		if _, ok := loadedBackends[bkType]; !ok {
			loadedBackends[bkType] = make(map[string]backendHandler)
		}
		loadedBackends[bkType][bkName] = bkHandler.(backendHandler)

	}

	// run all loaded backends
	for bkt, _ := range loadedBackends {
		for bkn, _ := range loadedBackends[bkt] {
			if err := loadedBackends[bkt][bkn].Run(broker); err != nil {
				return fmt.Errorf("failed to run backend '%s': %s", bkn, err)
			}
		}
	}

	return nil
}
