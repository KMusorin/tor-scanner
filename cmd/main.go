package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/leaanthony/clir"
	"github.com/KMusorin/tor-scanner/internal/checker"

	"github.com/logrusorgru/aurora/v4"
)

var (
	count   = 5
	threads = 50
	maxPing = 250

	output     = ""
	input      = ""
	bridgeType = ""
	country    = ""

	list = []checker.Relays{}
)

const (
	defaultInputOnionoo = "https://raw.githubusercontent.com/ValdikSS/tor-onionoo-mirror/refs/heads/master/details-running-relays-fingerprint-address-only.json"
	defaultInputObfs4   = "https://raw.githubusercontent.com/scriptzteam/Tor-Bridges-Collector/refs/heads/main/bridges-obfs4"
)

func main() {
	cli := clir.NewCli("Tor-scanner", "Easily Find Working Bridges for Tor", "v0.0.3")

	cli.IntFlag("count", "How many relays are needed", &count)
	cli.IntFlag("threads", "How many threads are needed", &threads)
	cli.IntFlag("ping", "Maximum Delay to Relay in ms", &maxPing)
	cli.StringFlag("country", "Including countries. EX: ru,es,us", &country)
	cli.StringFlag("output", "File to save", &output)
	cli.StringFlag("input", "You can set where to get bridges from (including the path to the file)", &input)
	cli.StringFlag("type", "Bridge type (obfs4, etc). Empty for the default relay list", &bridgeType)

	cli.Action(start)

	if err := cli.Run(); err != nil {
		fmt.Printf("Error encountered: %v\n", err)
	}
}

func start() error {
	if input == "" {
		if bridgeType == "" {
			input = defaultInputOnionoo
		} else {
			input = defaultInputObfs4
		}
	}

	bridgeLabel := bridgeType
	if bridgeLabel == "" {
		bridgeLabel = "vanilla"
	}

	fmt.Println(aurora.Cyan(fmt.Sprintf("Parsing %s bridges from %s...", bridgeLabel, input)))
	relays, err := checker.GetBridges(input, bridgeType)
	if err != nil {
		return err
	}
	if len(relays) == 0 {
		return fmt.Errorf("no relays found in the input")
	}
	workers := threads
	if workers > len(relays) {
		workers = len(relays)
	}
	if workers < 1 {
		workers = 1
	}
	fmt.Println(aurora.Green(fmt.Sprintf("Got %d relays", len(relays))))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	alive := make(chan checker.Relays)
	var wg sync.WaitGroup
	base := len(relays) / workers
	extra := len(relays) % workers
	for i := 0; i < workers; i++ {
		start := i*base + min(i, extra)
		end := start + base
		if i < extra {
			end++
		}
		chunk := relays[start:end]

		wg.Add(1)
		go func(c []checker.Relays) {
			defer wg.Done()
			checker.Check(ctx, alive, c, time.Duration(maxPing)*time.Millisecond, country)
		}(chunk)
	}
	go func() {
		wg.Wait()
		close(alive)
	}()

	for len(list) < count {
		addr, ok := <-alive
		if !ok {
			break
		}
		list = append(list, addr)
		label := fmt.Sprintf("%dms", addr.Ping.Milliseconds())
		if addr.Country != "" {
			label = strings.ToUpper(addr.Country) + " " + label
		}
		fmt.Println(aurora.Green("Found"), aurora.Cyan(addr.OrAddresses[0]), aurora.Blue("- "+label))
	}
	cancel()

	if len(list) == 0 {
		return fmt.Errorf("no working bridges found")
	}
	if len(list) < count {
		fmt.Println(aurora.Red(fmt.Sprintf("Only %d of %d requested bridges were found", len(list), count)))
	}

	fmt.Println("\n\nYour relays:")

	var text string
	for _, relay := range list {
		fmt.Println(relay.BridgeLine())
		text += relay.BridgeLine() + "\n"
	}
	fmt.Println("UseBridges 1")
	text += "UseBridges 1\n"

	if output != "" {
		fmt.Println()
		if err := os.WriteFile(output, []byte(text), 0644); err != nil {
			return err
		}

		fmt.Println(aurora.Green("Successfully written to a file"), aurora.Underline(output))
	}

	return nil
}
