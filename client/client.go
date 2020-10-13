package main

import (
	"flag"
	"io"
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/hasebe/grpc-sample/wait"
)

var (
	serverAddr = flag.String("server_addr", "localhost:9080", "The server address in the format of host:port")
	count      = flag.Int("count", 3, "The number to interact messages")
	interval   = flag.Int("interval", 10, "The second between interacting messages")
	method     = flag.String("method", "printTime", "Method to call: Must be one of 'printTime', 'waitByServer', and 'waitByClient'")
)

func printTime(client pb.WaitClient) {
	log.Printf("Getting time...")

	empty := &pb.Empty{}

	response, err := client.GetTime(context.Background(), empty)
	if err != nil {
		log.Fatalf("Error when calling GetTime: %s", err)
	}
	log.Printf("Response from Server: %s", response.Time)
}

func waitByServer(client pb.WaitClient, connDetail *pb.ConnectionDetail) {
	log.Printf("Calling WaitByServer...")

	stream, err := client.WaitByServer(context.Background(), connDetail)
	if err != nil {
		log.Fatalf("Error when calling WaitByServer: %s", err)
	}

	for {
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.WaitByServer(_) = _, %v", client, err)
		}
		log.Printf("Receiving a message to the server: %v", message.Time)
	}
	log.Println("Finished WaitByServer")
}

func waitByClient(client pb.WaitClient, connDetail *pb.ConnectionDetail) {
	log.Println("Calling WaitByClient...")
	stream, err := client.WaitByClient(context.Background())
	if err != nil {
		log.Fatalf("Error when calling WaitByClient: %s", err)
	}
	cnt := connDetail.Count
	waittime := connDetail.Interval
	for cnt > 0 {
		message := &pb.Message{Time: time.Now().String()}
		log.Printf("Sending a message to the server: %v", message.Time)
		stream.Send(message)
		time.Sleep(time.Duration(waittime) * time.Second)
		cnt--
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	log.Printf("Finished WaitByClient: %v", reply.Time)
}

func main() {
	flag.Parse()

	var conn *grpc.ClientConn

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect :%s", err)
	}
	defer conn.Close()

	client := pb.NewWaitClient(conn)
	connDetail := &pb.ConnectionDetail{Interval: int32(*interval), Count: int32(*count)}

	switch *method {
	case "printTime":
		printTime(client)
	case "waitByServer":
		waitByServer(client, connDetail)
	case "waitByClient":
		waitByClient(client, connDetail)
	default:
		log.Fatalf("Method is not found: %v", *method)
	}
}
