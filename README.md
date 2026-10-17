# Tor-scanner

[English](README-en.md)

**Tor-scanner** — это быстрый и простой инструмент для поиска рабочих мостов (bridges) для Tor.
Он проверяет мосты на доступность и отбирает только те, которые отвечают быстро и стабильно.

Поддерживается два типа источников:

- **Обычные релеи (vanilla)** — список Onionoo в формате JSON (источник по умолчанию).
- **obfs4-мосты** — линейный список из коллектора [Tor-Bridges-Collector](https://github.com/Delta-Kronecker/Tor-Bridges-Collector) или [Tor-Bridges-Collector-v2](https://github.com/scriptzteam/Tor-Bridges-Collector-v2).

Подходит для тех, кто живёт в странах с цензурой и нуждается в надёжных мостах для подключения к Tor.

---

## 🚀 Как использовать

Запустите приложение с нужными параметрами:

```bash
tor-scanner \
  --count 10 \
  --threads 20 \
  --ping 1000 \
  --country ru,es,us \
  --output tor.conf \
  --input bridges.list
```

### Параметры:

| Параметр | Описание |
|----------|----------|
| `--type` | Тип мостов: `obfs4` — использовать список obfs4-мостов; пустое значение (по умолчанию) — обычные релеи (vanilla) |
| `--count` | Сколько рабочих мостов нужно найти (по умолчанию: 5) |
| `--threads` | Сколько запросов делать одновременно (по умолчанию: 50) |
| `--ping` | Максимальная задержка (в миллисекундах), чтобы мост считался рабочим (по умолчанию: 250) |
| `--country` | Фильтр по странам. Пример: `ru,es,us` — будут проверяться только мосты из этих стран. Для obfs4 страна определяется по reverse DNS, поэтому IP без PTR-записи в результат не попадёт |
| `--lookup-country` | Принудительно определять страну по reverse DNS для мостов, у которых источник не указывает страну (например, obfs4). Код страны выводится в консоли рядом с найденным мостом. Без фильтра `--country` ничего не отбрасывает. Внимание: на каждый уникальный IP без страны выполняется PTR-запрос (таймаут 2 с), что замедляет сканирование; IP без PTR-записи останется без кода страны |
| `--output` | Имя файла, куда сохранить найденные мосты (например: `tor.conf`) |
| `--input` | Ссылка или путь к файлу со списком мостов. Если не указан, источник выбирается автоматически по `--type`: vanilla — tor-onionoo-mirror, obfs4 — `bridges-obfs4` |

---

## 💡 Примеры

### Найти 15 быстрых мостов из России и США, сохранить в файл:

```bash
tor-scanner --count 15 --threads 100 --ping 100 --country ru,us --output tor.conf
```

### Проверить список мостов и сохранить только рабочие:

```bash
tor-scanner --input my-bridges.list --output working-bridges.txt --ping 200
```

или

```bash
tor-scanner --input "https://onionoo.torproject.org/details?type=relay&running=true&fields=fingerprint,or_addresses,country" --output working-bridges.txt --ping 200
```

### Найти рабочие obfs4-мосты:

```bash
tor-scanner --type obfs4 --count 10 --threads 20 --ping 1000 --output tor.conf
```

Фильтр по странам для obfs4-мостов работает через reverse DNS (домен верхнего уровня): `--country ru,de`.

### Использовать свой список obfs4-мостов:

```bash
tor-scanner --type obfs4 --input my-obfs4-bridges.list --count 10 --output tor.conf
```

### Найти рабочие obfs4-мосты и показать их страны:

```bash
tor-scanner --type obfs4 --count 10 --threads 20 --ping 1000 --lookup-country --output tor.conf
```

`--lookup-country` определит страну каждого найденного моста по reverse DNS и выведет код страны (например, `DE 45ms`) рядом с мостом в консоли.

---

## 📄 Что на выходе

Программа формирует готовый к использованию конфиг для `torrc`:

- vanilla: `Bridge <ip:port> <fingerprint>`
- obfs4: `Bridge obfs4 <ip:port> <fingerprint> cert=... iat-mode=...`
- в конце добавляется строка `UseBridges 1`

Если найдено меньше мостов, чем запрошено через `--count`, в консоль выводится предупреждение, а найденные мосты всё равно сохраняются в файл. При полном отсутствии рабочих мостов программа завершится с кодом `1`. В Docker-образе это используется для фолбэка: при неудачном сканировании Tor поднимается на прежнем `bridges.conf`, а если его нет — контейнер перезапускается (`restart-policy=always`).

---

## 🛠️ Как собрать

Убедитесь, что у вас установлен Go (1.25+):

```bash
git clone https://github.com/KMusorin/tor-scanner.git
cd tor-scanner
go build -o tor-scanner cmd/main.go
```

Готово! Теперь можно запускать `./tor-scanner`.

[🐳 Использование Docker](docker/howto.md)
