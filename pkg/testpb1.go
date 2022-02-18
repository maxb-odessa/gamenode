package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "gamenode/pkg/gamenodepb"

	"google.golang.org/grpc"
)

func runJoy(client pb.GameNodeClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.Joy(ctx)
	if err != nil {
		log.Fatalf("%v.GameNode(_) = _, %v", client, err)
	}

	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}

			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("Got joy message %+v", in)
		}
	}()

	jb := pb.JoyEvent_Button_{Button: &pb.JoyEvent_Button{Pressed: false, Color: "GREEN"}}
	jd := pb.JoyEvent{Obj: &jb}
	jmsg := &pb.JoyMsg{
		Name: "ANOTHER X52PRO",
		Msg:  &pb.JoyMsg_Event{Event: &jd},
	}

	if err := stream.Send(jmsg); err != nil {
		log.Fatalf("Failed to send a note: %v", err)
	}

	time.Sleep(time.Second * 5)

	stream.CloseSend()
	<-waitc

}

func main() {
	conn, err := grpc.Dial("127.0.0.1:12346", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewGameNodeClient(conn)

	runJoy(client)
}
