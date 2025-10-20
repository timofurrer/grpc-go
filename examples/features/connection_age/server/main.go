/*
 *
 * Copyright 2024 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a gRPC server that demonstrates connection age tracking.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	port               = ":50051"
	maxConnectionAge   = 30 * time.Second // Connection will be closed after 30s
	maxConnectionGrace = 5 * time.Second  // Grace period for in-flight requests
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	// Get connection age context to manage long operations
	ageCtx, ok := grpc.ConnectionAgeContext(ctx)
	if !ok {
		// Connection age context not available, use request context
		ageCtx = ctx
		log.Printf("Connection age context not available, using request context")
	} else {
		log.Printf("Connection age context received")
	}

	// Simulate some processing time for demonstration
	select {
	case <-ageCtx.Done():
		log.Printf("Operation stopped due to connection age: %v", ageCtx.Err())
		return &pb.HelloReply{
			Message: fmt.Sprintf("Hello %s! (Stopped early due to connection age)", in.GetName()),
		}, nil
	case <-time.After(1 * time.Second):
		// Normal processing completed
		log.Printf("Normal processing completed")
	case <-ctx.Done():
		log.Printf("Request cancelled due to context: %v", ctx.Err())
		return nil, ctx.Err()
	}

	return &pb.HelloReply{Message: fmt.Sprintf("Hello %s!", in.GetName())}, nil
}


func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Configure keepalive parameters to demonstrate connection age behavior
	s := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge:      maxConnectionAge,
			MaxConnectionAgeGrace: maxConnectionGrace,
			Time:                  10 * time.Second, // Send keepalive ping every 10s
			Timeout:               5 * time.Second,  // Wait 5s for ping ack
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // Min time between client pings
			PermitWithoutStream: true,           // Allow pings without streams
		}),
	)

	pb.RegisterGreeterServer(s, &server{})

	log.Printf("Server listening at %v", lis.Addr())
	log.Printf("Max connection age: %v", maxConnectionAge)
	log.Printf("Max connection grace: %v", maxConnectionGrace)
	log.Printf("Connection age context will automatically manage operation timeouts")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}