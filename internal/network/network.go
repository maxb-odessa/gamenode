package network

import (
	"io"
	"net"
	"time"

	"github.com/maxb-odessa/sconf"
	"github.com/maxb-odessa/slog"
	"google.golang.org/grpc"

	pb "gamenode/pkg/gamenodepb"
)

type gameNodeServer struct {
	pb.UnimplementedGameNodeServer

	name string
	send func(interface{}) error
	recv func() (interface{}, error)
}

func (gns *gameNodeServer) Joy(stream pb.GameNode_JoyServer) error {
	g := new(gameNodeServer)
	/*
	   	g.backs = backends.Find("File")
	   foreach backs
	   ctx?
	*/
	send := func(data interface{}) error { d := data.(pb.JoyEvent); return stream.Send(&d) }
	recv := func() (interface{}, error) { return stream.Recv() }
	g.send = send
	g.recv = recv
	g.name = "Joy"

	return g.worker()
	/*
	   done

	   wait for all to finish
	*/
}

func (gns *gameNodeServer) File(stream pb.GameNode_FileServer) error {
	// get File backend inch and outch
	// gns.r, err := router.New("File"); if err != nil ...
	// gns.Register()
	// defer gns.Unregister()

	send := func(data interface{}) error { d := data.(pb.FileEvent); return stream.Send(&d) }
	recv := func() (interface{}, error) { return stream.Recv() }

	gns1 := new(gameNodeServer)
	gns1.send = send
	gns1.recv = recv
	gns1.name = "File"

	return gns1.worker()
}

func (gns *gameNodeServer) worker() error {

	go func() {
		for {
			in, err := gns.recv()
			if err == io.EOF {
				slog.Info("client disconnected from '%s' stream", gns.name)
				return
			}
			if err != nil {
				slog.Err("stream '%s' error: %v", gns.name, err)
				return
			}
			slog.Debug(9, "stream '%s', got %+v", gns.name, in)

			// send msg into toBackend chana
			// gns.ToBackend(in)
		}
	}()

	for {

		// wait for data from fromBackend chan
		//data := < gns.FromBackend()

		/*
			_ = &pb.FileEvent{
					message _Line_ {
						Name: name,
						Line: line,
					}
				Name: "", //r.BackendName()
				Event: &pb.FileEvent_Line_{
					Line: &pb.FileEvent_Line{Name: "journal", Line: `{"some":"json"}`},
				},
			}
		*/
		/*
			err := gns.send(*fe)
			if err == io.EOF {
				slog.Info("client disconnected from 'File' stream")
				return err
			}

			if err != nil {
				slog.Err("failed to send: %v", err)
				return err
			}
		*/
		time.Sleep(time.Second * 1)
	}

	return nil
}

func newServer() *gameNodeServer {
	return new(gameNodeServer)
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
