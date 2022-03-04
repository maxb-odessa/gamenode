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
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Second)
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
			log.Printf("Got file message %+v", in)
		}
	}()
	/*
		msg := pb.FileMsg{
			Name:  "ed journal",
			SeqNo: 99999,
			Msg: &pb.FileMsg_Event{
				Event: &pb.FileEvent{
					Line: "must be skipped!",
				},
			},
		}

		if err := stream.Send(&msg); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
	*/
	time.Sleep(time.Second * 500)

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

	runFile(client)
}
