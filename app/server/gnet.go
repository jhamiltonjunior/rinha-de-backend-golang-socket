package server

import (
	"bytes"
	"log"
	"os"
	"strconv"

	"github.com/jhamiltonjunior/rinha-de-backend/app/handler"
	"github.com/jhamiltonjunior/rinha-de-backend/app/services"
	"github.com/panjf2000/gnet/v2"
)

type paymentServer struct {
	gnet.BuiltinEventEngine
}

func (ps *paymentServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	return nil, gnet.None
}

func (ps *paymentServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	return gnet.None
}

func (ps *paymentServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	if bytes.HasPrefix(frame, []byte("POST /payments")) {
		headerEnd := bytes.Index(frame, []byte("\r\n\r\n"))
		if headerEnd == -1 {
			return nil, gnet.None
		}

		body := frame[headerEnd+4:]
		services.PublishMessage(services.PaymentSubject, body)

		out = []byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
		return out, gnet.None
	}

	if bytes.HasPrefix(frame, []byte("GET /payments-summary")) {
		body := handler.PaymentsSummary("")
		resp := []byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: " +
			strconv.Itoa(len(body)) + "\r\n\r\n")
		resp = append(resp, body...)
		return resp, gnet.None
	}

	if bytes.HasPrefix(frame, []byte("GET /payments-purge")) {
		handler.PaymentsPurge()
		return []byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"), gnet.None
	}

	return []byte("HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n"), gnet.None
}

func (ps *paymentServer) OnInitComplete(engine gnet.Engine) (action gnet.Action) {
    socketPath := os.Getenv("UNIX_SOCKET")
    if err := os.Chmod(socketPath, 0666); err != nil {
        log.Fatalf("failed to set permissions on socket: %v", err)
    }
    return gnet.None
}

func ListenAndServeGNET(_ string) {
	socketPath := os.Getenv("UNIX_SOCKET")
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		log.Fatalf("failed to remove existing socket: %v", err)
	}
	ps := &paymentServer{}
	addr := "unix://" + socketPath

	
	log.Fatal(gnet.Run(ps, addr, gnet.WithMulticore(true)))
}
