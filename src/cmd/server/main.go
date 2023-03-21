package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	hellopb "github.com/takeru-a/pra_gRPC/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type myServer struct{
	hellopb.UnimplementedGreetingServiceServer
}

func NewMyServer() *myServer{
	return &myServer{}
}

func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest)(*hellopb.HelloResponse, error){
	return &hellopb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil
}

func(s *myServer)HelloServerStream(req *hellopb.HelloRequest, stream hellopb.GreetingService_HelloServerStreamServer) error{
	resCount := 5
	for i := 0; i < resCount; i++{
		if err := stream.Send(&hellopb.HelloResponse{
			Message: fmt.Sprintf("[%d] Hellom, %s!", i , req.GetName()),			
		});err != nil{
			return err
		}
		time.Sleep(time.Second * 1)
	}
	return nil
}

func(s *myServer) HelloClientStream(stream hellopb.GreetingService_HelloClientStreamServer) error{
	nameList := make([]string, 0)

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF){
			message := fmt.Sprintf("Hello, %v!", nameList)
			return stream.SendAndClose(&hellopb.HelloResponse{
				Message: message,
			})
		}
		if err != nil{
			return err
		}
		nameList = append(nameList, req.GetName())
	}
}

func (s *myServer) HelloBiStreams(stream hellopb.GreetingService_HelloBiStreamsServer) error{
	for{
		req, err := stream.Recv()
		if errors.Is(err, io.EOF){
			return nil
		}
		if err != nil{
			return err
		}
		message := fmt.Sprintf("Hello, %v!", req.GetName())
		if err := stream.Send(&hellopb.HelloResponse{
			Message: message,
		}); err != nil{
			return err
		}
	}
}

func main() {
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d",port))
	if err != nil{
		panic(err)
	}

	// gRPCサーバを作成
	s := grpc.NewServer()

	hellopb.RegisterGreetingServiceServer(s, NewMyServer())

	reflection.Register(s)

	// 稼働
	go func(){
		log.Printf("start gRPC server port :%v", port)
		s.Serve(listener)
	}()

	q := make(chan os.Signal, 1)
	signal.Notify(q, os.Interrupt)
	<- q
	log.Println("stopping gRPC server...")
	s.GracefulStop()
}