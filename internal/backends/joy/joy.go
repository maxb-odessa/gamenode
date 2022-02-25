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

	return j, nil
}

func (j Joy) Run(broker *pubsub.Pubsub) error {
	j.broker = broker

	go j.producer()
	go j.consumer()

	return nil
}

func (j Joy) producer() {

	// start device reader here

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

	// start device writer here

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
