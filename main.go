package main

import (
	"os"

	"github.com/jhamiltonjunior/rinha-de-backend/app/database"
	"github.com/jhamiltonjunior/rinha-de-backend/app/server"
	"github.com/jhamiltonjunior/rinha-de-backend/app/services"
	"github.com/jhamiltonjunior/rinha-de-backend/app/worker"
)

func main() {
	clientRedis := database.InitializeRedis()

	natsURL := os.Getenv("NATS_URL")

	services.InitNATS(natsURL)

	worker.InitializeWorker(clientRedis)

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "3000"
	}

	// go pingQuantityOfSegureOChann()

	// os resultados foram bons com hertz, mas o fasthttp ainda parece ser mais rápido.
	// preciso testar o netpoll também, mas por enquanto vou manter o fasthttp como padrão.

	server.ListenAndServeGNET(appPort)
}

// func pingQuantityOfSegureOChann() {
// 	for {
// 		fmt.Printf("Quantidade de pagamentos pendentes: %d\n", len(worker.SegureOChann))
// 		fmt.Printf("Quantidade de pagamentos em retry: %d\n", len(worker.SegureOChann2))
// 		fmt.Println("========================================")
// 		time.Sleep(5 * time.Second)
// 	}
// }
