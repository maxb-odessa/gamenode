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
	seqNo  int32
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

func (f File) producer() {

	for {

		// wait for file event to happen
		s, _ := f.dev.read()
		slog.Debug(99, "file line: '%s'", s)

		// make a PB message from the string
		f.seqNo++
		retMsg := pb.FileMsg{
			Name:  f.name,
			SeqNo: f.seqNo,
			Msg: &pb.FileMsg_Event{
				Event: &pb.FileEvent{
					Line: s,
				},
			},
		}

		// send msg to network module
		f.broker.Publish(pubsub.Topic(pb.Backend_FILE|pubsub.PRODUCER), retMsg)

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

			ev := fmsg.GetEvent()

			// ignore non-event msgs
			if ev == nil {
				continue
			}

			// send event line to device
			if err := f.dev.write(ev.GetLine()); err != nil {

				// failed: compose an error msg and publish it back to router
				retMsg := pb.FileMsg{
					Name:  f.name,
					SeqNo: fmsg.GetSeqNo(),
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

	return
}
