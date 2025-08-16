package server

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/jhamiltonjunior/rinha-de-backend/app/handler"
	"github.com/jhamiltonjunior/rinha-de-backend/app/services"
)

var bufPool = sync.Pool{
	New: func() any { return new([4096]byte) },
}

var writerPool = sync.Pool{
	New: func() any {
		return bufio.NewWriter(nil)
	},
}

var (
	OK_RESPONSE          = []byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
	NOT_FOUND_RESPONSE   = []byte("HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n")
	UNAVAILABLE_RESPONSE = []byte("HTTP/1.1 503 Service Unavailable\r\nConnection: close\r\nContent-Length: 0\r\n\r\n")

	POST       = []byte("POST")
	GET        = []byte("GET")
	POSTLENGTH = len(POST)
	GETLENGTH  = len(GET)

	PAYMENTS_PATH         = []byte("/payments")
	PAYMENTS_SUMMARY_PATH = []byte("/payments-summary")
	PAYMENTS_PURGE_PATH   = []byte("/payments-purge")

	PAYMENTS_PATH_LENGTH         = len(PAYMENTS_PATH)
	PAYMENTS_SUMMARY_PATH_LENGTH = len(PAYMENTS_SUMMARY_PATH)
	PAYMENTS_PURGE_PATH_LENGTH   = len(PAYMENTS_PURGE_PATH)

	CONTENT_LENGTH_HEADER = []byte("\r\ncontent-length: ")

	HTTP_200_PREFIX  = []byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: ")
	HTTP_HEADERS_END = []byte("\r\n\r\n")
)

const maxConns = 4000

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
				handleConnV2(c)
			}(conn)
		default:
			conn.Write(UNAVAILABLE_RESPONSE)
			conn.Close()
		}
	}
}

func handleConnV2(c net.Conn) {
	bufPtr := bufPool.Get().(*[4096]byte)
	defer bufPool.Put(bufPtr)
	buf := bufPtr[:]

	bw := writerPool.Get().(*bufio.Writer)
	bw.Reset(c)
	defer func() {
		bw.Reset(nil)
		writerPool.Put(bw)
	}()

	for {
		deadline := time.Now().Add(1500 * time.Millisecond)
		_ = c.SetReadDeadline(deadline)
		_ = c.SetWriteDeadline(deadline)

		n, err := c.Read(buf)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}

		reqData := buf[:n]

		methodEnd := bytes.IndexByte(reqData, ' ')
		if methodEnd == -1 {
			return
		}
		method := reqData[:methodEnd]

		pathStart := methodEnd + 1
		pathEnd := bytes.IndexByte(reqData[pathStart:], ' ')
		if pathEnd == -1 {
			return
		}
		path := reqData[pathStart : pathStart+pathEnd]

		pathLen := len(path)
		methodLen := len(method)

		switch {
		case methodLen == POSTLENGTH && pathLen == PAYMENTS_PATH_LENGTH:
			bodyStart := bytes.Index(reqData, HTTP_HEADERS_END)
			if bodyStart == -1 {
				return
			}
			bodyStart += 4

			services.PublishMessage(services.PaymentSubject, reqData[bodyStart:])

			bw.Write(OK_RESPONSE)
			bw.Flush()

		case bytes.Equal(method, GET) && bytes.HasPrefix(path, PAYMENTS_SUMMARY_PATH):
			query := ""
			if queryIdx := bytes.IndexByte(path, '?'); queryIdx != -1 {
				query = string(path[queryIdx+1:])
			}

			body := handler.PaymentsSummary(query)

			bw.Write(HTTP_200_PREFIX)
			bw.Write(strconv.AppendInt(nil, int64(len(body)), 10))
			bw.Write(HTTP_HEADERS_END)
			bw.Write(body)
			bw.Flush()

		case methodLen == POSTLENGTH && pathLen == PAYMENTS_PURGE_PATH_LENGTH:
			handler.PaymentsPurge()
			bw.Write(OK_RESPONSE)
			bw.Flush()

		default:
			bw.Write(NOT_FOUND_RESPONSE)
			bw.Flush()
		}
	}
}
