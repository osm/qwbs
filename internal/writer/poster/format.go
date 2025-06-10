package poster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/osm/qwbs/internal/writer"
)

type Format uint8

const (
	Unknown Format = iota
	Discord
	JSON
	Text
)

var formatMap = map[string]Format{
	"discord": Discord,
	"json":    JSON,
	"text":    Text,
}

func ValidateFormat(format string) (Format, error) {
	f, ok := formatMap[format]
	if !ok {
		return Unknown, fmt.Errorf("unknown poster format %q", format)
	}

	return f, nil
}

func format(format Format, data *writer.Data) (io.Reader, string, error) {
	switch format {
	case Discord:
		return formatDiscord(data)
	case JSON:
		return formatJSON(data)
	case Text:
		return formatText(data)
	}

	return nil, "", fmt.Errorf("unknown format: %q", format)
}

func formatJSON(data *writer.Data) (io.Reader, string, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, "", err
	}

	return bytes.NewBuffer(payload), contentTypeJSON, nil
}

func formatText(data *writer.Data) (io.Reader, string, error) {
	bc := data.Broadcast
	payload := fmt.Sprintf("> %s [%s/%s] %s: %s",
		bc.Address, bc.Players, bc.MaxPlayers, bc.Name, bc.Message)

	return bytes.NewBufferString(payload), contentTypeText, nil
}
