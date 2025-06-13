package poster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/osm/qwbs/internal/qw/serverstatus"
	"github.com/osm/qwbs/internal/writer"
)

type DiscordPayload struct {
	Content string         `json:"content"`
	Embeds  []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func formatDiscord(data *writer.Data) (io.Reader, string, error) {
	bc := data.Broadcast
	sv := data.Server
	pl := data.Server.Players

	playerNames := func(players []serverstatus.Player) string {
		n := len(players)
		switch n {
		case 0:
			return ""
		case 1:
			return players[0].Name
		case 2:
			return players[0].Name + " and " + players[1].Name
		default:
			names := make([]string, n)
			for i, p := range players {
				names[i] = p.Name
			}
			return strings.Join(names[:n-1], ", ") + " and " + names[n-1]
		}
	}

	payload := DiscordPayload{
		Content: fmt.Sprintf("**%s**: %s", bc.Name, bc.Message),
		Embeds: []DiscordEmbed{
			{
				Title: fmt.Sprintf("%s/%s @ %s | %s",
					data.Players(), data.MaxPlayers(), sv.Map, bc.Address),
				Description: playerNames(pl),
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, "", err
	}

	return bytes.NewBuffer(jsonData), contentTypeJSON, nil
}
