package serverstatus

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/osm/qwbs/internal/qw/charset"
	"github.com/osm/qwbs/internal/qw/command"
	"github.com/osm/qwbs/internal/qw/infostring"
)

const (
	bufSize     = 1024 * 64
	readTimeout = time.Second
)

type Server struct {
	Map     string   `json:"map"`
	Mode    string   `json:"mode"`
	Players []Player `json:"players"`
}

type Player struct {
	Name string `json:"name"`
	Team string `json:"team"`
}

func Query(serverAddr string) (*Server, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve addr %q: %w", serverAddr, err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to perform UDP dial: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(command.GetStatusQueryBytes())
	if err != nil {
		return nil, fmt.Errorf("failed to send status query: %w", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	buf := make([]byte, bufSize)
	recvLen, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("no data received from the server: %w", err)
	}

	data := buf[:recvLen]
	header := command.GetHeaderBytes()
	if len(data) < len(header) || !bytes.Equal(data[:len(header)], header) {
		return nil, fmt.Errorf("unexpected response from server")
	}

	if recvLen <= len(header)+1 {
		return nil, fmt.Errorf("no payload data after header")
	}

	payload := data[len(header)+1:]

	infoEnd := bytes.IndexByte(payload, '\n')
	if infoEnd == -1 {
		infoEnd = len(payload)
	}

	info, err := infostring.Parse(payload[:infoEnd])
	if err != nil {
		return nil, fmt.Errorf("failed to parse serverinfo: %w", err)
	}

	mapName := infostring.Get(info, "map")
	mode := infostring.Get(info, "mode")

	var playerData []byte
	if infoEnd+1 < len(payload) {
		playerData = payload[infoEnd+1:]
	}

	players, err := parsePlayers(playerData)
	if err != nil {
		log.Printf("warning: failed to parse some players: %v", err)
	}

	return &Server{
		Map:     mapName,
		Mode:    mode,
		Players: players,
	}, nil
}

func parsePlayers(data []byte) ([]Player, error) {
	var players []Player

	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) == 0 || (len(line) == 1 && line[0] == 0x00) {
			continue
		}

		fields := parseFields(string(line))
		if len(fields) < 9 {
			return players, fmt.Errorf("malformed player line: %q", line)
		}

		players = append(players, Player{
			Name: charset.Parse(fields[4]),
			Team: charset.Parse(fields[8]),
		})
	}

	return players, nil
}

func parseFields(s string) []string {
	var fields []string
	var field strings.Builder
	inQuotes := false

	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case ' ':
			if inQuotes {
				field.WriteByte(c)
			} else if field.Len() > 0 {
				fields = append(fields, field.String())
				field.Reset()
			}
		case '"':
			inQuotes = !inQuotes
			if !inQuotes {
				fields = append(fields, field.String())
				field.Reset()
			}
		default:
			field.WriteByte(c)
		}
	}

	if field.Len() > 0 {
		fields = append(fields, field.String())
	}

	return fields
}
