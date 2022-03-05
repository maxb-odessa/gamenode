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
	seqNo  int32
}

func Init(confScope string) (interface{}, error) {
	var err error

	slog.Debug(9, "INIT kbd backend, scope '%s'", confScope)

	f := Kbd{
		name: confScope,
		dev:  newDev(confScope),
	}

	// read config here, do preps
	if err = f.dev.run(); err != nil {
		return nil, err
	}

	// start reading kbd

	return f, nil
}

func (f Kbd) Run(broker *pubsub.Pubsub) error {
	f.broker = broker

	go f.producer()
	go f.consumer()

	return nil
}

func (f Kbd) producer() {

	for {

		// TODO: not implemented atm
		f.dev.read()

	} //for

	return
}

func (f Kbd) consumer() {

	// subscribe to kbd chan
	ch := f.broker.Subscribe(pubsub.Topic(pb.Backend_FILE | pubsub.CONSUMER))
	defer f.broker.Unsubscribe(ch)

	for {

		select {
		case m, ok := <-ch:
			if !ok {
				return
			}

			msg := m.(*pb.KbdMsg)

			// discard msgs not ment for us
			if f.name != msg.GetName() {
				continue
			}

			ev := msg.GetEvent()

			// ignore non-event msgs
			if ev == nil {
				continue
			}

			// send event line to device
			if err := f.dev.write(ev.GetObj()); err != nil {

				// failed: compose an error msg and publish it back to router
				retMsg := pb.KbdMsg{
					Name:  f.name,
					SeqNo: msg.GetSeqNo(),
					Msg: &pb.KbdMsg_Error{
						Error: &pb.Error{
							Code: 1, // TODO: mnemonic error codes
							Desc: err.Error(),
						},
					},
				}
				f.broker.Publish(pubsub.Topic(pb.Backend_FILE|pubsub.PRODUCER), retMsg)

			}

		} // select

	} //for

	return
}
