package server

import (
	"bytes"
	"log"
	"net"
	"os"
	"sync"

	"github.com/jhamiltonjunior/rinha-de-backend/app/handler"
	"github.com/jhamiltonjunior/rinha-de-backend/app/services"
	"github.com/valyala/fasthttp"
)

var (
	paymentsPath        = []byte("/payments")
	paymentsSummaryPath = []byte("/payments-summary")
	purgePaymentsPath   = []byte("/purge-payments")
)

// netpoll

var (
	BufferPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 0, 1024)
			return &b
		},
	}
)

func requestHandler(ctx *fasthttp.RequestCtx) {
	path := ctx.Path()
	method := ctx.Method()

	switch {
	case bytes.Equal(method, []byte(fasthttp.MethodPost)) && bytes.Equal(path, paymentsPath):
		bufPtr := BufferPool.Get().(*[]byte)
		*bufPtr = append((*bufPtr)[:0], ctx.PostBody()...)
		services.PublishMessage(services.PaymentSubject, *bufPtr)
		BufferPool.Put(bufPtr)
		ctx.SetStatusCode(fasthttp.StatusOK)

	case bytes.Equal(method, []byte(fasthttp.MethodGet)) && bytes.Equal(path, paymentsSummaryPath):
		query := ctx.URI().QueryArgs().String()
		ctx.SetContentType("application/json")
		ctx.SetBody(handler.PaymentsSummary(query))

	case bytes.Equal(method, []byte(fasthttp.MethodPost)) && bytes.Equal(path, purgePaymentsPath):
		handler.PaymentsPurge()
		ctx.SetStatusCode(fasthttp.StatusOK)

	default:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
	}
}

func ListenAndServeFastHTTP(appPort string) {
	socketPath := os.Getenv("UNIX_SOCKET")
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		log.Fatalf("failed to remove existing socket: %v", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("failed to listen on unix socket: %v", err)
	}
	defer listener.Close()

	err = os.Chmod(socketPath, 0666)
	if err != nil {
		log.Fatalf("failed to set permissions on socket: %v", err)
	}
	log.Println("Server starting on", socketPath)
	if err := fasthttp.Serve(listener, requestHandler); err != nil {
		log.Fatalf("fasthttp server error: %v", err)
	}
}
