package server

import (
	"bytes"
	"fmt"
	"io"
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
	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusError,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	buff := bytes.Buffer{}
	hErr := s.handler(&buff, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}

	response.WriteStatusLine(conn, response.StatusOK)
	response.WriteHeaders(conn, response.GetDefaultHeaders(0))
	conn.Write(buff.Bytes())
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (h *HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, h.StatusCode)
	response.WriteHeaders(w, response.GetDefaultHeaders(len([]byte(h.Message))))
	w.Write([]byte(h.Message))
}

type Handler func(w io.Writer, req *request.Request) *HandlerError
