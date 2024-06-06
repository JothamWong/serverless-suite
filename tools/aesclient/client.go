// Package main implements a client for the AES service
package main

import (
	"context"
	"flag"
	"time"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/JothamWong/serverless-suite/proto/aes"
)

const defaultName = "world"

var (
	addr = flag.String("addr", "localhost:50052", "address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
	n = flag.Int("n", 10, "Number of invocations")
)

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	// unsecure
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(*addr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewAesClient(conn)
	// contact server and print response
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Try once first to see if succeedss
	r, err := client.ShowEncryption(ctx, &pb.PlainTextMessage{PlaintextMessage: *name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greetings: %s", r.GetEncryptionInfo())
	// Now loop n times. To insert m5 magic instruction here (?)
	for i := 0; i < *n; i++ {
		client.ShowEncryption(ctx, &pb.PlainTextMessage{PlaintextMessage: *name})
		if i % 10 == 0 {
			log.Printf("Invoked for %d times\n", i)
		}
	}
	log.Printf("Finished calling function for %d times: %s", *n, r.GetEncryptionInfo())
}