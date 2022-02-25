package router

import (
	"fmt"

	"gamenode/internal/backends"
	"gamenode/internal/network"
	"gamenode/internal/pubsub"

	_ "gamenode/internal/backends/file"
	_ "gamenode/internal/backends/joy"
	_ "gamenode/internal/backends/kbd"
	_ "gamenode/internal/backends/snd"
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

	// subs to all nets (joy, file, etc) // pb.Backend_Type + pubsub.CONSUMER, PRODUCER

	// subs to all backs (joy, file, etc) // pb.Backend_Type + pubsub.CONSUMER, PRODUCER

	// go: select all net subsed chans
	// examine and publish to corresponding back chan (joy, file, etc)

	// go: select all back subsed chans
	// examine and publish to corresponding net chan (joy, file, etc)

	return nil
}
