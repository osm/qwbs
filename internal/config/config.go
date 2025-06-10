package config

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/osm/qwbs/internal/writer"
	"github.com/osm/qwbs/internal/writer/poster"
	"github.com/osm/qwbs/internal/writer/slogger"
)

type Config struct {
	Debug           bool
	ListenAddress   *net.UDPAddr
	MasterAddresses []*net.UDPAddr
	Writers         []writer.Writer
}

func FromFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	conf := &Config{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		opt := fields[0]
		args := fields[1:]
		switch opt {
		case "listen_address":
			err = conf.parseListenAddress(args)
		case "master_address":
			err = conf.parseMasterAddress(args)
		case "debug":
			err = conf.parseDebug(args)
		case "writer":
			err = conf.parseWriter(args)
		default:
			err = fmt.Errorf("unknown config option: %q", opt)
		}
		if err != nil {
			return nil, fmt.Errorf("error parsing %q: %w", opt, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if conf.ListenAddress == nil {
		return nil, fmt.Errorf("no listen address found in the configuration")
	}

	if len(conf.MasterAddresses) == 0 {
		return nil, fmt.Errorf("no master servers found in the configuration")
	}

	if len(conf.Writers) == 0 {
		return nil, fmt.Errorf("no writers found in the configuration")
	}

	return conf, nil
}

func (c *Config) parseDebug(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("debug requires exactly one argument")
	}

	v, err := strconv.ParseBool(args[0])
	if err != nil {
		return fmt.Errorf("invalid boolean value: %w", err)
	}

	c.Debug = v
	return nil
}

func (c *Config) parseListenAddress(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("listen_address requires exactly one argument")
	}

	addr, err := net.ResolveUDPAddr("udp", args[0])
	if err != nil {
		return fmt.Errorf("listen_address %q can't be resolved: %w", args[0], err)
	}

	c.ListenAddress = addr
	return nil
}

func (c *Config) parseMasterAddress(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("master_server requires exactly one argument")
	}

	addr, err := net.ResolveUDPAddr("udp", args[0])
	if err != nil {
		return fmt.Errorf("master_address %q can't be resolved: %w", args[0], err)
	}

	c.MasterAddresses = append(c.MasterAddresses, addr)
	return nil
}

func (c *Config) parseWriter(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("writer requires at least one argument")
	}

	typ := args[0]
	switch typ {
	case "slogger":
		return c.parseWriterSlogger(args)
	case "poster":
		return c.parseWriterPoster(args)
	default:
		return fmt.Errorf("unknown writer type: %q", typ)
	}
}

func (c *Config) parseWriterSlogger(args []string) error {
	var format string
	var output string

	if len(args) >= 2 {
		for _, arg := range args[1:] {
			if strings.HasPrefix(arg, "format=") {
				format = strings.TrimPrefix(arg, "format=")
			} else if strings.HasPrefix(arg, "output=") {
				output = strings.TrimPrefix(arg, "output=")
			} else {
				return fmt.Errorf("unknown slogger option: %q", arg)
			}
		}
	}

	var w io.Writer
	switch output {
	case "", "stderr":
		w = os.Stderr
	case "stdout":
		w = os.Stdout
	default:
		f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file %q: %w", output, err)
		}
		w = f
	}

	var handler slog.Handler
	switch format {
	case "", "text":
		handler = slog.NewTextHandler(w, nil)
	case "json":
		handler = slog.NewJSONHandler(w, nil)
	default:
		return fmt.Errorf("slogger format must be either text or json")
	}

	c.Writers = append(c.Writers, slogger.New(slog.New(handler)))
	return nil
}

func (c *Config) parseWriterPoster(args []string) error {
	var format poster.Format
	var url string
	var err error

	if len(args) < 3 {
		return fmt.Errorf("writer poster requires at least two arguments")
	}

	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "format=") {
			f := strings.TrimPrefix(arg, "format=")
			format, err = poster.ValidateFormat(f)
			if err != nil {
				return fmt.Errorf("unable to validate poster format %q: %w",
					f, err)
			}
		} else if strings.HasPrefix(arg, "url=") {
			url = strings.TrimPrefix(arg, "url=")
		} else {
			return fmt.Errorf("unknown poster option: %q", arg)
		}
	}

	c.Writers = append(c.Writers, poster.New(url, format))
	return nil
}
