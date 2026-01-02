package server

import (
	"fmt"
	"io"
	"net"
)

type Server struct {
	Closed bool
}

func runConnection(server *Server, conn io.ReadWriteCloser) {
	out := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!")
	conn.Write(out)
	conn.Close()
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

func Serve(port uint) (*Server, error) {

	listner, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{Closed: false}
	go func() {
		runServer(server, listner)
	}()

	return server, nil
}

func (s *Server) Close() error {
	s.Closed = true
	err := s.Close()
	return err
}
