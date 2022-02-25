package router

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gamenode/internal/backends"
	"gamenode/internal/network"
	"gamenode/internal/pubsub"
	pb "gamenode/pkg/gamenodepb"

	_ "gamenode/internal/backends/file"
	_ "gamenode/internal/backends/joy"
	_ "gamenode/internal/backends/kbd"
	_ "gamenode/internal/backends/snd"

	"github.com/maxb-odessa/slog"
)

func Start() error {

	// broker for network messages
	netBroker := pubsub.New()

	// broker for backends messages
	bksBroker := pubsub.New()

	// start backends
	if err := backends.Run(bksBroker); err != nil {
		return fmt.Errorf("backends start failed: %s", err)
	}

	// serve network requests
	if err := network.Run(netBroker); err != nil {
		return fmt.Errorf("network start failed: %s", err)
	}

	// process net<->backends messages
	return doRouting(bksBroker, netBroker)
}

func doRouting(bksBroker *pubsub.Pubsub, netBroker *pubsub.Pubsub) error {

	bkList := []pb.Backend_Type{
		pb.Backend_FILE,
		pb.Backend_JOY,
		pb.Backend_KBD,
		pb.Backend_SND,
	}

	for _, bkt := range bkList {

		// establish bridges net->backend for each backend type
		go func(bkType pb.Backend_Type) {
			ch := netBroker.Subscribe(pubsub.Topic(bkType | pubsub.PRODUCER))
			defer netBroker.Unsubscribe(ch)
			for {
				select {
				case msg, ok := <-ch:
					if !ok {
						return
					}
					bksBroker.Publish(pubsub.Topic(bkType|pubsub.CONSUMER), msg)
				}
			}
		}(bkt)

		// establish bridges backend->net for each backend type
		go func(bkType pb.Backend_Type) {
			ch := bksBroker.Subscribe(pubsub.Topic(bkType | pubsub.PRODUCER))
			defer bksBroker.Unsubscribe(ch)
			for {
				select {
				case msg, ok := <-ch:
					if !ok {
						return
					}
					netBroker.Publish(pubsub.Topic(bkType|pubsub.CONSUMER), msg)
				}
			}
		}(bkt)
	}

	// wait for a system signal
	done := make(chan bool)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		for sig := range sigChan {
			slog.Info("got signal %d\n", sig)
			done <- true
		}
	}()

	<-done

	// TODO: graceful exit, cleanups
	slog.Info("stopped")

	return nil
}
