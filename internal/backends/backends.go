package backends

import (
	"fmt"

	"gamenode/internal/backends/file"
	"gamenode/internal/backends/joy"
	"gamenode/internal/backends/kbd"
	"gamenode/internal/backends/snd"
	"gamenode/internal/pubsub"

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

const (
	CONSUMER = 0x1000
	PRODUCER = 0x2000
)

var netPubsub *pubsub.Pubsub
var bkPubsub *pubsub.Pubsub

func NetSubscribe(bkt pb.Backend_Type) <-chan interface{} {
	return netPubsub.Subscribe(pubsub.Topic(bkt))
}

func NetPublish(bkt pb.Backend_Type, msg interface{}) {
	netPubsub.Publish(pubsub.Topic(bkt), msg)
}

func NetUnsubscribe(ch <-chan interface{}) {
	netPubsub.Unsubscribe(ch)
}

func BkSubscribe(bkt pb.Backend_Type) <-chan interface{} {
	return bkPubsub.Subscribe(pubsub.Topic(bkt))
}

func BkPublish(bkt pb.Backend_Type, msg interface{}) {
	bkPubsub.Publish(pubsub.Topic(bkt), msg)
}

func BkUnsubscribe(ch <-chan interface{}) {
	bkPubsub.Unsubscribe(ch)
}

func processNetToBkReq(bkType pb.Backend_Type, bkList map[string]backendHandler) {

	slog.Debug(9, "started processing net->bk reqs for '%s'", pb.Backend_Type_name[int32(bkType)])
	defer slog.Debug(9, "stopped processing net->bk reqs for '%s'", pb.Backend_Type_name[int32(bkType)])

	ch := NetSubscribe(bkType | PRODUCER)

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			slog.Debug(5, "got net msg: %+v", msg)
			// find running backend with NAME, get its Consumer and send msg into it
		}
	}
}

func Start() error {

	loadedBackends = make(map[pb.Backend_Type]map[string]backendHandler)

	// create pubsub object for network module comms
	netPubsub = pubsub.NewPubsub()

	// create pubsub object for backends comms
	bkPubsub = pubsub.NewPubsub()

	// start all configured backends
	for _, confScope := range sconf.Scopes() {

		// which backend type to run
		bkt, err := sconf.ValAsStr(confScope, "backend")
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

		// subscribe the backend to pubsub channel

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

	// do the same for backends, subscribe them

	// make chan for eath bktype, set this chan in each bk of type bktype as producer
	// select on all bktype chans, if has data - netPublish(bktype, msg)

	// run network "listeners"
	for bkType, bkList := range loadedBackends {
		go processNetReq(bkType, bkList)
	}

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
