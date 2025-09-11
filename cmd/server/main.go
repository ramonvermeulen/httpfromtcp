package server

import (
	"fmt"
	"net"

	"github.com/ramonvermeulen/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		closed:   false,
	}

	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if s.closed {
			return
		}
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	response.WriteStatusLine(conn, 200)
	response.WriteHeaders(conn, response.GetDefaultHeaders(0))
	conn.Close()
}
