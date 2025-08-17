package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jhamiltonjunior/rinha-de-backend/app/utils"
	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	Key         = 0
	keys        = []string{"payment_history_1", "payment_history_2", "payment_history_3", "payment_history_4"}
)


func InitializeRedis() *redis.Client {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "redis_cache:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	return RedisClient
}

var paymentChan = make(chan []byte, 10000)

func StartRedisWorker(client *redis.Client) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var buffer [][]byte

	for {
		select {
		case payment := <-paymentChan:
			buffer = append(buffer, payment)
			if len(buffer) >= 10 {
				flushToRedis(client, buffer)
				buffer = buffer[:0]
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				flushToRedis(client, buffer)
				buffer = buffer[:0]
			}
		}
	}
}

func flushToRedis(client *redis.Client, payments [][]byte) {
	ctx := context.Background()
	pipe := client.Pipeline()

	for _, paymentData := range payments {
		key := keys[Key]
		pipe.LPush(ctx, key, paymentData)
		Key = (Key + 1) % len(keys)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("Erro no batch insert: %v", err)
	}
}

func CreatePaymentHistoryInMemory(client *redis.Client, paymentData map[string]any, typeService string) {

	newEntry := map[string]any{
		"amount":        paymentData["amount"],
		"requestedAt":   paymentData["requestedAt"],
		"type":          typeService,
	}

	entryBytes, err := json.Marshal(newEntry)
	if err != nil {
		log.Printf("Erro ao serializar entrada: %v", err)
		return
	}

	paymentChan <- entryBytes
}

func GetPaymentHistoryInMemory(client *redis.Client, from, to string) ([]PaymentHistory, error) {
	fromTime, err := time.Parse(utils.LayoutDate, from)
	if err != nil {
		log.Printf("Erro ao analisar data 'from': %v", err)
		return nil, err
	}
	toTime, err := time.Parse(utils.LayoutDate, to)
	if err != nil {
		log.Printf("Erro ao analisar data 'to': %v", err)
		return nil, err
	}

	fromNum := fromTime.UnixNano()
	toNum := toTime.UnixNano()

	ctx := context.Background()
	keys := []string{"payment_history_1", "payment_history_2", "payment_history_3", "payment_history_4"}

	var dataList []string
	for _, key := range keys {
		list, err := client.LRange(ctx, key, 0, -1).Result()
		if err != nil {
			log.Printf("Erro ao recuperar histórico do Redis para %s: %v", key, err)
			return nil, err
		}
		dataList = append(dataList, list...)
	}

	var filteredHistory []PaymentHistory
	for _, item := range dataList {
		var entry map[string]any
		if err := json.Unmarshal([]byte(item), &entry); err != nil {
			fmt.Println("Erro ao deserializar entrada:", err)
			continue
		}

		requestedAtStr, ok := entry["requestedAt"].(string)
		if !ok {
			continue
		}

		entryTime, err := time.Parse(utils.LayoutDate, requestedAtStr)
		if err != nil {
			fmt.Println("Erro ao analisar data de entrada:", err)
			continue
		}

		entryNum := entryTime.UnixNano()

		if entryNum > fromNum && entryNum < toNum {
			payment := PaymentHistory{
				// CorrelationId: entry["correlationId"].(string),
				Amount:      entry["amount"].(float64),
				RequestedAt: requestedAtStr,
				Type:        entry["type"].(string),
			}
			filteredHistory = append(filteredHistory, payment)
		}
	}

	return filteredHistory, nil
}

func PurgePaymentHistoryInMemory(client *redis.Client) {
	ctx := context.Background()

	for _, k := range keys {
		err := client.Del(ctx, k).Err()
		if err != nil {
			log.Printf("Erro ao limpar histórico de pagamentos: %v", err)
		}
	}
}
