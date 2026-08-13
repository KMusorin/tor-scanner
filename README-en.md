# Tor-scanner

**Tor-scanner** is a fast and simple tool to find working bridges for Tor.
It checks bridges for availability and selects only those that respond quickly and reliably.

Two types of sources are supported:

- **Regular relays (vanilla)** — Onionoo list in JSON format (default source).
- **obfs4 bridges** — line-based list from the [Tor-Bridges-Collector](https://github.com/scriptzteam/Tor-Bridges-Collector).

Suitable for users living in censored regions who need reliable bridges to connect to Tor.

---

![Demo](.github/demo.gif)

## 🚀 How to Use

Run the application with the desired parameters:

```bash
tor-scanner \
  --count 10 \
  --threads 20 \
  --ping 1000 \
  --country ru,es,us \
  --output tor.conf \
  --input bridges.list
```

### Parameters:

| Parameter | Description |
|----------|-------------|
| `--count` | Number of working bridges to find (default: 5) |
| `--threads` | Number of concurrent requests (default: 50) — higher values mean faster scanning |
| `--ping` | Maximum delay (in milliseconds) for a bridge to be considered responsive (default: 250) |
| `--country` | Filter bridges by country codes. Example: `ru,es,us` — checks only bridges from these countries. For obfs4 the country is resolved via reverse DNS, so IPs without a PTR record won't match |
| `--output` | File to save found bridges (e.g., `tor.conf`) |
| `--input` | URL or path to a file containing a list of bridges. If omitted, the source is chosen automatically by `--type`: vanilla — tor-onionoo-mirror, obfs4 — `bridges-obfs4` |
| `--type` | Bridge type: `obfs4` — use the obfs4 bridge list; empty value (default) — regular relays (vanilla) |

---

## 💡 Examples

### Find 15 fast bridges from Russia and USA, save to file:

```bash
tor-scanner --count 15 --threads 100 --ping 100 --country ru,us --output tor.conf
```

### Check a list of bridges and save only working ones:

```bash
tor-scanner --input my-bridges.list --output working-bridges.txt --ping 200
```

or

```bash
tor-scanner --input "https://onionoo.torproject.org/details?type=relay&running=true&fields=fingerprint,or_addresses,country" --output working-bridges.txt --ping 200
```

### Find working obfs4 bridges:

```bash
tor-scanner --type obfs4 --count 10 --threads 20 --ping 1000 --output tor.conf
```

Country filtering for obfs4 bridges works via reverse DNS (top-level domain): `--country ru,de`.

### Use your own obfs4 bridge list:

```bash
tor-scanner --type obfs4 --input my-obfs4-bridges.list --count 10 --output tor.conf
```

---

## 📄 Output

The tool generates a ready-to-use `torrc` config:

- vanilla: `Bridge <ip:port> <fingerprint>`
- obfs4: `Bridge obfs4 <ip:port> <fingerprint> cert=... iat-mode=...`
- ends with `UseBridges 1`

If fewer bridges are found than requested via `--count`, a warning is printed and the found bridges are still saved to the file. If no working bridges are found at all, the tool exits with an error.

---

## 🛠️ How to Build

Make sure you have Go (1.25+) installed:

```bash
git clone https://github.com/KMusorin/tor-scanner.git
cd tor-scanner
go build -o tor-scanner cmd/main.go
```

Done! Now you can run `./tor-scanner`.

[🐳 Using Docker](docker/howto-en.md)
