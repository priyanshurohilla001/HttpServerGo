package main

import (
	"crypto/sha256"
	"fmt"
	"httpTime/internal/headers"
	"httpTime/internal/request"
	"httpTime/internal/response"
	"httpTime/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {

	if req.RequestLine.RequestTarget == "/video" {
		videoData, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Printf("Error reading video file: %v", err)
			w.WriteStatusLine(response.StatusCodeInternalServerError)
			w.WriteHeaders(response.GetDefaultHeaders(0))
			return
		}

		w.WriteStatusLine(response.StatusCodeSuccess)

		h := response.GetDefaultHeaders(len(videoData))
		h.Override("Content-Type", "video/mp4")

		w.WriteHeaders(h)
		w.WriteBody(videoData)

		return
	}

	if req.RequestLine.RequestTarget == "/httpbin/html" {
		targetURL := "https://httpbin.org/html"
		proxyHtmlWithTrailers(w, targetURL)
		return // Stop processing so we don't fall through!
	}

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		proxyHandler(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler200(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}
	handler200(w, req)
	return
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func proxyHandler(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + target
	fmt.Println("Proxying to", url)
	resp, err := http.Get(url)
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Remove("Content-Length")
	w.WriteHeaders(h)

	const maxChunkSize = 1024
	buffer := make([]byte, maxChunkSize)
	for {
		n, err := resp.Body.Read(buffer)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}
}

func proxyHtmlWithTrailers(w *response.Writer, target string) {
	resp, err := http.Get(target)
	if err != nil {
		log.Printf("Proxy error reaching %s: %v", target, err)
		w.WriteStatusLine(response.StatusCodeInternalServerError)
		w.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	h.Remove("Content-Length")
	h.Set("Transfer-Encoding", "chunked")

	h.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(h)

	var mainBody []byte
	buf := make([]byte, 1024)

	for {
		n, readErr := resp.Body.Read(buf)

		if n > 0 {
			_, writeErr := w.WriteChunkedBody(buf[:n])
			if writeErr != nil {
				log.Printf("Client disconnected early: %v", writeErr)
				return
			}

			mainBody = append(mainBody, buf[:n]...)
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			log.Printf("Error reading from httpbin: %v", readErr)
			return
		}
	}

	hash := sha256.Sum256(mainBody)
	hashString := fmt.Sprintf("%x", hash)
	lengthString := fmt.Sprintf("%d", len(mainBody))

	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", hashString)
	trailers.Set("X-Content-Length", lengthString)

	err = w.WriteTrailers(trailers)
	if err != nil {
		log.Printf("Error writing trailers: %v", err)
	}
}
