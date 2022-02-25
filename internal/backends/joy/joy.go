package joy

import (
	"time"

	"gamenode/internal/pubsub"
	pb "gamenode/pkg/gamenodepb"

	"github.com/maxb-odessa/slog"
)

type Joy struct {
	name string
}

func Init(confScope string) (interface{}, error) {
	slog.Debug(9, "joy INIT %s", confScope)
	return Joy{name: confScope}, nil
}

func (j Joy) Run(broker *pubsub.Pubsub) error {
	go j.producer(broker)
	go j.consumer(broker)
	return nil
}

func (j Joy) producer(broker *pubsub.Pubsub) {

	for {
		slog.Debug(9, "BK %s, producing", j.name)
		time.Sleep(time.Second * 2)
	}

}

func (j Joy) consumer(broker *pubsub.Pubsub) {
	ch := broker.Subscribe(pubsub.Topic(pb.Backend_JOY + pubsub.CONSUMER))
	defer broker.Unsubscribe(ch)
	for {
		slog.Debug(9, "BK %s, consuming", j.name)
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			slog.Debug(9, "BK %s got msg: %+v", j.name, msg)
		}
	}
}
