package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

type PaymentHistory struct {
	CorrelationId string  `bson:"correlationId"`
	Amount        float64 `bson:"amount"`
	RequestedAt   string  `bson:"requestedAt"`
	Type          string     `bson:"type"`
}

func InitializeMongoDB() *mongo.Client {
	uri := "mongodb://root:very_hard_password@mongodb:27017/?authSource=admin"

	clientOptions := options.Client().ApplyURI(uri)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	MongoClient = client
	return client
}

func CreatePaymentHistory(client *mongo.Client, paymentData map[string]interface{}, typeService string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := client.Database("payment").Collection("payment_history").InsertOne(ctx, bson.M{
		"correlationId": paymentData["correlationId"],
		"amount":        paymentData["amount"],
		"requestedAt":   paymentData["requestedAt"],
		"type":          typeService,
	})

	if err != nil {
		log.Printf("Erro ao inserir pagamento: %v", err)
	}
}

func GetPaymentHistory(client *mongo.Client, from, to string) ([]PaymentHistory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"requestedAt": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	cursor, err := client.Database("payment").Collection("payment_history").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []PaymentHistory
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
