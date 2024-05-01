package main

import (
	"context"
	"log"
	"net"

	pb "producer/grpc"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type grpcServer struct {
	pb.UnimplementedGetInfoServer
}

var kafkaPub = kafka.NewWriter(kafka.WriterConfig{
	Brokers: []string{"kafka:9092"},
	Topic:   "vote-topic",
})

func publishToKafka(data []byte) error {
	return kafkaPub.WriteMessages(context.Background(), kafka.Message{
		Value: data,
	})
}

func (g *grpcServer) ReturnInfo(ctx context.Context, req *pb.RequestId) (*pb.ReplyInfo, error) {
	log.Printf("Received from gRPC client: %v", req)

	encodedMessage, err := proto.Marshal(req)
	if err != nil {
		log.Printf("Error encoding gRPC message: %v", err)
		return nil, err
	}

	if err := publishToKafka(encodedMessage); err != nil {
		log.Printf("Error publishing to Kafka: %v", err)
		return nil, err
	}

	return &pb.ReplyInfo{Info: "Data received and forwarded to Kafka"}, nil
}

func setupServer() {
	serverAddr := ":3001"
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Unable to listen on %v: %v", serverAddr, err)
	}

	grpcSvc := grpc.NewServer()
	pb.RegisterGetInfoServer(grpcSvc, &grpcServer{})

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := grpcSvc.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed to serve: %v", err)
	}
}

func main() {
	setupServer()
}
