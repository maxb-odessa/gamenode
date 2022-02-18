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

	prodCh chan interface{}
	consCh chan interface{}

	name   string
	bkType pb.Backend_Type
	err    func(interface{}) error
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
		send: func(data interface{}) error {
			d := data.(pb.FileEvent)
			m := &pb.FileMsg{
				Msg: &pb.FileMsg_Event{Event: &d},
			}
			return stream.Send(m)
		},
		err: func(data interface{}) error {
			d := data.(pb.Error)
			m := &pb.FileMsg{
				Msg: &pb.FileMsg_Error{Error: &d},
			}
			return stream.Send(m)
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
		send: func(data interface{}) error {
			d := data.(pb.JoyEvent)
			m := &pb.JoyMsg{
				Msg: &pb.JoyMsg_Event{Event: &d},
			}
			return stream.Send(m)
		},
		err: func(data interface{}) error {
			d := data.(pb.Error)
			m := &pb.JoyMsg{
				Msg: &pb.JoyMsg_Error{Error: &d},
			}
			return stream.Send(m)
		},
		recv: func() (interface{}, error) {
			return stream.Recv()
		},
	}
	// /* TEST
	jmsg := &pb.JoyMsg{
		Name: "x52pro",
		Msg: &pb.JoyMsg_Event{
			Event: &pb.JoyEvent{
				Obj: &pb.JoyEvent_Button_{
					Button: &pb.JoyEvent_Button{
						Pressed: false,
						Color:   "BLUE"},
				},
			},
		},
	}
	stream.Send(jmsg)
	//*/

	return g.worker()

}

func (gns *gameNodeServer) Kbd(stream pb.GameNode_KbdServer) error {

	// make new obj for each new stream
	g := &gameNodeServer{
		name:   "KbdStream",
		bkType: pb.Backend_KBD,
		send: func(data interface{}) error {
			d := data.(pb.KbdEvent)
			m := &pb.KbdMsg{
				Msg: &pb.KbdMsg_Event{Event: &d},
			}
			return stream.Send(m)
		},
		err: func(data interface{}) error {
			d := data.(pb.Error)
			m := &pb.KbdMsg{
				Msg: &pb.KbdMsg_Error{Error: &d},
			}
			return stream.Send(m)
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
		send: func(data interface{}) error {
			d := data.(pb.SndEvent)
			m := &pb.SndMsg{
				Msg: &pb.SndMsg_Event{Event: &d},
			}
			return stream.Send(m)
		},
		err: func(data interface{}) error {
			d := data.(pb.Error)
			m := &pb.SndMsg{
				Msg: &pb.SndMsg_Error{Error: &d},
			}
			return stream.Send(m)
		},
		recv: func() (interface{}, error) {
			return stream.Recv()
		},
	}

	return g.worker()
}

func (gns *gameNodeServer) getProdConsCh() (err error) {

	gns.prodCh, err = backends.GetProducer(gns.bkType)
	if err != nil {
		return
	}

	gns.consCh, err = backends.GetConsumer(gns.bkType)
	if err != nil {
		close(gns.prodCh)
		return
	}

	return
}

func (gns *gameNodeServer) sendError(err error) {
	e := pb.Error{
		Code: 1,
		Desc: err.Error(),
	}
	gns.err(e)
}

func (gns *gameNodeServer) worker() (err error) {

	// get backend producer and consumer channels
	if err = gns.getProdConsCh(); err != nil {
		slog.Err("can not serve grpc stream '%s': %s", gns.name, err)
		gns.sendError(err)
		return
	}

	slog.Debug(1, "client connected to stream '%s'", gns.name)
	defer slog.Debug(1, "client disconnected from stream '%s'", gns.name)

	done := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(2)

	// start stream receiver, send data to backend consumer
	go func() {
		defer wg.Done()
		defer close(gns.consCh)

		// tell the sender we're done
		defer func() { done <- true }()

		for {

			inData, err := gns.recv()

			if err == io.EOF {
				slog.Info("client disconnected from '%s' stream", gns.name)
				break
			}

			if err != nil {
				slog.Err("stream '%s' error: %s", gns.name, err)
				break
			}

			slog.Debug(9, "stream '%s', read '%+v'", gns.name, inData)

			select {
			case gns.consCh <- inData:
			default:
				return
			}

		} // for...

	}()

	// start stream sender, read data drom backend producer
	go func() {
		defer wg.Done()
		defer close(gns.prodCh)

		for {
			var outData interface{}

			// wait for data from fromBackend chan
			select {
			case outData = <-gns.prodCh:
			case <-done: // the reader is done (stream closed, etc)
				return
			}

			slog.Debug(9, "stream '%s', sending %+v", gns.name, outData)

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
