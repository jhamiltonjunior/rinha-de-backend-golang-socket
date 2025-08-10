package server

import (
	"bytes"
	"fmt"

	"github.com/jhamiltonjunior/rinha-de-backend/app/handler"
	"github.com/tidwall/evio"
)

var (
	postMethod = []byte("POST")
	getMethod  = []byte("GET")
	// paymentsPath        = []byte("/payments")
	// paymentsSummaryPath = []byte("/payments-summary")
	// purgePaymentsPath   = []byte("/purge-payments")

	httpEOL           = []byte("\r\n")
	httpHeaderBodySep = []byte("\r\n\r\n")
)

func parseRequest(buf []byte) (method, path, query, body []byte, ok bool) {
	// Encontra fim da primeira linha: "POST /payments HTTP/1.1"
	endFirstLine := bytes.Index(buf, httpEOL)
	if endFirstLine == -1 {
		return
	}
	firstLine := buf[:endFirstLine]

	// Método: até primeiro espaço
	space1 := bytes.IndexByte(firstLine, ' ')
	if space1 == -1 {
		return
	}
	method = firstLine[:space1]

	// Caminho e HTTP version
	rest := firstLine[space1+1:]
	space2 := bytes.IndexByte(rest, ' ')
	if space2 == -1 {
		return
	}
	fullPath := rest[:space2]

	// Separar path e query
	qIdx := bytes.IndexByte(fullPath, '?')
	if qIdx != -1 {
		path = fullPath[:qIdx]
		query = fullPath[qIdx+1:]
	} else {
		path = fullPath
	}

	// Separar body do header
	idx := bytes.Index(buf, httpHeaderBodySep)
	if idx == -1 {
		return
	}
	body = buf[idx+len(httpHeaderBodySep):]

	ok = true
	return
}

func ListenAndServeEvio(port string) {
	var events evio.Events

	events.Data = func(c evio.Conn, data []byte) (out []byte, action evio.Action) {
		method, path, query, body, ok := parseRequest(data)
		if !ok {
			// Resposta 400 Bad Request
			out = []byte("HTTP/1.1 400 Bad Request\r\nContent-Length: 0\r\n\r\n")
			action = evio.Close
			return
		}

		switch {
		case bytes.Equal(method, postMethod) && bytes.Equal(path, paymentsPath):
			// Processa em background para não bloquear o event-loop
			go handler.Payments(body)

			out = []byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
			action = evio.Close
			return

		case bytes.Equal(method, getMethod) && bytes.Equal(path, paymentsSummaryPath):
			res := handler.PaymentsSummary(string(query))
			out = []byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: ")
			out = append(out, []byte(fmt.Sprintf("%d\r\n\r\n", len(res)))...)
			out = append(out, res...)
			action = evio.Close
			return

		case bytes.Equal(method, postMethod) && bytes.Equal(path, purgePaymentsPath):
			handler.PaymentsPurge()
			out = []byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
			action = evio.Close
			return

		default:
			out = []byte("HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n")
			action = evio.Close
			return
		}
	}

	addr := fmt.Sprintf("tcp4://:%s", port)

	fmt.Printf("Starting evio server at %s\n", addr)
	if err := evio.Serve(events, addr); err != nil {
		panic(err)
	}
}
