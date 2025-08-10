package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jhamiltonjunior/rinha-de-backend/app/database"
	"github.com/jhamiltonjunior/rinha-de-backend/app/services"
	"github.com/jhamiltonjunior/rinha-de-backend/app/utils"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

type PaymentWorker struct {
	Body              []byte
	VouTeDarOContexto context.Context
	RetryCount        int
	RequestedAt       string
}

var (
	SegureOChann  = make(chan PaymentWorker, 4000)
	SegureOChann2 = make(chan PaymentWorker, 4000)
	BufferPool    = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 0, 1024)
			return &b
		},
	}
)

func InitializeWorker(client *redis.Client) {
	defaultURL := os.Getenv("PAYMENT_PROCESSOR_URL_DEFAULT")
	fallbackURL := os.Getenv("PAYMENT_PROCESSOR_URL_FALLBACK")
	const numWorkers = 20
	queueGroup := "worker-group-1"

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	workerLogic := func(_ int) nats.MsgHandler {
		return func(msg *nats.Msg) {
			paymentWorker := PaymentWorker{
				Body:              msg.Data,
				VouTeDarOContexto: context.TODO(),
				RetryCount:        0,
			}
			if !workerFunc(client, defaultURL, fallbackURL, paymentWorker) {
				paymentWorker.RetryCount++
				SegureOChann2 <- paymentWorker
			}
		}
	}

	for i := 1; i <= numWorkers; i++ {
		workerID := i
		_, err := services.NC.QueueSubscribe(services.PaymentSubject, queueGroup, workerLogic(workerID))
		if err != nil {
			log.Fatalf("Worker %d failed to subscribe: %v", workerID, err)
		}
	}

	for i := 1; i <= numWorkers; i++ {
		go retryworkLoop(client, defaultURL, fallbackURL)
	}
}

func workerFunc(client *redis.Client, defaultURL, fallbackURL string, payment PaymentWorker) bool {
	body, ok := ProcessPayment(payment.Body, payment.VouTeDarOContexto, defaultURL, payment.RequestedAt)
	if ok {
		BufferPool.Put(&payment.Body)
		database.CreatePaymentHistoryInMemory(client, body, "default")
		return true
	}

	if payment.RetryCount <= 15 {
		// fmt.Println(payment.RetryCount)
		return false
	}

	body, ok = ProcessPayment(payment.Body, payment.VouTeDarOContexto, fallbackURL, payment.RequestedAt)
	if ok {
		BufferPool.Put(&payment.Body)
		database.CreatePaymentHistoryInMemory(client, body, "fallback")
		return true
	}

	return false
}

func retryworkLoop(client *redis.Client, defaultURL, fallbackURL string) {
	for payment := range SegureOChann2 {
		func(payment PaymentWorker) {
			cxt, cancel := context.WithTimeout(context.Background(), 105*time.Second)
			defer cancel()
			payment.VouTeDarOContexto = cxt

			if !workerFunc(client, defaultURL, fallbackURL, payment) {
				payment.RetryCount = payment.RetryCount + 1
				SegureOChann2 <- payment
			}
		}(payment)
	}
}

func ProcessPayment(paymentBytes []byte, ctx context.Context, theBestURLEver string, requestedAt string) (map[string]any, bool) {
	var payment map[string]any
	if err := json.Unmarshal(paymentBytes, &payment); err != nil {
		fmt.Println("Erro ao deserializar o pagamento:", err)
		return nil, false
	}

	payment["requestedAt"] = time.Now().UTC().Format(utils.LayoutDate)

	paymentBytes, err := json.Marshal(payment)
	if err != nil {
		fmt.Println("Erro ao serializar o pagamento:", err)
		return nil, false
	}

	return payment, sendToPaymentService(paymentBytes, theBestURLEver, ctx)
}

func sendToPaymentService(paymentBytes []byte, reqURL string, ctx context.Context) bool {
	_, status := utils.Request("POST", paymentBytes, reqURL+"/payments", ctx)
	// fmt.Printf("status: %d\n", status)
	return status == 200 || status == 201
}
