package server

import (
	"bytes"
	"context"
	"log"
	"os" // Import the 'os' package for file operations
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/jhamiltonjunior/rinha-de-backend/app/handler"
	"github.com/jhamiltonjunior/rinha-de-backend/app/services"
)

var (
	// Path definitions remain the same.
	methodPost = []byte("POST")
	methodGet  = []byte("GET")
)

// The request handler logic does not need to change, as it operates
// at the application layer, independent of the network transport.
func requestHandlerHertz(c context.Context, ctx *app.RequestContext) {
	path := ctx.Path()
	method := ctx.Method()

	switch {
	case bytes.Equal(method, methodPost) && bytes.Equal(path, paymentsPath):
		timestart := time.Now()
		bufPtr := BufferPool.Get().(*[]byte)
		*bufPtr = append((*bufPtr)[:0], ctx.Request.Body()...)
		services.PublishMessage(services.PaymentSubject, *bufPtr)
		BufferPool.Put(bufPtr)
		timeend := time.Since(timestart)
		log.Printf("Tempo gasto para copiar o body: %s", timeend)
		ctx.Status(consts.StatusOK)

	case bytes.Equal(method, methodGet) && bytes.Equal(path, paymentsSummaryPath):
		query := string(ctx.QueryArgs().QueryString())
		ctx.Header("Content-Type", "application/json")
		ctx.Write(handler.PaymentsSummary(query))

	case bytes.Equal(method, methodPost) && bytes.Equal(path, purgePaymentsPath):
		handler.PaymentsPurge()
		ctx.Status(consts.StatusOK)

	default:
		ctx.Status(consts.StatusNotFound)
	}
}

// ListenAndServeHertz is updated to listen on a Unix socket.
// The function now accepts a socket file path instead of a port number.
func ListenAndServeHertz(_ string) {
	socketPath := os.Getenv("UNIX_SOCKET")
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		log.Fatalf("failed to remove existing socket: %v", err)
	}
	// Here's the main change:
	// We configure the Hertz server to listen on a Unix socket by providing
	// the WithNetwork("unix") and WithHostPorts(socketPath) options.
	h := server.Default(
		server.WithNetwork("unix"),
		server.WithHostPorts(socketPath),
	)

	// The route registration logic remains the same.
	h.NoRoute(app.HandlerFunc(func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusNotFound, "404 Not Found")
	}))

	err := os.Chmod(socketPath, 0666)
	if err != nil {
		log.Fatalf("failed to set permissions on socket: %v", err)
	}

	h.POST("/payments", app.HandlerFunc(requestHandlerHertz))
	h.GET("/payments-summary", app.HandlerFunc(requestHandlerHertz))
	h.POST("/purge-payments", app.HandlerFunc(requestHandlerHertz))

	go func() {
		// Aplica permissões após o socket ser criado
		for {
			if _, err := os.Stat(socketPath); err == nil {
				_ = os.Chmod(socketPath, 0666)
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Update the log message to show it's listening on a socket.
	log.Printf("Hertz server starting on unix socket: %s", socketPath)
	h.Spin()
}
