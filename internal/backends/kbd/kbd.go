package kbd

import (
	"github.com/maxb-odessa/gamenode/internal/pubsub"
	pb "github.com/maxb-odessa/gamenode/pkg/gamenodepb"

	"github.com/maxb-odessa/slog"
)

type device interface {
	run() error
	read() (interface{}, error)
	write(interface{}) error
}

type Kbd struct {
	name   string
	broker *pubsub.Pubsub
	dev    device
}

func Init(confScope string) (interface{}, error) {
	var err error

	slog.Debug(9, "INIT kbd backend, scope '%s'", confScope)

	k := Kbd{
		name: confScope,
		dev:  newDev(confScope),
	}

	// read config here, do preps
	if err = k.dev.run(); err != nil {
		return nil, err
	}

	// start reading kbd

	return k, nil
}

func (k Kbd) Run(broker *pubsub.Pubsub) error {
	k.broker = broker

	go k.producer()
	go k.consumer()

	return nil
}

func (k Kbd) producer() {

	for {

		// TODO: not implemented atm
		k.dev.read()

	} //for

	return
}

func (k Kbd) consumer() {

	// subscribe to kbd chan
	ch := k.broker.Subscribe(pubsub.Topic(pb.Backend_KBD | pubsub.CONSUMER))
	defer k.broker.Unsubscribe(ch)

	for {

		select {
		case m, ok := <-ch:
			if !ok {
				return
			}

			msg := m.(*pb.KbdMsg)

			// discard msgs not ment for us
			if k.name != msg.GetName() {
				slog.Debug(5, "KBD: msg named '%s' is not for '%s'", msg.GetName(), k.name)
				continue
			}

			ev := msg.GetEvent()

			// ignore non-event msgs
			if ev == nil {
				continue
			}

			// send event line to device
			if err := k.dev.write(ev); err != nil {

				// failed: compose an error msg and publish it back to router
				retMsg := pb.KbdMsg{
					Name: k.name,
					Msg: &pb.KbdMsg_Error{
						Error: &pb.Error{
							Code: 1, // TODO: mnemonic error codes
							Desc: err.Error(),
						},
					},
				}
				k.broker.Publish(pubsub.Topic(pb.Backend_KBD|pubsub.PRODUCER), retMsg)

			}

		} // select

	} //for

	return
}
