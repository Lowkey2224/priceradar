#Priceradar
Das **"PriceRadar"-Gesamtsystem** ist ein fantastisches Lernprojekt. Indem wir Crawler, API und CLI in einem System kombinieren, baust du im Grunde eine moderne **Microservice-Architektur in Go**.

Hier ist ein strukturierter Plan, wie du das System Schritt für Schritt aufbaust, ohne von der Komplexität erschlagen zu werden.

---

## 🏗️ System-Architektur im Überblick

Das System besteht aus 3 Hauptkomponenten, die über eine gemeinsame Datenbank (z. B. PostgreSQL) miteinander kommunizieren:

1. **`priceradar-worker` (Crawler/Engine):** Der Concurrency-Kern. Ein Hintergrunddienst, der Jobs verarbeitet, Produktseiten scrapt, Preise extrahiert und Schwellenwerte prüft.
2. **`priceradar-api` (REST-Backend):** Verwaltet Produkte, Zielpreise und Alarme. Nimmt Daten aus der CLI entgegen und liefert Auswertungen.
3. **`priceradar-cli` (Terminal-Tool):** Das Admin-Interface für deinen Server. Zum schnellen Hinzufügen von URLs, Auslösen manueller Scrapes und Anzeigen von Live-Statistiken.

---

## 📅 Der 4-Phasen-Entwicklungsplan

### Phase 1: Der Worker & Scraper (Concurrency & Standard-Lib)

> **Ziel:** Go-Concurrency (Goroutines, Channels, Context) und HTML-Parsing verstehen.

* **Schritt 1:** Schreibe ein Paket `scraper`, das ein HTML-Dokument per HTTP holt (`net/http`) und den Preis extrahiert.
* *Tipp:* Nutze `golang.org/x/net/html` oder die beliebte Library `goquery` (wie jQuery für Go).


* **Schritt 2:** Baue den **Worker-Pool**. Er erstelle eine feste Anzahl an Goroutines (z. B. 5 Worker), die über einen Go-**Channel** URLs zum Scrapen entgegennehmen.
* **Schritt 3:** Implementiere **Rate Limiting & Backoff**. Wenn ein HTTP 429 zurückkommt, nutze `time.Sleep` mit *Exponential Backoff* oder `time.Ticker`, um Anfragen pro Domain zu drosseln.
* **Schritt 4:** Nutze **`context.Context`** mit Timeout, damit hängende HTTP-Requests nach 5 Sekunden abgebrochen werden.

### Phase 2: Die REST-API & Datenbank (Web & SQL)

> **Ziel:** Sauberes Go-Projektlayout, Routing und DB-Zugriff lernen.

* **Schritt 1:** Setze PostgreSQL auf (z. B. via Docker) und erstelle Tabellen für `products`, `price_history` und `alerts`.
* **Schritt 2:** Nutze **`sqlc`**, um typsicheren Go-Code aus deinen SQL-Queries zu generieren.
* **Schritt 3:** Baue das HTTP-Backend mit **`Chi`** oder **`Gin`**:
* `POST /products` (Neue URL + Zielpreis hinzufügen)
* `GET /products` (Liste aller Produkte inkl. letztem Preis)
* `GET /products/{id}/history` (Preisverlauf für Diagramme)
* `POST /scrape-jobs` (Triggert sofortigen Crawl-Auftrag)


* **Schritt 4:** Verbinde Worker und API über die DB: Die API schreibt Jobs in die Datenbank (oder einen Channel), der Worker holt und verarbeitet sie.

### Phase 3: Die CLI (Kommandozeilen-Engineering)

> **Ziel:** Schnell ausführbare Binaries, Argument-Parsing und Terminal-UIs bauen.

* **Schritt 1:** Setze die CLI mit dem Standard-Framework **`cobra`** auf (wird auch von Kubernetes/Docker genutzt).
* **Schritt 2:** Implementiere Subcommands, die mit deiner REST-API kommunizieren:
* `priceradar add [https://shop.com/gpu](https://shop.com/gpu) --target 400.00`
* `priceradar list` (Gibt eine formatierte Tabelle im Terminal aus)
* `priceradar trigger` (Startet den Crawler manuell über die API)


* **Schritt 3:** Nutze `lipgloss` oder `tablewriter` für hübsches Terminal-Formatting (Farben, Tabellen, Progress-Bars).

### Phase 4: Job-Matching & Alerting (Business-Logik)

> **Ziel:** Fehlerbehandlung und Event-Logik.

* **Schritt 1:** Sobald der Worker einen neuen Preis speichert, vergleicht er ihn mit dem `target_price`.
* **Schritt 2:** Fällt der Preis darunter, löst der Worker ein Event aus (z. B. ein Webhook-Aufruf an einen Discord-/Slack-Channel oder eine Konsolenausgabe).
* **Schritt 3:** Baue Graceful Shutdowns ein: Wenn du `Ctrl+C` drückst, soll der Worker laufende HTTP-Requests geordnet beenden, bevor er stoppt.

---

## 📂 Vorgeschlagene Ordnerstruktur (Standard Go Layout)

Um deinen Code direkt nach Best Practices zu strukturieren:

```text
priceradar/
├── cmd/
│   ├── api/        # main.go für den REST-Server
│   ├── worker/     # main.go für den Scraper-Daemon
│   └── cli/        # main.go für das CLI-Tool
├── internal/       # Privater Code (kann nicht von extern importiert werden)
│   ├── db/         # Generierter sqlc-Code & DB-Verbindung
│   ├── scraper/    # Scraper-Logik, Worker-Pool & Channels
│   ├── api/        # HTTP Handler & Middleware
│   └── config/     # Umweltvariablen / App-Konfiguration
├── db/
│   ├── migrations/ # SQL-Migrationsdateien (golang-migrate)
│   └── queries/    # .sql Dateien für sqlc
├── go.mod
└── go.sum

```

---

## 🛠️ Dein Tech-Stack für dieses Projekt

* **Routing:** `[github.com/go-chi/chi/v5](https://github.com/go-chi/chi/v5)`
* **CLI:** `[github.com/spf13/cobra](https://github.com/spf13/cobra)`
* **DB-Tooling:** `sqlc` (Generator) + `pgx` (Postgres Driver)
* **HTML Parsing:** `[github.com/PuerkitoBio/goquery](https://github.com/PuerkitoBio/goquery)`
* **Terminal UI:** `[github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)` (Optional)