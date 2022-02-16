package backends

import (
	"fmt"

	"gamenode/internal/backends/file"
	"gamenode/internal/backends/joy"
	"gamenode/internal/backends/kbd"
	"gamenode/internal/backends/snd"

	pb "gamenode/pkg/gamenodepb"

	"github.com/maxb-odessa/sconf"
)

type BackendHandler interface {
	Name() string
	Consumer() chan interface{}
	Producer() chan interface{}
}

// loaded and started backends
var loadedBackends map[pb.Backend_Type]map[string]BackendHandler

type BackendRunFunc func(string) (interface{}, error)

// registered backends that we support
var registeredBackends = map[pb.Backend_Type]BackendRunFunc{
	pb.Backend_FILE: file.Run,
	pb.Backend_JOY:  joy.Run,
	pb.Backend_KBD:  kbd.Run,
	pb.Backend_SND:  snd.Run,
}

// map configured backend name to protobuf const
var backendsTypeToPbMap = map[string]pb.Backend_Type{
	"file": pb.Backend_FILE,
	"joy":  pb.Backend_JOY,
	"kbd":  pb.Backend_KBD,
	"snd":  pb.Backend_SND,
}

func Start() error {

	loadedBackends = make(map[pb.Backend_Type]map[string]BackendHandler)

	// start all configured backends

	for _, confScope := range sconf.Scopes() {

		// which backend type to run
		bkt, err := sconf.ValAsStr(confScope, "backend")
		if err != nil {
			// it's ok, not all configured scopes are backends
			continue
		}

		// is this backend exists/registered?
		bkType, ok := backendsTypeToPbMap[bkt]
		if !ok {
			return fmt.Errorf("backend type '%s' is not supported", bkt)
		}

		bkName := confScope

		// run selected backend and get its handler
		bkRunFunc := registeredBackends[bkType]
		bkHandler, err := bkRunFunc(confScope)
		if err != nil {
			return fmt.Errorf("failed to run backend '%s': %s", bkName, err)
		}

		// save started backend handler
		if _, ok := loadedBackends[bkType]; !ok {
			loadedBackends[bkType] = make(map[string]BackendHandler)
		}
		loadedBackends[bkType][bkName] = bkHandler.(BackendHandler)

	}

	// make a chan for each pb backend type
	// ...

	return nil
}
