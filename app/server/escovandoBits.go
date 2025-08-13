package server

import (
	"fmt"
	"log"
	"net"
	"os"
)
// docker compose down -v && docker compose up --build
func EscovandoBits(_ string) {
	socketPath := os.Getenv("UNIX_SOCKET")
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	fmt.Println("iniciando a delicia")
	fmt.Println("iniciando a delicia 2")
	err = os.Chmod(socketPath, 0666)
	if err != nil {
		log.Fatal("chmod error:", err)
	} else {
		fmt.Println("permissão alterada com sucesso")
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			// leitura simples (fasthttp/net/http faz parsing)
			buf := make([]byte, 4096)
			n, _ := c.Read(buf)

			// if buf[:n][0] == byte('{') {
			// 	fmt.Println("Recebido JSON")
			// } else {
			// 	fmt.Println("Não é JSON")
			// }

			fmt.Printf("raw request bytes: %s\n", buf[:n])
			// responder
			c.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nOK"))
		}(conn)
	}
}

// {
// 	"default": {
// 		"totalRequests": 0,
// 		"totalAmount": 0
// 	},
// 	"fallback": {
// 		"totalRequests": 0,
// 		"totalAmount": 0
// 	}
// }