package checker

import (
	"context"
	"net"
	"slices"
	"strings"
	"sync"
	"time"
)

var (
	countryCache sync.Map
	lookupPTR    = lookupCountry
)

func Check(ctx context.Context, alive chan Relays, relays []Relays, timeout time.Duration, country string, forceLookup bool) {
	countries := strings.Split(country, ",")
	needFilter := countries[0] != ""

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	for _, relay := range relays {
		select {
		case <-ctx.Done():
			return
		default:
			if needFilter || forceLookup {
				if relay.Country == "" {
					relay.Country = relayCountry(ctx, relay)
				}
			}
			if needFilter && !slices.Contains(countries, relay.Country) {
				continue
			}
			start := time.Now()
			conn, err := dialer.Dial("tcp", relay.OrAddresses[0])
			if err == nil {
				relay.Ping = time.Since(start)
				conn.Close()

				select {
				case alive <- relay:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func relayCountry(ctx context.Context, relay Relays) string {
	if relay.Country != "" {
		return relay.Country
	}

	host, _, err := net.SplitHostPort(relay.OrAddresses[0])
	if err != nil {
		return ""
	}

	if cached, ok := countryCache.Load(host); ok {
		return cached.(string)
	}

	country := lookupPTR(ctx, host)
	countryCache.Store(host, country)
	return country
}

func lookupCountry(ctx context.Context, host string) string {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	names, err := (&net.Resolver{}).LookupAddr(ctx, host)
	if err != nil || len(names) == 0 {
		return ""
	}

	return countryFromPTR(names[0])
}

func countryFromPTR(ptr string) string {
	last := strings.ToLower(strings.Trim(ptr, "."))
	if i := strings.LastIndex(last, "."); i >= 0 {
		last = last[i+1:]
	}

	if len(last) != 2 {
		return ""
	}
	for _, c := range last {
		if c < 'a' || c > 'z' {
			return ""
		}
	}
	return last
}
