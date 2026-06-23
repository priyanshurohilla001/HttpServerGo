package main

import (
	"httpTime/internal/request"
	"httpTime/internal/response"
	"httpTime/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {

	if req.RequestLine.RequestTarget == "/yourproblem" {
		htmlBody := []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)

		w.WriteStatusLine(response.StatusBadRequest)

		h := response.GetDefaultHeaders(len(htmlBody))
		h.Set("Content-Type", "text/html") // Tell browser to render HTML

		w.WriteHeaders(h)
		w.WriteBody(htmlBody)
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		htmlBody := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)

		w.WriteStatusLine(response.StatusInternalServerError)

		h := response.GetDefaultHeaders(len(htmlBody))
		h.Set("Content-Type", "text/html") // Tell browser to render HTML

		w.WriteHeaders(h)
		w.WriteBody(htmlBody)
		return
	}

	htmlBody := []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)

	w.WriteStatusLine(response.StatusOK)

	h := response.GetDefaultHeaders(len(htmlBody))
	h.Set("Content-Type", "text/html") // Tell browser to render HTML

	w.WriteHeaders(h)
	w.WriteBody(htmlBody)
}
