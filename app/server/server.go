package server

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jhamiltonjunior/rinha-de-backend/app/handler"
	"github.com/valyala/fasthttp"
)

func ListenAndServe(appPort string) {
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/payments":
			if ctx.IsPost() {
				handler.Payments(ctx)
				return
			}
		case "/payments-summary":
			if ctx.IsGet() {
				handler.PaymentsSummary(ctx)
				return
			}
		case "/purge-payments":
			if ctx.IsPost() {
				handler.PaymentsPurge(ctx)
				return
			}
		}

		ctx.Error("Not Found", fasthttp.StatusNotFound)
	}

	fmt.Println("UNIX SOCKET:" + os.Getenv("UNIX_SOCKET"))

	UNIX_SOCKET := os.Getenv("UNIX_SOCKET")

	err := os.Remove(UNIX_SOCKET)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error removing UNIX socket: %v", err)
	}

	listener, err := net.Listen("unix", UNIX_SOCKET)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	os.Chmod(UNIX_SOCKET, 0666)

	if err := fasthttp.Serve(listener, requestHandler); err != nil {
		fmt.Printf("Error in ListenAndServe: %s\n", err)
	}
}
