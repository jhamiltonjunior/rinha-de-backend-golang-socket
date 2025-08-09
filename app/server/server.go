package server

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/jhamiltonjunior/rinha-de-backend/app/handler"
	"github.com/panjf2000/gnet/v2"
)

type httpServer struct {
	gnet.BuiltinEventEngine
}

func (hs *httpServer) OnTraffic(c gnet.Conn) gnet.Action {
	buf, _ := c.Next(-1)
	if len(buf) == 0 {
		return gnet.None
	}

	firstLine, _ := bufio.NewReader(bytes.NewReader(buf)).ReadString('\n')
	parts := strings.Split(strings.TrimSpace(firstLine), " ")
	if len(parts) < 2 {
		c.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return gnet.Close
	}
	method := parts[0]
	pathWithQuery := parts[1]
	pathAndQuery := strings.Split(pathWithQuery, "?")
	path := pathAndQuery[0]
	query := ""
	if len(pathAndQuery) > 1 {
		query = pathAndQuery[1]
	}

	switch {
	case method == "POST" && path == "/payments":
		raw := buf
		sep := []byte("\r\n\r\n")
		idx := bytes.Index(raw, sep)

		body := raw[idx+len(sep):]

		go handler.Payments(body)
		c.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return gnet.Close

	case method == "GET" && strings.HasPrefix(path, "/payments-summary"):
		c.Write(handler.PaymentsSummary(query))
		return gnet.Close

	case method == "POST" && path == "/purge-payments":
		handler.PaymentsPurge()
		return gnet.Close

	}

	c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	return gnet.Close
}

func ListenAndServe(appPort string) {
	hs := &httpServer{}

	addr := fmt.Sprintf("tcp://:%s", appPort)

	err := gnet.Run(hs, addr,
		gnet.WithMulticore(true),
		gnet.WithLockOSThread(true),
		gnet.WithReusePort(true),
	)
	if err != nil {
		log.Fatalf("gnet server failed to start: %v", err)
	}
}
