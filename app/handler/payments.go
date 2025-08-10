package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jhamiltonjunior/rinha-de-backend/app/database"
	"github.com/jhamiltonjunior/rinha-de-backend/app/worker"
)

// --- Struct definitions remain the same ---

type Details struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type TypeDetails struct {
	Default  Details `json:"default"`
	Fallback Details `json:"fallback"`
}

func response(statusCode, statusText, contentType string, body []byte) []byte {
	contentLength := strconv.Itoa(len(body))
	// Using a strings.Builder is efficient for concatenating strings.
	var b strings.Builder
	b.WriteString("HTTP/1.1 ")
	b.WriteString(statusCode)
	b.WriteString(" ")
	b.WriteString(statusText)
	b.WriteString("\r\nContent-Type: ")
	b.WriteString(contentType)
	b.WriteString("\r\nContent-Length: ")
	b.WriteString(contentLength)
	b.WriteString("\r\n\r\n")

	// Combine headers and body
	responseBytes := make([]byte, 0, b.Len()+len(body))
	responseBytes = append(responseBytes, []byte(b.String())...)
	responseBytes = append(responseBytes, body...)
	return responseBytes
}

func emptyResponse(statusCode, statusText string) []byte {
	return []byte("HTTP/1.1 " + statusCode + " " + statusText + "\r\nContent-Length: 0\r\n\r\n")
}

// Helper functions for common responses.
func Accepted() []byte            { return emptyResponse("202", "Accepted") }
func NotFound() []byte            { return emptyResponse("404", "Not Found") }
func InternalServerError() []byte { return emptyResponse("500", "Internal Server Error") }
func BadRequest(body string) []byte {
	return response("400", "Bad Request", "text/plain", []byte(body))
}
func OK(body []byte) []byte { return response("200", "OK", "application/json", body) }

var cxt = context.Background()

func Payments(body []byte) {
	bufPtr := worker.BufferPool.Get().(*[]byte)
	*bufPtr = append((*bufPtr)[:0], body...)

	paymentWorker := worker.PaymentWorker{
		Body:              *bufPtr,
		VouTeDarOContexto: cxt,
		RetryCount:        0,
	}

	worker.SegureOChann <- paymentWorker

}

func PaymentsSummary(path string) []byte {
	from := "1970-01-01T00:00:00.000Z"
	to := "9999-12-31T23:59:00.000Z"

	if path != "" {
		queryParams, err := url.ParseQuery(path)
		if err == nil {
			if f := queryParams.Get("from"); f != "" {
				from = f
			}
			if t := queryParams.Get("to"); t != "" {
				to = t
			}
		}
	}

	payments, err := database.GetPaymentHistoryInMemory(database.RedisClient, from, to)
	if err != nil {
		fmt.Println("Error fetching payment history:", err)
		return InternalServerError()
	}

	var typeDetails TypeDetails
	for _, payment := range payments {
		switch payment.Type {
		case "default":
			typeDetails.Default.TotalRequests++
			typeDetails.Default.TotalAmount += payment.Amount
		case "fallback":
			typeDetails.Fallback.TotalRequests++
			typeDetails.Fallback.TotalAmount += payment.Amount
		}
	}

	paymentsSummary, err := json.Marshal(typeDetails)
	if err != nil {
		fmt.Println("Error serializing payment summary:", err)
		return InternalServerError()
	}

	return []byte(paymentsSummary)
}

func PaymentsPurge() {
	database.PurgePaymentHistoryInMemory(database.RedisClient)
}
