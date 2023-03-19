package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

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