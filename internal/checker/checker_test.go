package checker

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestParseBridgeLines(t *testing.T) {
	data := []byte("obfs4 1.2.3.4:443 FINGER1 cert=AAA iat-mode=0\n" +
		"notobfs4 5.6.7.8:80 BADTYPE cert=x iat-mode=0\n" +
		"obfs4 9.9.9.9:9001 FING2 cert=BBB iat-mode=2\n" +
		"garbage line without fields\n" +
		"obfs4 4.4.4.4:9100 SHORT\n")

	relays := parseBridgeLines(data, "obfs4")

	if len(relays) != 3 {
		t.Fatalf("expected 3 relays, got %d", len(relays))
	}

	if relays[0].OrAddresses[0] != "1.2.3.4:443" || relays[0].Fingerprint != "FINGER1" {
		t.Fatalf("unexpected first relay: %+v", relays[0])
	}
	if relays[0].Extra != "cert=AAA iat-mode=0" {
		t.Fatalf("unexpected extra: %q", relays[0].Extra)
	}

	if relays[2].Extra != "" {
		t.Fatalf("expected empty extra for short line, got %q", relays[2].Extra)
	}
}

func TestParseBridgeLinesTypeMismatch(t *testing.T) {
	data := []byte("snowflake 1.2.3.4:443 FING1 cert=AAA\n" +
		"obfs4 9.9.9.9:9001 FING2 cert=BBB iat-mode=2\n")

	relays := parseBridgeLines(data, "obfs4")

	if len(relays) != 1 || relays[0].Fingerprint != "FING2" {
		t.Fatalf("expected only the obfs4 relay, got %+v", relays)
	}
}

func TestBridgeLine(t *testing.T) {
	onion := Relays{Fingerprint: "FP1", OrAddresses: []string{"1.2.3.4:443"}}
	if got := onion.BridgeLine(); got != "Bridge 1.2.3.4:443 FP1" {
		t.Fatalf("unexpected onion bridge line: %q", got)
	}

	obfs := Relays{Type: "obfs4", Fingerprint: "FP2", OrAddresses: []string{"1.2.3.4:443"}, Extra: "cert=AAA iat-mode=0"}
	if got := obfs.BridgeLine(); got != "Bridge obfs4 1.2.3.4:443 FP2 cert=AAA iat-mode=0" {
		t.Fatalf("unexpected obfs4 bridge line: %q", got)
	}
}

func TestCountryFromPTR(t *testing.T) {
	cases := map[string]string{
		"host.example.de.":      "de",
		"host.example.de":       "de",
		"DE.":                   "de",
		"4.3.2.1.in-addr.arpa.": "",
		"host.example.com.":     "",
		"1.2.3.4":               "",
		"host.example.org.":     "",
	}
	for ptr, want := range cases {
		if got := countryFromPTR(ptr); got != want {
			t.Fatalf("countryFromPTR(%q) = %q, want %q", ptr, got, want)
		}
	}
}

func TestCheckMeasuresPing(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	relays := []Relays{{OrAddresses: []string{ln.Addr().String()}}}
	alive := make(chan Relays, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	go func() {
		conn, err := ln.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	Check(ctx, alive, relays, time.Second, "")

	select {
	case r := <-alive:
		if r.Ping <= 0 {
			t.Fatalf("expected ping > 0, got %v", r.Ping)
		}
	case <-time.After(time.Second * 2):
		t.Fatal("no relay received")
	}
}

func TestCheckFiltersByCountryBeforeDial(t *testing.T) {
	original := lookupPTR
	defer func() { lookupPTR = original }()

	lookupPTR = func(host string) string {
		if host == "203.0.113.1" {
			return "de"
		}
		return "ru"
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	accepted := make(chan string, 2)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			accepted <- conn.RemoteAddr().String()
			conn.Close()
		}
	}()

	matchingAddr := ln.Addr().String()
	relays := []Relays{
		{OrAddresses: []string{"203.0.113.1:443"}},
		{OrAddresses: []string{matchingAddr}},
	}
	alive := make(chan Relays, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	Check(ctx, alive, relays, time.Second, "ru")

	select {
	case r := <-alive:
		if r.OrAddresses[0] != matchingAddr || r.Country != "ru" {
			t.Fatalf("expected the matching relay, got %+v", r)
		}
	case <-time.After(time.Second * 2):
		t.Fatal("matching relay not received")
	}

	select {
	case <-accepted:
	case <-time.After(time.Second * 2):
		t.Fatal("expected one dial to the matching relay")
	}

	select {
	case <-accepted:
		t.Fatal("dialed a relay that does not match --country")
	case <-time.After(300 * time.Millisecond):
	}
}
