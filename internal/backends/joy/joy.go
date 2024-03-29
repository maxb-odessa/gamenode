package joy

import (
	"gamenode/internal/pubsub"
	pb "gamenode/pkg/gamenodepb"

	"github.com/maxb-odessa/slog"
)

type Joy struct {
	name    string
	broker  *pubsub.Pubsub
	readCh  chan interface{}
	writeCh chan interface{}
}

func Init(confScope string) (interface{}, error) {

	slog.Debug(9, "joy INIT %s", confScope)

	j := Joy{
		name:    confScope,
		readCh:  make(chan interface{}),
		writeCh: make(chan interface{}),
	}

	// read config here, do preps

	// connect to joy device

	return j, nil
}

func (j Joy) Run(broker *pubsub.Pubsub) error {
	j.broker = broker

	if err := j.writer(); err != nil {
		return err
	}

	if err := j.reader(); err != nil {
		return err
	}

	go j.producer()
	go j.consumer()

	return nil
}

func (j Joy) reader() error {
	// connect to joy

	// start reading joy device

	// push joy events into readCh

	return nil
}

func (j Joy) writer() error {

	// start reading joy events to apply

	// write event to joy device

	return nil
}

func (j Joy) producer() {

	for {
		// wait for joystick event to happen
		select {
		case msg, ok := <-j.readCh:
			if !ok {
				return
			}
			j.broker.Publish(pubsub.Topic(pb.Backend_JOY|pubsub.PRODUCER), msg)
		}
	}

	return
}

func (j Joy) consumer() {

	// subscribe to joy chan
	ch := j.broker.Subscribe(pubsub.Topic(pb.Backend_JOY | pubsub.CONSUMER))
	defer j.broker.Unsubscribe(ch)

	for {

		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}

			slog.Debug(9, "BK '%s' got msg: %+v", j.name, msg)

			jmsg := msg.(*pb.JoyMsg)

			// discard msgs not ment for us
			if j.name != jmsg.GetName() {
				slog.Debug(9, "BK '%s' discarding msg, not for us", j.name)
				continue
			}

			slog.Debug(9, "BK '%s', consuming msg %+v", j.name, jmsg)

			// send data to device writer
			j.writeCh <- jmsg

		} // select

	} //for

	return
}
