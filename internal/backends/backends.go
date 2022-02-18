package backends

import (
	"fmt"
	"sync"

	"gamenode/internal/backends/file"
	"gamenode/internal/backends/joy"
	"gamenode/internal/backends/kbd"
	"gamenode/internal/backends/snd"

	pb "gamenode/pkg/gamenodepb"

	"github.com/maxb-odessa/sconf"
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

	// create pubsub object for network module comms
	netPubsub = NewPubsub()

	// do the same for backends, subscribe them

	/*
		go func() {
			ch := netPubsub.Subscribe(2)

			// make bk-net broker, send only matched msgs (NET chan)
			// bks must subs to another chan (BK chan)
			for {

				select {
				case d := <-ch:
					slog.Info("BR: got '%+v'", d)
				default:
				}

				jmsg := pb.JoyMsg{
					Name: "x52pro",
					Msg: &pb.JoyMsg_Event{
						Event: &pb.JoyEvent{
							Obj: &pb.JoyEvent_Button_{
								Button: &pb.JoyEvent_Button{
									Pressed: false,
									Color:   "BLUE"},
							},
						},
					},
				}

				PS.Publish(1, jmsg)
				//slog.Info("BR: pub data")

				time.Sleep(time.Second * 1)

			}
		}()
	*/

	return nil
}

const (
	NET_CONSUMER  = 1
	NET_PUBLISHER = 2
)

var netPubsub *Pubsub

func GetNetPubsub() *Pubsub {
	return netPubsub
}

// https://eli.thegreenplace.net/2020/pubsub-using-channels-in-go/

type Pubsub struct {
	sync.RWMutex
	subs   map[pb.Backend_Type][]chan interface{}
	closed bool
}

func NewPubsub() *Pubsub {
	ps := &Pubsub{}
	ps.subs = make(map[pb.Backend_Type][]chan interface{})
	return ps
}

func (ps *Pubsub) Subscribe(topic pb.Backend_Type) <-chan interface{} {
	ps.Lock()
	defer ps.Unlock()

	ch := make(chan interface{}, 8)
	ps.subs[topic] = append(ps.subs[topic], ch)
	return ch
}

func (ps *Pubsub) Publish(topic pb.Backend_Type, msg interface{}) {
	ps.RLock()
	defer ps.RUnlock()

	if ps.closed {
		return
	}

	for _, ch := range ps.subs[topic] {
		go func(ch chan interface{}) {
			ch <- msg
		}(ch)
	}
}

func (ps *Pubsub) Unsubscribe() {
	ps.Lock()
	defer ps.Unlock()

	if !ps.closed {
		ps.closed = true
		for _, subs := range ps.subs {
			for _, ch := range subs {
				close(ch)
			}
		}
	}
}

/*
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
*/
