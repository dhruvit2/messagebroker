package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	broker "github.com/dhruvit2/messagebroker/pkg/broker"
	pb "github.com/dhruvit2/messagebroker/pkg/pb"
	replication "github.com/dhruvit2/messagebroker/pkg/replication"
	storage "github.com/dhruvit2/messagebroker/pkg/storage"
	"google.golang.org/grpc"
)

// Server implements the gRPC message broker server
type Server struct {
	pb.UnimplementedMessageBrokerServer
	broker         *broker.BrokerImpl
	replicationMgr *replication.ReplicationManager
	storage        *storage.Storage
	grpcServer     *grpc.Server
}

func main() {
	// Parse flags
	id := flag.Int("id", 1, "Broker ID")
	host := flag.String("host", "localhost", "Broker host")
	port := flag.Int("port", 9092, "Broker port")
	coordinatorURL := flag.String("coordinator", "localhost:2379", "Coordinator (etcd) URL")
	dataDir := flag.String("data-dir", "/tmp/messagebroker", "Data directory")
	flag.Parse()

	// Create broker config
	config := &broker.BrokerConfig{
		ID:                int32(*id),
		Host:              *host,
		Port:              *port,
		CoordinatorURL:    *coordinatorURL,
		MaxPartitions:     1000,
		ReplicationFactor: 3,
		MinISR:            2,
		RetentionMs:       604800000,  // 7 days
		SegmentMs:         86400000,   // 1 day
		SegmentBytes:      1073741824, // 1GB
		DataDir:           *dataDir,
		MetadataDir:       *dataDir + "/meta",
	}

	// Create broker instance
	b := broker.NewBroker(config)
	if err := b.Start(); err != nil {
		log.Fatalf("Failed to start broker: %v", err)
	}

	// Create replication manager
	replMgr := replication.NewReplicationManager(*id, config.ReplicationFactor, config.MinISR)

	// Create storage
	storageInstance := storage.NewStorage(*dataDir, config.SegmentBytes, config.RetentionMs)

	// Create gRPC server
	server := &Server{
		broker:         b,
		replicationMgr: replMgr,
		storage:        storageInstance,
	}

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server.grpcServer = grpc.NewServer()
	pb.RegisterMessageBrokerServer(server.grpcServer, server)

	log.Printf("Broker %d listening on %s:%d", *id, *host, *port)

	// Run server in goroutine
	go func() {
		if err := server.grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for signal to shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down broker...")
	server.grpcServer.GracefulStop()
	b.Stop()
	log.Println("Broker stopped")
}

// CreateTopic creates a new topic
func (s *Server) CreateTopic(ctx context.Context, req *pb.CreateTopicRequest) (*pb.CreateTopicResponse, error) {
	topic := &broker.Topic{
		Name:              req.Topic,
		NumPartitions:     req.NumPartitions,
		ReplicationFactor: req.ReplicationFactor,
		CreatedAt:         time.Now(),
	}

	if err := s.broker.CreateTopic(ctx, topic); err != nil {
		return nil, err
	}

	return &pb.CreateTopicResponse{
		Topic:      req.Topic,
		Partitions: req.NumPartitions,
		Success:    true,
	}, nil
}

// ProduceMessage sends a message to a topic
func (s *Server) ProduceMessage(ctx context.Context, req *pb.ProduceRequest) (*pb.ProduceResponse, error) {
	// Determine partition
	partition := req.Partition
	if partition < 0 {
		// Select partition based on key
		partition = 0 // Simplified
	}

	message := &broker.Message{
		ID:        generateMessageID(),
		Topic:     req.Topic,
		Partition: partition,
		Key:       req.Key,
		Value:     req.Value,
		Timestamp: time.Now(),
	}

	offset, err := s.broker.ProduceMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	// Store message
	s.storage.WriteMessage(req.Topic, partition, offset, map[string]interface{}{
		"key":   req.Key,
		"value": req.Value,
	})

	return &pb.ProduceResponse{
		Topic:     req.Topic,
		Partition: partition,
		Offset:    offset,
	}, nil
}

// ConsumeMessages fetches messages from a topic partition
func (s *Server) ConsumeMessages(ctx context.Context, req *pb.ConsumeRequest) (*pb.ConsumeResponse, error) {
	messages, err := s.broker.ConsumeMessages(ctx, req.Topic, req.Partition, req.Offset, req.MaxMessages)
	if err != nil {
		return nil, err
	}

	resp := &pb.ConsumeResponse{
		Topic:     req.Topic,
		Partition: req.Partition,
		Messages:  make([]*pb.Message, 0),
	}

	for _, msg := range messages {
		resp.Messages = append(resp.Messages, &pb.Message{
			Key:    msg.Key,
			Value:  msg.Value,
			Offset: msg.Offset,
		})
	}

	return resp, nil
}

// GetTopicMetadata returns metadata for a topic
func (s *Server) GetTopicMetadata(ctx context.Context, req *pb.GetTopicMetadataRequest) (*pb.TopicMetadata, error) {
	topic, err := s.broker.GetTopic(ctx, req.Topic)
	if err != nil {
		return nil, err
	}

	metadata := &pb.TopicMetadata{
		Topic:      topic.Name,
		Partitions: make([]*pb.PartitionMetadata, 0),
	}

	for i := int32(0); i < topic.NumPartitions; i++ {
		partition, _ := s.broker.GetPartition(ctx, topic.Name, i)
		if partition != nil {
			metadata.Partitions = append(metadata.Partitions, &pb.PartitionMetadata{
				Id:       partition.ID,
				Leader:   partition.Leader,
				Replicas: partition.Replicas,
				Isr:      partition.ISR,
			})
		}
	}

	return metadata, nil
}

// BrokerMetadata returns this broker's metadata
func (s *Server) BrokerMetadata(ctx context.Context, req *pb.BrokerMetadataRequest) (*pb.BrokerMetadataResponse, error) {
	brokerMeta, err := s.broker.GetBrokerMetadata(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.BrokerMetadataResponse{
		Id:        brokerMeta.BrokerID,
		Host:      brokerMeta.Host,
		Port:      int32(brokerMeta.Port),
		IsHealthy: s.broker.IsHealthy(),
	}, nil
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}
