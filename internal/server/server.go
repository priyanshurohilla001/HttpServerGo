package server

import (
	"bytes"
	"fmt"
	"httpTime/internal/request"
	"httpTime/internal/response"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Server struct {
	Listener net.Listener
	Active   atomic.Bool
	Handler  Handler
}

func Serve(port int, Handler Handler) (*Server, error) {
	Listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	if err != nil {
		return nil, err
	}

	ser := &Server{
		Listener: Listener,
		Handler:  Handler,
	}

	ser.Active.Store(true)

	go ser.listen()

	return ser, nil

}

func (s *Server) Close() error {
	s.Active.Store(false)
	err := s.Listener.Close()
	return err
}

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()

		if err != nil {
			if s.Active.Load() {
				fmt.Printf("Error occurred while accepting connection : %v", err)
			}
			return
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	var buf bytes.Buffer

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}

		if writeErr := hErr.Write(conn); writeErr != nil {
			log.Printf("failed to write bad request to conn: %v (original error: %v)", writeErr, err)
		}

		return
	}

	if handlerErr := s.Handler(&buf, req); handlerErr != nil {
		if writeErr := handlerErr.Write(conn); writeErr != nil {
			log.Printf("Error occurred while writing handler error response: %v (handler error: %s)", writeErr, handlerErr.Message)
		}
		return
	}

	if err = response.WriteStatusLine(conn, response.StatusOK); err != nil {
		log.Printf("Error Occurred while Writing Status line to conn : %v", err)
		return
	}

	defaultHeaders := response.GetDefaultHeaders(buf.Len())

	err = response.WriteHeaders(conn, defaultHeaders)
	if err != nil {
		log.Printf("Error Occurred while Writing Default Headers to conn : %v", err)
		return
	}

	_, err = buf.WriteTo(conn)
	if err != nil {
		log.Printf("Error Occurred while Writing Body to conn : %v", err)
	}
}

func (e *HandlerError) Write(conn net.Conn) error {
	err := response.WriteStatusLine(conn, response.StatusCode(e.StatusCode))
	if err != nil {
		return err
	}

	defaultHeaders := response.GetDefaultHeaders(len(e.Message))

	err = response.WriteHeaders(conn, defaultHeaders)
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(e.Message))
	return err
}
