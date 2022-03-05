package kbd

import (
	"fmt"

	pb "github.com/maxb-odessa/gamenode/pkg/gamenodepb"
)

type handler struct {
	confScope string
}

func newDev(confScope string) *handler {
	return &handler{}
}

func (h *handler) run() error {
	return nil
}

func (h *handler) read() (interface{}, error) {
	// not reading anything atm
	ch := make(chan bool)
	<-ch
	return nil, nil
}

func (h *handler) write(i interface{}) error {

	o := i.(*pb.KbdEvent)

	if key := o.GetKey(); key != nil {
		return h.writeKey(key)
	} else if led := o.GetLed(); led != nil {
		return h.writeLed(led)
	}

	return fmt.Errorf("unknown object")
}

func (h *handler) writeKey(key *pb.KbdEvent_Key) error {
	return nil
}

func (h *handler) writeLed(key *pb.KbdEvent_Led) error {
	return nil
}
