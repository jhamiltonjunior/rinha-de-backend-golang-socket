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
)

// verificar o utra alterntiva ao redis

func InitializeRedis() *redis.Client {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "redis_cache:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	return RedisClient
}

func CreatePaymentHistoryInMemory(client *redis.Client, paymentData map[string]any, typeService string) {
	ctx := context.Background()

	keys := []string{"payment_history_1", "payment_history_2", "payment_history_3", "payment_history_4"}

	newEntry := map[string]any{
		// "correlationId": paymentData["correlationId"],
		"amount":        paymentData["amount"],
		"requestedAt":   paymentData["requestedAt"],
		"type":          typeService,
	}

	entryBytes, err := json.Marshal(newEntry)
	if err != nil {
		log.Printf("Erro ao serializar entrada: %v", err)
		return
	}

	// if typeService == "default" {
	// 	paymentData = append(paymentData, []byte(fmt.Sprintf(`,"type":"%"}`, 0))...)
	// } else {
	// 	paymentData = append(paymentData, []byte(fmt.Sprintf(`,"type":"%s"}`, 1))...)
	// }

	err = client.LPush(ctx, keys[Key], entryBytes).Err()
	if err != nil {
		log.Printf("Erro ao adicionar entrada ao histórico de pagamentos: %v", err)
	}

	Key = (Key + 1) % len(keys)
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
			continue
		}

		requestedAtStr, ok := entry["requestedAt"].(string)
		if !ok {
			continue
		}

		entryTime, err := time.Parse(utils.LayoutDate, requestedAtStr)
		if err != nil {
			continue
		}

		entryNum := entryTime.UnixNano()

		if entryNum > fromNum && entryNum < toNum {
			payment := PaymentHistory{
				// CorrelationId: entry["correlationId"].(string),
				Amount:        entry["amount"].(float64),
				RequestedAt:   requestedAtStr,
				Type:          entry["type"].(string),
			}
			filteredHistory = append(filteredHistory, payment)
		}
	}

	return filteredHistory, nil
}

func PurgePaymentHistoryInMemory(client *redis.Client) {
	ctx := context.Background()
	// key := "payment_history"

	keys := []string{"payment_history_1", "payment_history_2", "payment_history_3", "payment_history_4"}

	for _, k := range keys {
		err := client.Del(ctx, k).Err()
		if err != nil {
			log.Printf("Erro ao limpar histórico de pagamentos: %v", err)
		} else {
			fmt.Printf("Histórico de pagamentos '%s' limpo com sucesso.\n", k)
		}
	}

	// err := client.Del(ctx, key).Err()
	// if err != nil {
	// 	log.Printf("Erro ao limpar histórico de pagamentos: %v", err)
	// } else {
	// 	fmt.Println("Histórico de pagamentos limpo com sucesso.")
	// }
}
