package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/JothamWong/serverless-suite/proto/aes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	zipkin                    = flag.String("zipkin", "http://localhost:9411/api/v2/spans", "zipkin url")
	address                   = flag.String("addr", "0.0.0.0:50052", "Address:Port the grpc server is listening to")
	key_string                = flag.String("key", "6368616e676520746869732070617373", "The key which is used for encryption")
	default_plaintext_message = flag.String("default-plaintext", "defaultplaintext", "Default plaintext when the function is called with the plaintext_message world")
)

func AESModeCTR(plaintext []byte) []byte {
		// Reference: cipher documentation
	// https://golang.org/pkg/crypto/cipher/#Stream

	key, _ := hex.DecodeString(*key_string)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	// We will use 0 to be predictable
	iv := make([]byte, aes.BlockSize)
	ciphertext := make([]byte, len(plaintext))

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, plaintext)
	return ciphertext
}

type server struct {
	pb.UnimplementedAesServer
}

// ShowEncryption implements aes.AesServer
func (s *server) ShowEncryption(ctx context.Context, in *pb.PlainTextMessage) (*pb.ReturnEncryptionInfo, error) {
	var plaintext, ciphertext []byte
	if in.GetPlaintextMessage() == "" || in.GetPlaintextMessage() == "world" {
		plaintext = []byte(*default_plaintext_message)
	} else {
		plaintext = []byte(in.GetPlaintextMessage())
	}
	ciphertext = AESModeCTR(plaintext)
	resp := fmt.Sprintf("fn: AES | plaintext: %s | ciphertext: %x | runtime: golang", plaintext, ciphertext)
	return &pb.ReturnEncryptionInfo{EncryptionInfo: resp}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Start AES-go server. Addr: %s\n", *address)

	var grpcServer *grpc.Server
	grpcServer = grpc.NewServer()
	pb.RegisterAesServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	
}