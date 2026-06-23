package response

import (
	"fmt"
	"httpTime/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case 200:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, "OK")
		if err != nil {
			return err
		}
	case 400:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, "Bad Request")
		if err != nil {
			return err
		}
	case 500:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, "Internal Server Error")
		if err != nil {
			return err
		}
	default:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, "")
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {

	for key, value := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "\r\n")
	if err != nil {
		return err
	}

	return nil

}
