package server

import (
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	Active   atomic.Bool
}

func Serve(port int) (*Server, error) {
	Listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	if err != nil {
		return nil, err
	}

	ser := &Server{
		Listener: Listener,
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
	for s.Active.Load() {
		conn, err := s.Listener.Accept()

		if err != nil {
			if s.Active.Load() {
				fmt.Printf("Error occured while accepting connection : %v", err)
			}
			return
		}

		go s.handle(conn)
	}
}
func (s *Server) handle(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello World!\n"))
	if err != nil {
		fmt.Printf("Error occured while Writing to connection : %v", err)
	}
	conn.Close()
}
