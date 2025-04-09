package server

import (
	"bytes"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/isotronic/httpfromtcp/internal/request"
	"github.com/isotronic/httpfromtcp/internal/response"
)

type Server struct {
	handler		Handler
	listener 	net.Listener
	isClosed 	atomic.Bool
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func Serve(port int, handler Handler) (*Server, error) {
	p := strconv.Itoa(port)
	l, err := net.Listen("tcp", ":" + p)
	if err != nil {
		return nil, err
	}
	server := Server{
		handler: handler,
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
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println("Error parsing request:", err)
		return
	}

	buf := bytes.Buffer{}
	handlerError := s.handler(&buf, req)
	if handlerError != nil {
		err = response.WriteStatusLine(conn, handlerError.StatusCode)
		if err != nil {
			log.Println("Error writing status line:", err)
			return
		}

		headers := response.GetDefaultHeaders(len(handlerError.Message))
		err = response.WriteHeaders(conn, headers)
		if err != nil {
			log.Println("Error writing headers:", err)
			return
		}

		_, err = conn.Write([]byte(handlerError.Message))
		if err != nil {
			log.Println("Error writing response:", err)
			return
		}
	} else {
		err = response.WriteStatusLine(conn, response.StatusOK)
		if err != nil {
			log.Println("Error writing status line:", err)
			return
		}

		headers := response.GetDefaultHeaders(len(buf.String()))
		err = response.WriteHeaders(conn, headers)
		if err != nil {
			log.Println("Error writing headers:", err)
			return
		}

		_, err = conn.Write(buf.Bytes())
		if err != nil {
				log.Println("Error writing body:", err)
				return
		}
	}
}
