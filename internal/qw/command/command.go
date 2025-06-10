package command

import (
	"bytes"
	"fmt"
	"strconv"
)

type Command uint8

const (
	Unknown Command = iota
	ACK
	Broadcast
	GetChallenge
	Ping
	Status
)

var (
	ack          = []byte{0x6c}
	broadcast    = []byte("broadcast")
	getChallenge = []byte("getchallenge\n")
	header       = []byte{0xff, 0xff, 0xff, 0xff}
	heartbeat    = []byte{0x61, 0x0a}
	ping         = []byte{0x6b, 0x0a}
	print        = []byte{0x6e}
	shutdown     = []byte{0x43, 0x0a}
	status       = []byte("status")
	statusQuery  = []byte("status 19")
)

func Parse(buf []byte) (Command, []byte) {
	headerLen := len(header)
	if len(buf) < headerLen || !bytes.Equal(buf[:headerLen], header) {
		return Unknown, nil
	}

	payload := buf[headerLen:]
	switch {
	case len(payload) >= len(getChallenge) && bytes.HasPrefix(payload, getChallenge):
		return GetChallenge, nil
	case len(payload) >= len(broadcast) && bytes.HasPrefix(payload, broadcast):
		return Broadcast, payload[len(broadcast):]
	case len(payload) >= len(status) && bytes.HasPrefix(payload, status):
		return Status, nil
	case len(payload) >= len(ping) && bytes.HasPrefix(payload, ping):
		return Ping, nil
	case len(payload) >= len(ack) && bytes.HasPrefix(payload, ack):
		return ACK, nil
	default:
		return Unknown, nil
	}
}

func GetHeaderBytes() []byte {
	return header
}

func GetPingBytes() []byte {
	return ping
}

func GetShutdownBytes() []byte {
	return shutdown
}

func GetACKBytes() []byte {
	return ack
}

func GetHeartbeatBytes(sequence int) []byte {
	var buf []byte

	buf = append(buf, heartbeat...)
	buf = strconv.AppendInt(buf, int64(sequence), 10)
	buf = append(buf, 0x0a)
	buf = strconv.AppendInt(buf, int64(0), 10)
	buf = append(buf, 0x0a)

	return buf
}

func GetPrintBytes(format string, args ...any) []byte {
	var buf []byte

	buf = append(buf, header...)
	buf = append(buf, print...)
	buf = append(buf, []byte(fmt.Sprintf(format, args...))...)

	return buf
}

func GetStatusQueryBytes() []byte {
	var buf []byte

	buf = append(buf, header...)
	buf = append(buf, statusQuery...)

	return buf
}
