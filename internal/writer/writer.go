package writer

import (
	"context"
	"log/slog"

	"github.com/osm/qwbs/internal/qw/broadcast"
	"github.com/osm/qwbs/internal/qw/serverstatus"
)

type Data struct {
	Broadcast *broadcast.Broadcast `json:"broadcast"`
	Server    *serverstatus.Server `json:"server"`
}

type Writer interface {
	Write(ctx context.Context, logger *slog.Logger, data *Data)
}
