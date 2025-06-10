package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/osm/qwbs/internal/qw/broadcast"
	"github.com/osm/qwbs/internal/qw/command"
	"github.com/osm/qwbs/internal/qw/master"
	"github.com/osm/qwbs/internal/qw/serverstatus"
	"github.com/osm/qwbs/internal/version"
	"github.com/osm/qwbs/internal/writer"
)

const (
	bufSize     = 1024 * 64
	readTimeout = time.Second
	url         = "https://github.com/osm/qwbs"
)

type Server struct {
	conn        *net.UDPConn
	logger      *slog.Logger
	listenAddr  *net.UDPAddr
	masters     map[string]*master.Master
	masterAddrs []*net.UDPAddr
	writers     []writer.Writer
}

func New(
	logger *slog.Logger,
	listenAddr *net.UDPAddr,
	masterAddrs []*net.UDPAddr,
	writers []writer.Writer) *Server {
	return &Server{
		logger:      logger,
		listenAddr:  listenAddr,
		masters:     make(map[string]*master.Master),
		masterAddrs: masterAddrs,
		writers:     writers,
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	conn, err := net.ListenUDP("udp", s.listenAddr)
	if err != nil {
		return err
	}

	s.conn = conn
	s.initMasters()
	s.registerMasters()

	buf := make([]byte, bufSize)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			s.logger.Error("Failed to set read deadline", "error", err)
			continue
		}

		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				s.logger.Info("Closing server")
				return nil
			}

			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				s.registerMasters()
				continue
			}

			s.logger.Error("Failed to read from UDP",
				"client", clientAddr, "error", err)
			continue
		}

		cmd, payload := command.Parse(buf[:n])
		switch cmd {
		case command.ACK:
			s.handleMasterACK(ctx, clientAddr)
		case command.Ping:
			s.handlePing(clientAddr)
		case command.GetChallenge:
			s.handleGetChallenge(clientAddr)
		case command.Status:
			s.handleStatus(clientAddr)
		case command.Broadcast:
			s.handleBroadcast(ctx, clientAddr, payload)
		default:
			s.logger.Debug("Unexpected data received",
				"client", clientAddr, "length", n)
		}
	}
}

func (s *Server) initMasters() {
	for _, addr := range s.masterAddrs {
		key := addr.String()

		if _, ok := s.masters[key]; ok {
			continue
		}

		s.masters[key] = master.New(s.conn, addr, s.logger)
	}
}

func (s *Server) registerMasters() {
	for _, m := range s.masters {
		if err := m.Register(); err != nil {
			s.logger.Error("Failed to register master", "error", err)
		}
	}
}

func (s *Server) handleMasterACK(ctx context.Context, clientAddr *net.UDPAddr) {
	m, ok := s.masters[clientAddr.String()]
	if !ok {
		s.logger.Error("Unexpected ACK received", "client", clientAddr)
		return
	}

	go m.Heartbeat(ctx)
}

func (s *Server) handlePing(clientAddr *net.UDPAddr) {
	s.logger.Debug("Sending ACK", "client", clientAddr)

	_, err := s.conn.WriteToUDP(command.GetACKBytes(), clientAddr)
	if err != nil {
		s.logger.Error("Failed to send ACK",
			"client", clientAddr, "error", err)
	}
}

func (s *Server) handleGetChallenge(clientAddr *net.UDPAddr) {
	_, err := s.conn.WriteToUDP(command.GetPrintBytes("%s", version.Name()), clientAddr)
	if err != nil {
		s.logger.Error("Failed to send get challenge response",
			"client", clientAddr, "error", err)
	}
}

func (s *Server) handleStatus(clientAddr *net.UDPAddr) {
	payload := command.GetPrintBytes("\\*version\\%s\\broadcast\\1\\url\\%s\n",
		version.Long(), url)

	_, err := s.conn.WriteToUDP(payload, clientAddr)
	if err != nil {
		s.logger.Error("Failed to send status response", "client", clientAddr, "error", err)
	}
}

func (s *Server) handleBroadcast(ctx context.Context, clientAddr *net.UDPAddr, payload []byte) {
	bc, err := broadcast.Parse(clientAddr, payload)
	if err != nil {
		s.logger.Error("Failed to parse broadcast", "error", err)
		return
	}

	sd, err := serverstatus.Query(bc.Address)
	if err != nil {
		s.logger.Error("Failed to get server status", "error", err)
		return
	}

	for _, w := range s.writers {
		go w.Write(ctx, s.logger, &writer.Data{Broadcast: bc, Server: sd})
	}
}
