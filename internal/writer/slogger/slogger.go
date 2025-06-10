package slogger

import (
	"context"
	"log/slog"
	"reflect"
	"strings"

	"github.com/osm/qwbs/internal/writer"
)

type Slogger struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Slogger {
	return &Slogger{logger: logger}
}

func (s *Slogger) Write(_ context.Context, _ *slog.Logger, data *writer.Data) {
	var fields []any
	val := reflect.ValueOf(data.Broadcast).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		k := strings.ToLower(typ.Field(i).Name)
		v := val.Field(i).Interface()
		fields = append(fields, k, v)
	}

	s.logger.Info("Broadcast received", fields...)
}
