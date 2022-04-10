package kbd

import (
	"fmt"

	"github.com/bendahl/uinput"

	pb "github.com/maxb-odessa/gamenode/pkg/gamenodepb"
)

type handler struct {
	confScope string
	vk        uinput.Keyboard
}

func newDev(confScope string) *handler {
	return &handler{}
}

func (h *handler) run() error {

	vk, err := uinput.CreateKeyboard("/dev/uinput", []byte("Gamenode Virtual Keyboard"))
	if err != nil {
		return err
	}

	h.vk = vk

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
	if key.Pressed {
		return h.vk.KeyDown(int(key.Code))
	} else {
		return h.vk.KeyUp(int(key.Code))
	}
}

func (h *handler) writeLed(key *pb.KbdEvent_Led) error {
	// TODO
	return nil
}
