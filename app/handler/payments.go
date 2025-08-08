package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jhamiltonjunior/rinha-de-backend/app/database"
	"github.com/jhamiltonjunior/rinha-de-backend/app/utils"
	"github.com/jhamiltonjunior/rinha-de-backend/app/worker"
	"github.com/valyala/fasthttp"
)

type Details struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type TypeDetails struct {
	Default  Details `json:"default"`
	Fallback Details `json:"fallback"`
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 1024)
		return &b
	},
}

func Payments(ctx *fasthttp.RequestCtx) {
	// bodyCopy := make([]byte, len(ctx.PostBody()))
	// copy(bodyCopy, ctx.PostBody())

	fmt.Println("Received payment request")
	bufPtr := bufferPool.Get().(*[]byte)

	body := ctx.PostBody()
	*bufPtr = append((*bufPtr)[:0], body...)
	ctx.SetStatusCode(202)

	cxt := context.TODO()

	now := time.Now().UTC()

	paymentWorker := worker.PaymentWorker{
		Body:              *bufPtr,
		VouTeDarOContexto: cxt,
		RequestedAt:       now.Format(utils.LayoutDate),
		RetryCount:        0,
	}

	worker.SegureOChann <- paymentWorker
}

func PaymentsSummary(ctx *fasthttp.RequestCtx) {
	from := ctx.QueryArgs().Peek("from")
	to := ctx.QueryArgs().Peek("to")

	if len(from) == 0 {
		from = []byte("1970-01-01T00:00:00.000Z")
	}

	if len(to) == 0 {
		to = []byte("9999-12-31T23:59:00.000Z")
	}

	payments, err := database.GetPaymentHistoryInMemory(database.RedisClient, string(from), string(to))
	if err != nil {
		fmt.Println("Erro ao buscar histórico de pagamentos:", err)
		sendJSONResponse(ctx, fasthttp.StatusInternalServerError)
		return
	}

	var typeDetails TypeDetails
	for _, payment := range payments {
		switch payment.Type {
		case "default":
			typeDetails.Default.TotalRequests++
			typeDetails.Default.TotalAmount += payment.Amount
		case "fallback":
			typeDetails.Fallback.TotalRequests++
			typeDetails.Fallback.TotalAmount += payment.Amount
		}
	}

	// ?from=2025-07-13T00:00:00&to=2025-07-13T14:33:48

	paymentsSummary, err := json.Marshal(typeDetails)
	if err != nil {
		fmt.Println("Erro ao serializar resumo de pagamentos:", err)
		sendJSONResponse(ctx, fasthttp.StatusInternalServerError)
		return
	}

    ctx.SetContentType("application/json")
    ctx.SetStatusCode(fasthttp.StatusOK)

    ctx.SetBody(paymentsSummary)
}

func PaymentsPurge(ctx *fasthttp.RequestCtx) {
	database.PurgePaymentHistoryInMemory(database.RedisClient)
	sendJSONResponse(ctx, fasthttp.StatusAccepted)
}
