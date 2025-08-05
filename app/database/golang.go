package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jhamiltonjunior/rinha-de-backend/app/utils"
)

var Gotabase [][]byte

func CreatePaymentHistoryInMemorySlice(paymentData []byte, typeService string) {
	newProp := fmt.Appendf(nil, `"type":"%s"`, typeService)

	if len(paymentData) > 0 && paymentData[len(paymentData)-1] == '}' {
		var buf bytes.Buffer
		buf.Write(paymentData[:len(paymentData)-1])
		buf.WriteByte(',')
		buf.Write(newProp)
		buf.WriteByte('}')

		paymentData = buf.Bytes()
	}

	Gotabase = append(Gotabase, paymentData)
	fmt.Printf("Payment history in memory slice: %v\n", Gotabase)
}

func GetPaymentHistoryInMemorySlice(from, to string) ([]PaymentHistory, error) {
	fromTime, err := time.Parse(utils.LayoutDate, from)
	if err != nil {
		log.Printf("Erro ao analisar data 'from': %v", err)
		return nil, err
	}
	toTime, err := time.Parse(utils.LayoutDate, to)
	if err != nil {
		log.Printf("Erro ao analisar data 'to': %v", err)
		return nil, err
	}






	fazer conexao com os outros servidores 






	fromNum := fromTime.UnixNano()
	toNum := toTime.UnixNano()

	fmt.Printf("fromNum: %d, toNum: %d\n", fromNum, toNum)
	fmt.Printf("Payment history in memory slice: %v\n", Gotabase)
	fmt.Printf("Gotabase %s\n", Gotabase)

	var filteredHistory []PaymentHistory
	for _, item := range Gotabase {
		var entry map[string]any
		if err := json.Unmarshal(item, &entry); err != nil {
			fmt.Println("Erro ao analisar item:", err)
			continue
		}

		fmt.Println("Entry:", entry)
		fmt.Println("Item:", string(item))

		requestedAtStr, ok := entry["requestedAt"].(string)
		if !ok {
			fmt.Println("Campo 'requestedAt' não encontrado ou não é uma string")
			continue
		}

		entryTime, err := time.Parse(utils.LayoutDate, requestedAtStr)
		if err != nil {
			continue
		}

		entryNum := entryTime.UnixNano()
		fmt.Printf("Entry to: %d\n", toNum)
		fmt.Printf("Entry from: %d\n", fromNum)
		fmt.Printf("Entry number: %d\n", entryNum)

		if entryNum > fromNum && entryNum < toNum {
			payment := PaymentHistory{
				// CorrelationId: entry["correlationId"].(string),
				Amount:      entry["amount"].(float64),
				RequestedAt: requestedAtStr,
				Type:        entry["type"].(string),
			}
			filteredHistory = append(filteredHistory, payment)
		}
	}

	fmt.Printf("Filtered history: %v\n", filteredHistory)

	return filteredHistory, nil
}

func PurgePaymentHistoryInMemorySlice() {
	Gotabase = nil
}
