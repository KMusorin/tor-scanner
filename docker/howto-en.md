# ⚙️ Build and Run Docker Image with Automatic Tor Bridge Discovery

---

## 🔨 Build the Docker Image

First, navigate to the **project root directory** (one level above `docker/`). Then build:

```bash
cd ../
docker build -t tor-scanner -f docker/Dockerfile .
```

---

## 🚀 Run the Container

```bash
docker run -p 9050:9050 --rm -d --name tor-node tor-scanner
```

You can also pass parameters

```bash
docker run -p 9050:9050 --rm -d --name tor-node tor-scanner -count 3 -ping 100 -country us
```

Once running, Tor will be available as a SOCKS5 proxy at `localhost:9050`. The internal Tor port is fixed at `9050`; it is exposed via the `-p <external>:9050` mapping.

### 🔧 Environment variables

| Variable | Description | Default |
|---|---|---|
| `EXIT_NODES` | Restrict exit nodes, e.g. `{us},{de}` | unset |
| `EXCLUDE_EXIT_NODES` | Exclude exit nodes, e.g. `{cn},{ru}` | unset |
| `STRICT_NODES` | Strict node selection, e.g. `1` | unset |
| `CLIENT_USE_IPV6` | Use IPv6 for exit nodes, e.g. `1` | unset |

```bash
docker run -p 9050:9050 --rm -d --name tor-node \
    -e EXIT_NODES='{us}' \
    -e STRICT_NODES='1' \
    tor-scanner -count 3 -ping 100
```

> ⚙️ `SocksPort` is fixed (`0.0.0.0:9050`) and the `ClientTransportPlugin` for obfs4 (lyrebird) is hard-coded in the included config. To use a different internal port, set it in the second part of `-p`, e.g. `-p 1080:9050`.

### ✅ Verify the proxy is working:

```bash
curl -x socks5h://localhost:9050 https://ifconfig.co
```

---

## ❓ Tor Takes Too Long to Start?

Even with working bridges, Tor may take up to a minute to fully bootstrap.
Check the logs to monitor progress:

```bash
docker logs tor-node
```

A successful startup will show lines like:

```
Nov 16 09:31:15.000 [notice] Bootstrapped 56% (loading_descriptors): Loading relay descriptors
Nov 16 09:31:15.000 [notice] Bootstrapped 62% (loading_descriptors): Loading relay descriptors
Nov 16 09:31:16.000 [notice] Bootstrapped 67% (loading_descriptors): Loading relay descriptors
Nov 16 09:31:16.000 [notice] Bootstrapped 75% (enough_dirinfo): Loaded enough directory info to build circuits
Nov 16 09:31:17.000 [notice] Bootstrapped 90% (ap_handshake_done): Handshake finished with a relay to build circuits
Nov 16 09:31:17.000 [notice] Bootstrapped 95% (circuit_create): Establishing a Tor circuit
Nov 16 09:31:18.000 [notice] Bootstrapped 100% (done): Done
```

> ✅ Look for: `Bootstrapped 100% (done): Done` — this means Tor is ready.

### 🔁 If stuck, restart:

```bash
docker stop tor-node
docker run -p 9050:9050 --rm -d --name tor-node tor-scanner
```

---

## ⚠️ If the scan finds no bridges

`tor-scanner` exits with code `1` when a scan yields nothing (for example, the router's network is not ready yet right after boot). In that case `entrypoint.sh` does not abort the startup:

- if `/etc/tor/torrc.d/bridges.conf` already holds a list from a previous run, Tor starts on it (a stale but working list);
- if the file is missing (very first run), the container exits with code `1` and `restart-policy=always` restarts it (with `restart-interval=1m`, once a minute) until a scan succeeds.

Note: with `root-dir` mounted on persistent storage, `bridges.conf` survives restarts, so the previous list can be used as a fallback.
