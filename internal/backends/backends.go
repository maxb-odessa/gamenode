package backends

import (
	"fmt"

	"gamenode/internal/backends/file"
	"gamenode/internal/backends/joy"
	"gamenode/internal/backends/kbd"
	"gamenode/internal/backends/snd"

	pb "gamenode/pkg/gamenodepb"

	"github.com/maxb-odessa/sconf"
	"github.com/maxb-odessa/slog"
)

type backendHandler interface {
	Name() string
	Consumer() chan interface{}
	Producer() chan interface{}
}

// loaded and started backends
var loadedBackends map[pb.Backend_Type]map[string]backendHandler

type backendRunFunc func(string) (interface{}, error)

// registered backends that we support
var registeredBackends = map[pb.Backend_Type]backendRunFunc{
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

var backendsPbToTypeMap = map[pb.Backend_Type]string{
	pb.Backend_FILE: "file",
	pb.Backend_JOY:  "joy",
	pb.Backend_KBD:  "kbd",
	pb.Backend_SND:  "snd",
}

func Start() error {

	loadedBackends = make(map[pb.Backend_Type]map[string]backendHandler)

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
			loadedBackends[bkType] = make(map[string]backendHandler)
		}
		loadedBackends[bkType][bkName] = bkHandler.(backendHandler)

	}

	return nil
}

func GetProducer(bkType pb.Backend_Type) (chan interface{}, error) {

	bkTypeStr := backendsPbToTypeMap[bkType]

	retCh := make(chan interface{})

	lbk, ok := loadedBackends[bkType]
	if !ok {
		return nil, fmt.Errorf("no backends of type '%s' running", bkTypeStr)
	}

	for bkn, bkh := range lbk {

		slog.Debug(5, "connecting backend '%s' producer chan to general '%s' chan", bkn, bkTypeStr)
		go func(bkCh chan interface{}) {
			for {
				select {
				case data := <-bkCh:
					select {
					case retCh <- data:
					default:
						slog.Debug(5, "disconnecting backend '%s' producer chan from general '%s' chan", bkn, bkTypeStr)
						return
					}
				}
			}
		}(bkh.Producer())

	}

	return retCh, nil
}

func GetConsumer(bkType pb.Backend_Type) (chan interface{}, error) {

	retCh := make(chan interface{})

	return retCh, nil
}
