// server implements the TCP server
package server

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"

	"simp_Redis/commands"
	"simp_Redis/store"
)

// "Server" is a object for TCP server for the key-value store
type Server struct {
	store    *store.Store
	listener net.Listener
	quit     chan struct{}
}

func NewServer(s *store.Store) *Server {
	return &Server{
		store: s,
		quit:  make(chan struct{}),
	}
}

// begins listening for connections on the mentioned address
func (s *Server) Start(address string) error {

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.listener = listener

	log.Printf("Server started on %s", address)

	// Start a goroutine to clean up expired keys periodically
	go s.periodicCleanup()

	for {

		select {
		case <-s.quit:
			return nil
		default:

			s.listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))

			// Accepting a new connection
			conn, err := s.listener.Accept()
			if err != nil {

				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("Error accepting connection: %v", err)
				continue
			}

			// Handle each connection in a separate goroutine, (multithreading)
			go s.handleConnection(conn)
		}
	}
}

// processes commands from a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Set a read timeout to prevent hanging connections
	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

	log.Printf("New connection from %s", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	for {

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}

		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		if line == "" {
			continue
		}

		// Processes the command
		response := commands.Process(s.store, line)

		// Send response back to the client
		_, err = conn.Write([]byte(response + "\r\n"))
		if err != nil {
			log.Printf("Error writing to connection: %v", err)
			return
		}
	}
}

// The "periodicCleanup code" removes expired keys at regular intervals
func (s *Server) periodicCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.store.CleanupExpired()
		case <-s.quit:
			return
		}
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {

	close(s.quit)

	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
