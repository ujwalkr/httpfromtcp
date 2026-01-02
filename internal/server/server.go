package server

import (
	"bytes"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	Closed  bool
	handler Handler
}

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	headers := response.GetDefaultHeaders(20)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		response.WriteStatusLine(conn, response.StatusBadRequest)
		response.WriteHeaders(conn, *headers)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	handlerError := s.handler(writer, r)

	var body []byte = nil
	var statusCode response.StatusCode = response.StatusOK
	if handlerError != nil {
		statusCode = handlerError.StatuCode
		body = []byte(handlerError.Message)
	} else {
		body = writer.Bytes()
	}

	headers.Replace("content-length", fmt.Sprintf("%d", len(body)))
	response.WriteStatusLine(conn, statusCode)
	response.WriteHeaders(conn, *headers)
	conn.Write(body)
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
