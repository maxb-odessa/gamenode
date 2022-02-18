package network

import (
	"io"
	"net"
	"sync"

	"github.com/maxb-odessa/sconf"
	"github.com/maxb-odessa/slog"
	"google.golang.org/grpc"

	"gamenode/internal/backends"
	pb "gamenode/pkg/gamenodepb"
)

type gameNodeServer struct {
	pb.UnimplementedGameNodeServer

	name   string
	bkType pb.Backend_Type
	send   func(interface{}) error
	recv   func() (interface{}, error)
}

func newServer() *gameNodeServer {
	return new(gameNodeServer)
}

func (gns *gameNodeServer) File(stream pb.GameNode_FileServer) error {

	// make new obj for each new stream
	g := &gameNodeServer{
		name:   "FileStream",
		bkType: pb.Backend_FILE,
		send: func(m interface{}) error {
			d := m.(pb.FileMsg)
			return stream.Send(&d)
		},
		recv: func() (interface{}, error) {
			return stream.Recv()
		},
	}

	return g.worker()
}

func (gns *gameNodeServer) Joy(stream pb.GameNode_JoyServer) error {

	// make new obj for each new stream
	g := &gameNodeServer{
		name:   "JoyStream",
		bkType: pb.Backend_JOY,
		send: func(m interface{}) error {
			d := m.(pb.JoyMsg)
			return stream.Send(&d)
		},
		recv: func() (interface{}, error) {
			return stream.Recv()
		},
	}

	return g.worker()

}

func (gns *gameNodeServer) Kbd(stream pb.GameNode_KbdServer) error {

	// make new obj for each new stream
	g := &gameNodeServer{
		name:   "KbdStream",
		bkType: pb.Backend_KBD,
		send: func(m interface{}) error {
			d := m.(pb.KbdMsg)
			return stream.Send(&d)
		},
		recv: func() (interface{}, error) {
			return stream.Recv()
		},
	}

	return g.worker()
}

func (gns *gameNodeServer) Snd(stream pb.GameNode_SndServer) error {

	// make new obj for each new stream
	g := &gameNodeServer{
		name:   "SndStream",
		bkType: pb.Backend_SND,
		send: func(m interface{}) error {
			d := m.(pb.SndMsg)
			return stream.Send(&d)
		},
		recv: func() (interface{}, error) {
			return stream.Recv()
		},
	}

	return g.worker()
}

func (gns *gameNodeServer) worker() (err error) {

	// subscribe to backend channels

	pubsub := backends.GetNetPubsub()
	defer pubsub.Unsubscribe()

	consumerCh := pubsub.Subscribe(gns.bkType | backends.NET_CONSUMER)

	slog.Debug(1, "client connected to stream '%s'", gns.name)
	defer slog.Debug(1, "client disconnected from stream '%s'", gns.name)

	done := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(2)

	// start stream receiver, send data to backends
	go func() {
		defer wg.Done()

		// tell the sender we're done
		defer func() { done <- true }()

		for {

			inData, err := gns.recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				slog.Err("stream '%s' error: %s", gns.name, err)
				break
			}

			slog.Debug(9, "stream '%s', read '%+v'", gns.name, inData)

			// send the data to backends
			pubsub.Publish(gns.bkType|backends.NET_PUBLISHER, inData)

		} // for...

	}()

	// start stream sender, read data drom backend producer
	go func() {
		defer wg.Done()

		var outData interface{}

		for {
			// wait for data from fromBackend chan
			select {
			case <-done: // the reader is done (stream closed, etc)
				return
			case outData = <-consumerCh:
			}

			slog.Debug(99, "stream '%s', sending %+v", gns.name, outData)

			err := gns.send(outData)
			if err != nil {
				slog.Err("stream '%s', failed to send: %s", err)
				break
			}
		}
	}()

	wg.Wait()

	return nil
}

func Start() error {

	addrPort := "127.0.0.1:12346" // the default
	if ap, err := sconf.ValAsStr("network", "listen"); err == nil {
		addrPort = ap
	}

	lstn, err := net.Listen("tcp", addrPort)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterGameNodeServer(grpcServer, newServer())

	slog.Info("accepting gRPC at %s", addrPort)
	grpcServer.Serve(lstn)

	return nil
}
