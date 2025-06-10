package infostring

import (
	"fmt"
	"strings"

	"github.com/osm/qwbs/internal/qw/charset"
)

func Parse(data []byte) (map[string]string, error) {
	parts := strings.Split(strings.TrimSpace(string(data)), `\`)
	if parts[0] == "" {
		parts = parts[1:]
	}

	if len(parts)%2 != 0 {
		return nil, fmt.Errorf("broken info string: unbalanced key-value pairs")
	}

	kv := make(map[string]string)
	for i := 0; i < len(parts); i += 2 {
		kv[parts[i]] = parts[i+1]
	}

	return kv, nil
}

func Get(info map[string]string, key string) string {
	value, ok := info[key]
	if !ok {
		return "unknown"
	}

	return charset.Parse(value)
}
