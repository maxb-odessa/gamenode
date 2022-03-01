package file

import (
	"gamenode/internal/pubsub"
	pb "gamenode/pkg/gamenodepb"

	"github.com/maxb-odessa/slog"
)

type device interface {
	run() error
	read() (string, error)
	write(string) error
}

type File struct {
	name   string
	broker *pubsub.Pubsub
	dev    device
}

func Init(confScope string) (interface{}, error) {
	var err error

	slog.Debug(9, "file INIT %s", confScope)

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

func (f File) reader() error {
	// connect to file

	// start reading file device

	// push file events into readCh

	return nil
}

func (f File) writer() error {

	// not implemented: make an error and send it back

	return nil
}

func (f File) producer() {

	for {
		// wait for file event to happen
		s, _ := f.dev.read()
		slog.Debug(9, "file line: '%s'", s)
		// make a PB message from the string
		msg := pb.JoyMsg{}
		// send msg to network module
		f.broker.Publish(pubsub.Topic(pb.Backend_FILE|pubsub.PRODUCER), msg)
	} //for

	return
}

func (f File) consumer() {

	// subscribe to file chan
	ch := f.broker.Subscribe(pubsub.Topic(pb.Backend_FILE | pubsub.CONSUMER))
	defer f.broker.Unsubscribe(ch)

	for {

		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}

			fmsg := msg.(*pb.FileMsg)

			// discard msgs not ment for us
			if f.name != fmsg.GetName() {
				continue
			}

			// get data from PB message
			// s := ...
			// send data to device writer
			f.dev.write("")

		} // select

	} //for

	return
}
