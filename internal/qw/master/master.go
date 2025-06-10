package master

import (
	"context"
	"log/slog"
	"net"
	"sync/atomic"
	"time"

	"github.com/osm/qwbs/internal/qw/command"
)

const (
	heartbeatInterval = time.Second * 300
	registerInterval  = time.Second * 60
)

type Master struct {
	conn         *net.UDPConn
	addr         *net.UDPAddr
	logger       *slog.Logger
	isRunning    int32
	isRegistered int32
	lastRegister atomic.Value
	sequence     int
}

func New(conn *net.UDPConn, addr *net.UDPAddr, logger *slog.Logger) *Master {
	return &Master{
		conn:   conn,
		addr:   addr,
		logger: logger,
	}
}

func (m *Master) Register() error {
	if atomic.LoadInt32(&m.isRegistered) == 1 {
		return nil
	}

	lastAny := m.lastRegister.Load()
	if lastAny != nil {
		if last, ok := lastAny.(time.Time); ok && time.Since(last) < registerInterval {
			return nil
		}
	}

	if _, err := m.conn.WriteToUDP(command.GetPingBytes(), m.addr); err != nil {
		return err
	}

	m.lastRegister.Store(time.Now())
	m.logger.Info("Registering with master server", "master", m.addr)
	return nil
}

func (m *Master) Unregister() error {
	if atomic.LoadInt32(&m.isRegistered) == 0 {
		return nil
	}

	if _, err := m.conn.WriteToUDP(command.GetShutdownBytes(), m.addr); err != nil {
		return err
	}

	atomic.StoreInt32(&m.isRegistered, 0)
	m.logger.Info("Unregistering from master server", "master", m.addr)
	return nil
}

func (m *Master) Heartbeat(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&m.isRunning, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&m.isRunning, 0)

	atomic.StoreInt32(&m.isRegistered, 1)

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		m.logger.Info("Sending heartbeat to master server",
			"addr", m.addr,
			"sequence", m.sequence,
		)

		_, err := m.conn.WriteToUDP(command.GetHeartbeatBytes(m.sequence), m.addr)
		if err != nil {
			m.logger.Error("Failed to send heartbeat to master server",
				"master", m.addr, "sequence", m.sequence, "error", err)
			continue
		}

		m.sequence++

		select {
		case <-ticker.C:
		case <-ctx.Done():
			if err := m.Unregister(); err != nil {
				m.logger.Error("Failed to send shutdown packet to master server",
					"master", m.addr, "error", err)
			}
			return
		}
	}
}
