package server

import (
	"fmt"
	"net"

	"github.com/ramonvermeulen/httpfromtcp/internal/request"
	"github.com/ramonvermeulen/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   bool
}

func Serve(handler Handler, port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		handler:  handler,
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
	defer conn.Close()
	resWriter := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		headers := response.GetDefaultHeaders(0)
		resWriter.WriteStatusLine(response.StatusBadRequest)
		resWriter.WriteHeaders(headers)
	}
	s.handler(resWriter, req)
}

type Handler func(w *response.Writer, req *request.Request)
