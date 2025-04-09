package server

import (
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/isotronic/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
}

func Serve(port int) (*Server, error) {
	p := strconv.Itoa(port)
	l, err := net.Listen("tcp", ":" + p)
	if err != nil {
		return nil, err
	}
	server := Server{
		listener: l,
	}

	go server.listen()

	return &server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	err := s.listener.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) listen() {
	for {
		if s.isClosed.Load() {
			break
		}
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				break
			}
			log.Println("Error accepting connection:", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Println("Error writing status line:", err)
		return
	}
	headers := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Println("Error writing headers:", err)
		return
	}
	conn.Close()
}
