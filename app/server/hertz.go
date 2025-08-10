package server

import (
	"bytes"

	"context"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/jhamiltonjunior/rinha-de-backend/app/handler"

	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

var (
	// paymentsPath        = []byte("/payments")
	// paymentsSummaryPath = []byte("/payments-summary")
	// purgePaymentsPath   = []byte("/purge-payments")

	methodPost = []byte([]byte("POST"))
	methodGet  = []byte([]byte("GET"))
)

func requestHandlerHertz(c context.Context, ctx *app.RequestContext) {
	path := ctx.Path()
	method := ctx.Method()

	switch {
	case bytes.Equal(method, methodPost) && bytes.Equal(path, paymentsPath):
		// Se quiser rodar async como no seu fasthttp, use goroutine
		handler.Payments(ctx.Request.Body())
		ctx.Status(200)

	case bytes.Equal(method, methodGet) && bytes.Equal(path, paymentsSummaryPath):
		query := string(ctx.QueryArgs().QueryString())
		ctx.Header("Content-Type", "application/json")
		ctx.Write(handler.PaymentsSummary(query))

	case bytes.Equal(method, methodPost) && bytes.Equal(path, purgePaymentsPath):
		handler.PaymentsPurge()
		ctx.Status(200)

	default:
		ctx.Status(404)
	}
}

func ListenAndServeHertz(appPort string) {
	// It's more idiomatic in Hertz to configure the server with options,
	// like setting the port, during initialization.
	h := server.Default(server.WithHostPorts(":" + appPort))

	// The handler functions must be explicitly cast to app.HandlerFunc.
	h.NoRoute(app.HandlerFunc(func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusNotFound, "404 Not Found")
	}))

	// Registering routes.
	// Note the casting to app.HandlerFunc.
	h.POST("/payments", app.HandlerFunc(requestHandlerHertz))
	h.GET("/payments-summary", app.HandlerFunc(requestHandlerHertz))
	h.POST("/purge-payments", app.HandlerFunc(requestHandlerHertz))

	// h.Spin() is the primary method to start the server and block
	// until it's stopped. It will handle errors internally and log them.
	// Using h.Spin() avoids the need for a manual error check like the one you had.
	log.Printf("Hertz server starting on port %s", appPort)
	h.Spin()
}
