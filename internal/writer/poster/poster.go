package poster

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/osm/qwbs/internal/writer"
)

const (
	contentTypeJSON = "application/json"
	contentTypeText = "text/plain"
	timeout         = time.Second * 10
)

type Poster struct {
	url    string
	format Format
	client *http.Client
}

func New(url string, format Format) *Poster {
	return &Poster{
		url:    url,
		format: format,
		client: &http.Client{Timeout: timeout},
	}
}

func (p *Poster) Write(ctx context.Context, logger *slog.Logger, data *writer.Data) {
	body, contentType, err := format(p.format, data)
	if err != nil {
		logger.Error("Failed to format data", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.url, body)
	if err != nil {
		logger.Error("Failed to create HTTP request", "url", p.url, "error", err)
		return
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := p.client.Do(req)
	if err != nil {
		logger.Error("Failed to perform HTTP request", "url", p.url, "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Error("Received an unexpected response", "url", p.url, "status", resp.Status)
	}

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		logger.Error("Failed to drain HTTP response body", "url", p.url, "error", err)
	}
}
