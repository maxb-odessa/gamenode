package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "gamenode/pkg/gamenodepb"

	"google.golang.org/grpc"
)

func runFile(client pb.GameNodeClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := client.File(ctx)
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

	fl := &pb.FileEvent_Line{Name: "journal", Line: `{"some":"json"}`}
	fe := &pb.FileEvent{Name: "some file name", Event: &pb.FileEvent_Line_{Line: fl}}
	if err := stream.Send(fe); err != nil {
		log.Fatalf("Failed to send a note: %v", err)
	}

	stream.CloseSend()
	<-waitc

}

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

	fl := &pb.JoyEvent_Button{}
	fe := &pb.JoyEvent{Name: "some joy name", Event: &pb.JoyEvent_Button_{Button: fl}}
	if err := stream.Send(fe); err != nil {
		log.Fatalf("Failed to send a note: %v", err)
	}

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

	go runJoy(client)
	runFile(client)
}
