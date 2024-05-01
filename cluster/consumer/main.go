package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "consumer/grpc"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/proto"
)

const (
	kafkaServer = "kafka:9092"
	kafkaTopic  = "vote-topic"
	mongoURI    = "mongodb://mongodb:27017"
	redisAddr   = "redis:6379"
)

// Estructuras de configuración para los clientes de Kafka, Redis y MongoDB
func createClients() (*kafka.Reader, *redis.Client, *mongo.Collection, func(), error) {
	kafkaConfig := kafka.ReaderConfig{
		Brokers:     []string{kafkaServer},
		Topic:       kafkaTopic,
		MinBytes:    10e3,
		MaxBytes:    10e6,
		MaxAttempts: 5,
	}
	kafkaReader := kafka.NewReader(kafkaConfig)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	mongoOpts := options.Client().ApplyURI(mongoURI).SetMaxPoolSize(200).SetMaxConnIdleTime(60 * time.Second).SetConnectTimeout(30 * time.Second).SetSocketTimeout(60 * time.Second)
	mongoClient, err := mongo.Connect(context.Background(), mongoOpts)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	collection := mongoClient.Database("logs").Collection("vote-logs")
	cleanup := func() {
		kafkaReader.Close()
		redisClient.Close()
		mongoClient.Disconnect(context.Background())
	}

	return kafkaReader, redisClient, collection, cleanup, nil
}

func main() {
	reader, redisClient, logsCollection, cleanup, err := createClients()
	if err != nil {
		log.Fatalf("Failed to setup clients: %v", err)
	}
	defer cleanup()

	processVotes(reader, redisClient, logsCollection)
}

// Leer y procesar mensajes de Kafka
func processVotes(reader *kafka.Reader, redisClient *redis.Client, collection *mongo.Collection) {
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			logToMongo(collection, "Failed to read Kafka message", err)
			continue
		}

		var request pb.RequestId
		if err := proto.Unmarshal(msg.Value, &request); err != nil {
			logToMongo(collection, "Failed to unmarshal Kafka message", err)
			continue
		}

		record := fmt.Sprintf(`album: "%s", year: "%s", artist: "%s", ranked: %s`, request.Album, request.Year, request.Artist, request.Ranked)

		errorincr := redisClient.HIncrBy(context.Background(), "vote_records", record, 1).Err()
		if errorincr != nil {
			logToMongo(collection, "Failed to push to Redis", errorincr)
			continue
		}

		// if _, err = redisClient.HIncrBy(context.Background(), "vote_records", record).Err(); err != nil {
		// 	logToMongo(collection, "Failed to push to Redis", err)
		// 	continue
		// }

		msg_processed := fmt.Sprintf("Message processed and saved: %s", record)
		logToMongo(collection, msg_processed, nil)
	}
}

// Función para registrar mensajes y errores en MongoDB
func logToMongo(collection *mongo.Collection, message string, err error) {
	logEntry := bson.M{
		"timestamp": time.Now(),
		"message":   message,
		"error":     fmt.Sprint(err),
	}

	if _, err := collection.InsertOne(context.Background(), logEntry); err != nil {
		log.Printf("Failed to log in MongoDB: %v", err)
	}
	log.Printf("%s: %v", message, fmt.Sprint(err))
}
