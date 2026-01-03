package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"net"
)

type HandlerError struct {
	StatuCode response.StatusCode
	Message   string
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	Closed  bool
	handler Handler
}

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	r, err := request.RequestFromReader(conn)
	responseWriter := response.NewWriter(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(*response.GetDefaultHeaders(0))
		return
	}

	s.handler(responseWriter, r)
}

func runServer(s *Server, listner net.Listener) {
	for {
		conn, err := listner.Accept()
		if s.Closed {
			return
		}
		if err != nil {
			return
		}
		runConnection(s, conn)
	}
}

func Serve(port uint, handler Handler) (*Server, error) {

	listner, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		Closed:  false,
		handler: handler,
	}
	go func() {
		runServer(server, listner)
	}()

	return server, nil
}

func (s *Server) Close() error {
	s.Closed = true
	return nil
}
