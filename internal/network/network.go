package network

import (
	"io"
	"net"
	"sync"

	"github.com/maxb-odessa/sconf"
	"github.com/maxb-odessa/slog"

	"google.golang.org/grpc"

	"github.com/maxb-odessa/gamenode/internal/pubsub"
	pb "github.com/maxb-odessa/gamenode/pkg/gamenodepb"
)

type gameNodeServer struct {
	pb.UnimplementedGameNodeServer

	bkName string
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
		bkName: pb.Backend_Type_name[int32(pb.Backend_FILE)],
		bkType: pb.Backend_FILE,
		send: func(m interface{}) error {
			d := m.(*pb.FileMsg)
			return stream.Send(d)
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
		bkName: pb.Backend_Type_name[int32(pb.Backend_JOY)],
		bkType: pb.Backend_JOY,
		send: func(m interface{}) error {
			d := m.(*pb.JoyMsg)
			return stream.Send(d)
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
		bkName: pb.Backend_Type_name[int32(pb.Backend_KBD)],
		bkType: pb.Backend_KBD,
		send: func(m interface{}) error {
			d := m.(*pb.KbdMsg)
			return stream.Send(d)
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
		bkName: pb.Backend_Type_name[int32(pb.Backend_SND)],
		bkType: pb.Backend_SND,
		send: func(m interface{}) error {
			d := m.(*pb.SndMsg)
			return stream.Send(d)
		},
		recv: func() (interface{}, error) {
			return stream.Recv()
		},
	}

	return g.worker()
}

func (gns *gameNodeServer) worker() (err error) {

	// subscribe to backend channels
	consumerCh := broker.Subscribe(pubsub.Topic(gns.bkType | pubsub.CONSUMER))
	defer broker.Unsubscribe(consumerCh)

	slog.Debug(1, "client connected to stream '%s'", gns.bkName)
	defer slog.Debug(1, "client disconnected from stream '%s'", gns.bkName)

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
				slog.Err("stream '%s' error: %s", gns.bkName, err)
				break
			}

			slog.Debug(9, "stream '%s', read '%+v'", gns.bkName, inData)

			// send the data to backends
			broker.Publish(pubsub.Topic(gns.bkType|pubsub.PRODUCER), inData)

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

			slog.Debug(99, "stream '%s', sending %+v", gns.bkName, outData)

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

var broker *pubsub.Pubsub

func Run(brk *pubsub.Pubsub) error {

	addrPort := sconf.StrDef("network", "listen", "127.0.0.1:12346")

	lstn, err := net.Listen("tcp", addrPort)
	if err != nil {
		return err
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterGameNodeServer(grpcServer, newServer())

	broker = brk

	slog.Info("serving gRPC at %s", addrPort)
	go grpcServer.Serve(lstn)

	return nil
}
