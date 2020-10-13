package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	pb "github.com/hasebe/grpc-sample/wait"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 9080, "The server port")
)

type waitServer struct {
	pb.UnimplementedWaitServer
}

func (s *waitServer) GetTime(ctx context.Context, empty *pb.Empty) (*pb.Message, error) {
	log.Printf("GetTime is called...")
	return &pb.Message{Time: time.Now().String()}, nil
}

func (s *waitServer) WaitByServer(connDetail *pb.ConnectionDetail, stream pb.Wait_WaitByServerServer) error {
	log.Printf("WaitByServer is called...")
	count := connDetail.Count
	interval := connDetail.Interval
	for count > 0 {
		time.Sleep(time.Duration(interval) * time.Second)
		message := &pb.Message{Time: time.Now().String()}
		log.Printf("Sending a message to the client: %v", message.Time)
		if err := stream.Send(message); err != nil {
			return err
		}
		count--
	}
	log.Printf("Finished WaitByServer")
	return nil
}

func (s *waitServer) WaitByClient(stream pb.Wait_WaitByClientServer) error {
	log.Printf("WaitByClient is called...")
	startTime := time.Now()
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			log.Printf("Finished WaitByClient")
			return stream.SendAndClose(&pb.Message{Time: endTime.Sub(startTime).String()})
		}
		if err != nil {
			return err
		}
		log.Printf("Receiving a message from the client: %v", message.Time)
	}
}

func newServer() *waitServer {
	return &waitServer{}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen on port 9000: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWaitServer(grpcServer, newServer())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over on port 9000: %v", err)
	}
}
