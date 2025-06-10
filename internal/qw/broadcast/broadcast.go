package broadcast

import (
	"fmt"
	"net"
	"strconv"

	"github.com/osm/qwbs/internal/qw/infostring"
)

type Broadcast struct {
	Address    string `json:"address"`
	MaxPlayers string `json:"max_players"`
	Message    string `json:"message"`
	Name       string `json:"name"`
	Players    string `json:"players"`
}

func Parse(clientAddr *net.UDPAddr, payload []byte) (*Broadcast, error) {
	info, err := infostring.Parse(payload)
	if err != nil {
		return nil, fmt.Errorf("unexpected broadcast data received: %w", err)
	}

	addr, err := parseAddr(clientAddr, info)
	if err != nil {
		return nil, fmt.Errorf("unable to parse broadcast address: %w", err)
	}

	maxPlayers := infostring.Get(info, "maxplayers")
	message := infostring.Get(info, "message")
	name := infostring.Get(info, "name")
	players := infostring.Get(info, "players")

	bc := &Broadcast{
		Address:    addr,
		MaxPlayers: maxPlayers,
		Message:    message,
		Name:       name,
		Players:    players,
	}
	return bc, nil
}

func parseAddr(clientAddr *net.UDPAddr, info map[string]string) (string, error) {
	hostport := info["hostport"]
	portStr := info["port"]

	switch {
	case hostport != "":
		return hostport, nil
	case portStr != "":
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil || port == 0 {
			return "", fmt.Errorf("invalid port %q", portStr)
		}
		return fmt.Sprintf("%s:%d", clientAddr.IP.String(), uint16(port)), nil
	default:
		return clientAddr.String(), nil
	}
}
