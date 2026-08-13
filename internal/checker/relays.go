package checker

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type RelaysInfo struct {
	Relays []Relays `json:"relays"`
}

type Relays struct {
	Fingerprint string   `json:"fingerprint"`
	OrAddresses []string `json:"or_addresses"`
	Country     string   `json:"country"`
	Type        string
	Extra       string
	Ping        time.Duration
}

func (r Relays) BridgeLine() string {
	line := "Bridge " + r.OrAddresses[0] + " " + r.Fingerprint
	if r.Type != "" {
		line = "Bridge " + r.Type + " " + r.OrAddresses[0] + " " + r.Fingerprint
		if r.Extra != "" {
			line += " " + r.Extra
		}
	}
	return line
}

func GetRelays(input string) ([]Relays, error) {
	return GetBridges(input, "")
}

func GetBridges(input, bridgeType string) ([]Relays, error) {
	data, err := readInput(input)
	if err != nil {
		return nil, err
	}

	if bridgeType == "" {
		var info RelaysInfo
		json.Unmarshal(data, &info)
		return info.Relays, nil
	}

	return parseBridgeLines(data, bridgeType), nil
}

func parseBridgeLines(data []byte, bridgeType string) []Relays {
	var relays []Relays
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[0] != bridgeType {
			continue
		}
		relays = append(relays, Relays{
			Fingerprint: fields[2],
			OrAddresses: []string{fields[1]},
			Type:        bridgeType,
			Extra:       strings.Join(fields[3:], " "),
		})
	}
	return relays
}

func readInput(input string) ([]byte, error) {
	if strings.HasPrefix(input, "https://") {
		client := &http.Client{
			Timeout: time.Second * 15,
		}

		resp, err := client.Get(input)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		return io.ReadAll(resp.Body)
	}

	return os.ReadFile(input)
}
