package writer

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/osm/qwbs/internal/qw/broadcast"
	"github.com/osm/qwbs/internal/qw/serverstatus"
)

type Data struct {
	Broadcast *broadcast.Broadcast `json:"broadcast"`
	Server    *serverstatus.Server `json:"server"`
}

func (d *Data) MaxPlayers() string {
	if d.Broadcast.MaxPlayers != "unknown" {
		return d.Broadcast.MaxPlayers
	}

	return d.Server.MaxPlayers
}

func (d *Data) Players() string {
	if d.Broadcast.Players != "unknown" {
		return d.Broadcast.Players
	}

	return strconv.Itoa(len(d.Server.Players))
}

type Writer interface {
	Write(ctx context.Context, logger *slog.Logger, data *Data)
}
