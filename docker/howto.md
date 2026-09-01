# ⚙️ Сборка и запуск Docker-контейнера с Автоматическим Поиском Мостом

---

## 🔨 Сборка образа

Для начала следует перейти в **папку с самим проектом**. Только после начать сборку

```bash
cd ../
docker build -t tor-scanner -f docker/Dockerfile .
```

---

## 🚀 Запуск контейнера

```bash
docker run -p 9050:9050 --rm -d --name tor-node tor-scanner
```

Вы также можете передавать параметры

```bash
docker run -p 9050:9050 --rm -d --name tor-node tor-scanner -count 3 -ping 100 -country ru
```

После запуска станет доступна SOCKS5 прокси на `localhost:9050`. Внутренний порт Tor фиксирован и равен `9050`; наружу он пробрасывается параметром `-p <внешний>:9050`.

### 🔧 Параметры через переменные окружения

| Переменная | Описание | Значение по умолчанию |
|---|---|---|
| `EXIT_NODES` | Ограничить выходные узлы. Пример: `{ru},{de}` | не задано |
| `EXCLUDE_EXIT_NODES` | Исключить выходные узлы. Пример: `{us},{gb}` | не задано |
| `STRICT_NODES` | Строгий выбор узлов. Пример: `1` | не задано |
| `CLIENT_USE_IPV6` | Использовать IPv6 для выходных узлов. Пример: `1` | не задано |

```bash
docker run -p 9050:9050 --rm -d --name tor-node \
    -e EXIT_NODES='{ru}' \
    -e STRICT_NODES='1' \
    tor-scanner -count 3 -ping 100
```

> ⚙️ `SocksPort` фиксирован (`0.0.0.0:9050`), а `ClientTransportPlugin` для obfs4 (lyrebird) жёстко прописан в подключаемом конфиге. Если хотите использовать другой внутренний порт — укажите его во второй части `-p`, например `-p 1080:9050`.

### ✅ Проверка работы

```bash
curl -x socks5h://localhost:9050 https://ifconfig.co
```

---

## ❓ Очень долго запускается?

Даже с работающими мостами запуск Tor модет занять до минуты.
Проверим логи нашего контейнера

```bash
docker logs tor-node
```

При успешном запуске вывод должен быть примерно такой:

```
Nov 16 09:31:15.000 [notice] Bootstrapped 56% (loading_descriptors): Loading relay descriptors
Nov 16 09:31:15.000 [notice] Bootstrapped 62% (loading_descriptors): Loading relay descriptors
Nov 16 09:31:16.000 [notice] Bootstrapped 67% (loading_descriptors): Loading relay descriptors
Nov 16 09:31:16.000 [notice] Bootstrapped 75% (enough_dirinfo): Loaded enough directory info to build circuits
Nov 16 09:31:17.000 [notice] Bootstrapped 90% (ap_handshake_done): Handshake finished with a relay to build circuits
Nov 16 09:31:17.000 [notice] Bootstrapped 95% (circuit_create): Establishing a Tor circuit
Nov 16 09:31:18.000 [notice] Bootstrapped 100% (done): Done
```

> ✅ Убедитесь, что присутствует строка `Bootstrapped 100% (done): Done` - это успешный запуск Tor

### 🔁 Перезапускаем если что-то пошло не так

```bash
docker stop tor-node
docker run -p 9050:9050 --rm -d --name tor-node tor-scanner
```
