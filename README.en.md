# Cash Register Control

A cash-flow control system I built to replace a daily closing spreadsheet used across a
network of laboratory units. The goal was simple: get the team off Excel and give them a
fast way to record, double-check and track the day's numbers — while being easy to roll
out to several units, with no install and even running straight from a USB stick.

It ended up as a single app that runs in the browser **and** as a Windows program
(`.exe`), storing its data in a file right next to the executable.

> The data shipped with the project is **fictional**, for demo purposes only.

## Features

- **Daily entry per attendant**, splitting Pix, Cash, Card and External Collection. Every
  sale is stored as its own item (not just the total), which makes reconciliation easy.
- **Attendant registry** with autocomplete on entry (you can also type a one-off name when
  someone is covering the register, without registering them).
- **Private payment sources (tables/insurance plans)**: searchable registry, per-sale
  table tagging, optional discount, and **split payments** (when a customer pays part by
  card and part in cash, for example — it adds up as a single sale and counts the table
  only once).
- **Reports and rankings** by day, month, quarter, semester and year: top sellers,
  most-used tables, sales share, biggest sale of the day, discounts granted, total
  registrations, and more.
- **Interactive charts** (bars, pie, ranking) with hover tooltips.
- **Daily notes.**
- **One-click backup and restore.**

## Tech stack

- **HTML, CSS and JavaScript** (no framework, no external dependencies — it's a single file).
- **Go** for the launcher that turns the app into a Windows executable (local `localhost`
  server, data written to a JSON file next to the `.exe`).
- **Hand-written SVG** for the charts.
- `localStorage` + JSON file for persistence, depending on the usage mode.

## Running it

### Web version (quickest to try)
Just open `controle-caixa.html` in your browser (Chrome or Edge). Works offline.

### Windows program (.exe)
The executable is built from the `launcher/` folder (Go code):

```bash
cd launcher
# requires Go (https://go.dev)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-H windowsgui -s -w" -o ControleDeCaixa.exe .
```

When opened, the `.exe` starts a local server, shows the app in a window and saves
everything to `ControleDeCaixa-dados.json` in the same folder — that's what makes it
portable on a USB stick.

## Project structure

```
controle-caixa.html      -> the full app (UI + logic), web/demo version
launcher/
  main.go                -> Go launcher: serves the app and writes data to a file
  index.html             -> the app embedded in the executable (no data, program mode)
  versioninfo.json       -> .exe metadata/icon
docs/                    -> screenshots
```

## A few technical decisions (and why)

- **Single file, no framework.** Since the app had to run on a range of machines and
  sometimes from a USB stick, I avoided heavy builds and dependencies. A single HTML loads
  fast and never "breaks for a missing package".
- **Turning it into an `.exe` with a Go launcher.** Instead of bundling a whole browser
  (heavy), the launcher just starts a local server and opens the app in an Edge/Chrome
  app window. The final `.exe` stays small (~5 MB) and the data lives in a file next to it
  — which solved USB portability.
- **Each sale as an individual item.** Storing every value separately (instead of just the
  total) made the rankings, per-sale table tagging and split payments possible without
  rework later.
- **"No-scare" saving.** It auto-saves while you work, has a save button with on-screen
  confirmation, and saves again when the window closes, so no entry is lost.

## Author

**Sandro de Lima Pereira**

## License

MIT — see [LICENSE](LICENSE).
