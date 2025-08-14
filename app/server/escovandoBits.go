package server

import (
	"bufio"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jhamiltonjunior/rinha-de-backend/app/handler"
	"github.com/jhamiltonjunior/rinha-de-backend/app/services"
)

var bufPool = sync.Pool{
	New: func() any { return new([8192]byte) },
}
var OK = []byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")

const maxConns = 2000

func EscovandoBits(_ string) {
	socketPath := os.Getenv("UNIX_SOCKET")
	_ = os.Remove(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	if err := os.Chmod(socketPath, 0666); err != nil {
		log.Fatal("chmod error:", err)
	}

	sem := make(chan struct{}, maxConns)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}

		select {
		case sem <- struct{}{}:
			go func(c net.Conn) {
				defer func() {
					c.Close()
					<-sem
				}()
				handleConn(c)
			}(conn)
		default:
			conn.Write([]byte("HTTP/1.1 503 Service Unavailable\r\nConnection: close\r\nContent-Length: 0\r\n\r\n"))
			conn.Close()
		}
	}
}

func handleConn(c net.Conn) {
	br := bufio.NewReader(c)

	for {
		_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
		_ = c.SetWriteDeadline(time.Now().Add(2 * time.Second))

		reqLine, err := br.ReadString('\n')
		if err != nil {
			return
		}

		reqLine = strings.TrimSpace(reqLine)
		if reqLine == "" {
			return
		}

		var contentLength int
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			if strings.HasPrefix(strings.ToLower(line), "content-length:") {
				cl := strings.TrimSpace(line[len("content-length:"):])
				if v, err := strconv.Atoi(cl); err == nil {
					contentLength = v
				}
			}
		}

		switch {
		case strings.HasPrefix(reqLine, "POST /payments"):
			bufPtr := bufPool.Get().(*[8192]byte)
			defer bufPool.Put(bufPtr)
			buf := bufPtr[:]

			if contentLength > len(buf) {
				buf = make([]byte, contentLength)
			}
			_, err := br.Read(buf[:contentLength])
			if err != nil {
				return
			}

			services.PublishMessage(services.PaymentSubject, buf[:contentLength])
			c.Write(OK)

		case strings.HasPrefix(reqLine, "GET /payments-summary"):
			query := ""
			if idx := strings.Index(reqLine, "?"); idx != -1 {
				query = reqLine[idx+1:]
			}
			body := handler.PaymentsSummary(query)
			resp := []byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: " +
				strconv.Itoa(len(body)) + "\r\n\r\n")
			resp = append(resp, body...)
			c.Write(resp)

		case strings.HasPrefix(reqLine, "GET /payments-purge"):
			handler.PaymentsPurge()
			c.Write(OK)

		default:
			c.Write([]byte("HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n"))
		}
	}
}
