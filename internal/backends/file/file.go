package file

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

type File struct {
	name   string
	broker *pubsub.Pubsub
	dev    device
}

func Init(confScope string) (interface{}, error) {
	var err error

	slog.Debug(9, "INIT file backend, scope '%s'", confScope)

	f := File{
		name: confScope,
		dev:  newDev(confScope),
	}

	// read config here, do preps
	if err = f.dev.run(); err != nil {
		return nil, err
	}

	// start reading file

	return f, nil
}

func (f File) Run(broker *pubsub.Pubsub) error {
	f.broker = broker

	go f.producer()
	go f.consumer()

	return nil
}

func (f File) producer() {

	for {

		// wait for file event to happen
		l, _ := f.dev.read()
		line := l.(string)
		slog.Debug(9, "file line: %s", line)

		// compose full PB message
		retMsg := &pb.FileMsg{
			Name: f.name,
			Msg: &pb.FileMsg_Event{
				Event: &pb.FileEvent{
					Obj: &pb.FileEvent_Line_{
						Line: &pb.FileEvent_Line{
							Line: line,
						},
					},
				},
			},
		}

		// send msg to network module
		f.broker.Publish(pubsub.Topic(pb.Backend_FILE|pubsub.PRODUCER), retMsg)

	} //for

}

func (f File) consumer() {

	// subscribe to file chan
	ch := f.broker.Subscribe(pubsub.Topic(pb.Backend_FILE | pubsub.CONSUMER))
	defer f.broker.Unsubscribe(ch)

	for {

		select {
		case m, ok := <-ch:
			if !ok {
				return
			}

			msg := m.(*pb.FileMsg)

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
				retMsg := &pb.FileMsg{
					Name: f.name,
					Msg: &pb.FileMsg_Error{
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

}
