package server

// verificar esses erros e tentar novamente

// var (
// 	// Pre-allocating byte slices for paths and methods avoids string-to-byte conversions
// 	// in the hot path of request handling.
// 	// paymentsPath        = []byte("/payments")
// 	// paymentsSummaryPath = []byte("/payments-summary")
// 	// purgePaymentsPath   = []byte("/purge-payments")
// 	methodPost = []byte(fasthttp.MethodPost)
// 	methodGet  = []byte(fasthttp.MethodGet)
// )

// // requestPool holds fasthttp.Request objects to be reused.
// // Using a sync.Pool reduces memory allocations, which is critical for
// // low-latency applications as it minimizes GC pressure.
// var requestPool = sync.Pool{
// 	New: func() interface{} {
// 		return fasthttp.AcquireRequest()
// 	},
// }

// // responsePool holds fasthttp.Response objects to be reused.
// var responsePool = sync.Pool{
// 	New: func() interface{} {
// 		return fasthttp.AcquireResponse()
// 	},
// }

// // httpHandler is the core function that processes incoming connections.
// // It runs in a loop to handle multiple requests on the same connection (HTTP Keep-Alive).
// func httpHandler(ctx context.Context, connection netpoll.Connection) error {
// 	reader := connection.Reader()

// 	for {
// 		// Acquire request and response objects from the pool.
// 		req := requestPool.Get().(*fasthttp.Request)
// 		resp := responsePool.Get().(*fasthttp.Response)

// 		// It's crucial to release the objects back to the pool once they are no longer needed.
// 		// We use defer to ensure this happens even if a panic occurs.
// 		defer func() {
// 			req.Reset()
// 			resp.Reset()
// 			requestPool.Put(req)
// 			responsePool.Put(resp)
// 		}()

// 		// Set a read deadline to prevent connections from hanging indefinitely.
// 		connection.SetReadTimeout(5 * time.Second)

// 		// fasthttp.Request.Read reads from the netpoll connection reader and parses the HTTP request.
// 		// This is much more efficient than parsing manually.
// 		if err := req.Read(reader); err != nil {
// 			// io.EOF indicates the client has closed the connection.
// 			// netpoll.ErrReadTimeout indicates a timeout. Both are normal ways to end the loop.
// 			if err == io.EOF || err == netpoll.ErrReadTimeout {
// 				return nil // Gracefully close the connection handler.
// 			}
// 			// For other errors, log them and close the connection.
// 			log.Printf("Error reading request: %v", err)
// 			return err
// 		}

// 		// --- Routing Logic ---
// 		// This logic is similar to the original but adapted for the pooled req/resp objects.
// 		path := req.URI().Path()
// 		method := req.Header.Method()

// 		switch {
// 		case bytes.Equal(method, methodPost) && bytes.Equal(path, paymentsPath):
// 			// To safely pass the request body to a goroutine, we must create a copy.
// 			// The buffer underlying req.Body() will be reused by the pool for the next request.
// 			bodyCopy := make([]byte, len(req.Body()))
// 			copy(bodyCopy, req.Body())
// 			go handler.Payments(bodyCopy)
// 			resp.SetStatusCode(fasthttp.StatusOK)

// 		case bytes.Equal(method, methodGet) && bytes.Equal(path, paymentsSummaryPath):
// 			query := req.URI().QueryArgs().String()
// 			resp.Header.Set("Content-Type", "application/json")
// 			// The handler.PaymentsSummary now returns a byte slice directly.
// 			// This avoids extra conversions in the hot path.
// 			resp.SetBody(handler.PaymentsSummary(query))

// 		case bytes.Equal(method, methodPost) && bytes.Equal(path, purgePaymentsPath):
// 			handler.PaymentsPurge()
// 			resp.SetStatusCode(fasthttp.StatusOK)

// 		default:
// 			resp.SetStatusCode(fasthttp.StatusNotFound)
// 		}
// 		// --- End Routing Logic ---

// 		// Set a write deadline.
// 		connection.SetWriteTimeout(5 * time.Second)

// 		// Write the prepared response to the connection's writer.
// 		bufWriter := bufio.NewWriter(connection.Writer())
// 		if err := resp.Write(bufWriter); err != nil {
// 			log.Printf("Error writing response: %v", err)
// 			return err
// 		}
// 		if err := bufWriter.Flush(); err != nil {
// 			log.Printf("Error flushing response: %v", err)
// 			return err
// 		}

// 		// If the client requested to close the connection, we honor it and exit the loop.
// 		if req.ConnectionClose() {
// 			return nil
// 		}
// 	}
// }

// // ListenAndServe initializes the netpoll server.
// func ListenAndServeNetpoll(appPort string) {
// 	// Create a TCP listener.
// 	addr := ":" + appPort
// 	listener, err := net.Listen("tcp", addr)
// 	if err != nil {
// 		log.Fatalf("Failed to create listener: %v", err)
// 	}

// 	// Create a new event loop with our httpHandler as the OnRequest callback.
// 	// We configure it with options for better performance.
// 	eventLoop, err := netpoll.NewEventLoop(httpHandler,
// 		netpoll.WithReadTimeout(10*time.Second),
// 		netpoll.WithIdleTimeout(20*time.Second),
// 	)
// 	if err != nil {
// 		log.Fatalf("Failed to create event loop: %v", err)
// 	}

// 	log.Printf("Netpoll server starting on %s", addr)

// 	// Start the event loop to serve the listener. This is a blocking call.
// 	if err := eventLoop.Serve(listener); err != nil {
// 		log.Fatalf("Failed to start server: %v", err)
// 	}
// }
