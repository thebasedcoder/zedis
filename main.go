package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
)

const defaultListenAddr = ":5002"

type Config struct {
	ListenAddr string
}

type Server struct {
	Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan []byte
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln

	go s.loop()

	slog.Info("Server running...", "listen address", s.ListenAddr)
	return s.accepLoop()

}

func (s *Server) handleRawMsg(rawMsg []byte) error {
	fmt.Println(string(rawMsg))
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case rawMsg := <-s.msgCh:
			if err := s.handleRawMsg(rawMsg); err != nil {
				slog.Error("Raw message Error", "error", err)
			}
			// fmt.Println(rawMsg)
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh:
			s.peers[peer] = true

		}
	}
}

func (s *Server) accepLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "error", err)
			continue
		}
		s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer
	slog.Info("New peer connected", "Remote Address", conn.RemoteAddr())
	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err, "Remote Address", conn.RemoteAddr())
	}
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan []byte),
	}
}

func main() {
	server := NewServer(Config{})
	log.Fatal(server.Start())
}
