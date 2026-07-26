#Entwicklungsplan Detail
Hier ist dein detaillierter, feingliedriger Entwicklungsplan. Er führt dich **Schritt für Schritt** durch die Implementierung, sodass du nie vor einer unlösbaren "Wall of Code" stehst.

---

## 🛠️ Phase 1: Der Scraper Core & Concurrency (Der Go-Einstieg)

In dieser Phase arbeitest du vorerst nur in der Konsole / kleinen Tests, um Go’s Concurrency-Konzepte sauber zu verinnerlichen.

### Schritt 1.1: HTML-Fetching & Basis-Parsing

1. Lege das Verzeichnis `internal/scraper` an.
2. Erstelle eine Datei `scraper.go` mit einer Funktion `FetchPrice(ctx context.Context, url string) (float64, error)`.
3. Nutze `net/http`, um eine HTTP-GET-Anfrage abzusetzen.
> **Wichtig:** Setze explizit einen Browser `User-Agent` im Header, damit Shops deine Anfrage nicht sofort blockieren.


4. Verwende `goquery`, um das HTML zu parsen und ein Preiselement (z. B. via CSS-Selektor `.price` oder `[itemprop="price"]`) auszulesen.
5. Wandle den gefundenen String (z. B. `"399,99 €"`) in ein valides `float64` um.

### Schritt 1.2: Ein einfacher Worker-Pool (Channels & Goroutines)

1. Erstelle die Datei `worker.go`.
2. Definiere ein Struct `Job`, das eine `ID`, eine `URL` und einen `TargetPrice` enthält.
3. Definiere ein Struct `Result`, das den `Job`, den `FetchedPrice` und ein `Error`-Feld enthält.
4. Erstelle zwei Channels:
* `jobs := make(chan Job, 10)`
* `results := make(chan Result, 10)`


5. Schreibe die Funktion `worker(id int, jobs <-chan Job, results chan<- Result)`:
* Nutze eine `for job := range jobs`-Schleife, um kontinuierlich Jobs abzuarbeiten.
* Rufe `FetchPrice` auf und sende das `Result` in den `results`-Channel.



### Schritt 1.3: Abbrüche und Timeouts mit `context.Context`

1. Erweitere deine HTTP-Requests im Scraper so, dass sie `http.NewRequestWithContext(ctx, "GET", url, nil)` nutzen.
2. Übergib dem Worker einen Context mit Timeout: `ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)`.
3. Teste, was passiert, wenn eine Webseite nicht antwortet: Der Request muss sauber nach 5 Sekunden abbrechen!

### Schritt 1.4: Dynamic Rate Limiting & Backoff

1. Implementiere eine Logik für den HTTP-Statuscode `429 Too Many Requests`.
2. Wenn Status `429` zurückkommt, nutze `time.Sleep` mit *Exponential Backoff* (z. B. 2s $\rightarrow$ 4s $\rightarrow$ 8s Pause) für die betroffene Domain.

---

## 🗄️ Phase 2: Datenbank, Domain Logic & REST-API

Jetzt bringst du Ordnung, Datenhaltung und HTTP-Endpunkte ins Spiel.

### Schritt 2.1: Datenbank-Setup & Migrationen

1. Starte lokal einen PostgreSQL-Docker-Container (oder erstelle deine DB auf Neon.tech).
2. Erstelle unter `db/migrations/` zwei SQL-Dateien mit `golang-migrate`:
* `000001_create_products_table.up.sql`: Erstellt Tabelle `products` (`id`, `url`, `title`, `target_price`, `current_price`, `created_at`).
* `000002_create_price_history_table.up.sql`: Erstellt Tabelle `price_history` (`id`, `product_id`, `price`, `fetched_at`).


3. Führe die Migrationen aus.

### Schritt 2.2: Generierung von Go-Code mit `sqlc`

1. Erstelle eine `sqlc.yaml`-Konfigurationsdatei im Root.
2. Schreibe unter `db/queries/products.sql` rohe SQL-Statements für CRUD-Operationen:
* `-- name: CreateProduct :one` $\rightarrow$ `INSERT INTO products...`
* `-- name: ListProducts :many` $\rightarrow$ `SELECT * FROM products...`
* `-- name: UpdateProductPrice :exec` $\rightarrow$ `UPDATE products SET current_price = $1...`


3. Führe `sqlc generate` im Terminal aus. Du hast jetzt extrem performanten, typsicheren Go-Code unter `internal/db/` ohne ORM-Overhead!

### Schritt 2.3: REST-API mit Chi und JSON-Handling

1. Installiere den Router `go-chi/chi/v5`.
2. Erstelle `cmd/api/main.go` und `internal/api/handlers.go`.
3. Schreibe Handler-Funktionen:
* `CreateProductHandler`: Liest JSON aus dem Request-Body, validiert die URL und speichert das Produkt via `sqlc` in der DB.
* `ListProductsHandler`: Holt alle Produkte aus der DB und gibt sie als JSON zurück.
* `GetPriceHistoryHandler`: Gibt historische Preisdaten für ein Produkt aus.


4. Implementiere eine einfache Middleware für Logging und Panic-Recovery.

### Schritt 2.4: Integration von Worker und API

1. Lasse beim Starten des API-Servers (`cmd/api/main.go`) im Hintergrund über eine Goroutine den Worker-Pool laufen (`go startWorkerPool(...)`).
2. Implementiere einen Endpunkt `POST /scrape-jobs/trigger`, der alle URLs aus der Datenbank zieht und als Jobs in den Worker-Channel pusht.

---

## 💻 Phase 3: Die CLI (`cmd/cli`)

Die CLI dient dir als komfortables Steuerungswerkzeug auf deinem lokalen Rechner.

### Schritt 3.1: Cobra-Setup

1. Installiere das CLI-Framework `[github.com/spf13/cobra](https://github.com/spf13/cobra)`.
2. Initialisiere die CLI in `cmd/cli/main.go` mit einem Basis-Befehl `priceradar`.

### Schritt 3.2: Subcommands bauen

1. **Befehl `add`:**
* Syntax: `priceradar add <URL> --target <preis>`
* Logik: Sendet einen HTTP-`POST`-Request an deine REST-API.


2. **Befehl `list`:**
* Syntax: `priceradar list`
* Logik: Liest Produkte von der API (`GET /products`) und gibt sie als schön formatierte Tabelle auf der Konsole aus (z. B. mit `olekukonko/tablewriter` oder `charmbracelet/lipgloss`).


3. **Befehl `trigger`:**
* Syntax: `priceradar trigger`
* Logik: Schickt einen Request an den Trigger-Endpunkt der API, um den Scrape-Vorgang sofort zu starten.



---

## 🔔 Phase 4: Business Logic, Alerts & Cloud-Deployment

Der finale Feinschliff für Produktionstauglichkeit.

### Schritt 4.1: Schwellenwert-Check & Alerts

1. Sobald der Worker einen neuen Preis speichert, prüft er: Ist `fetched_price <= target_price`?
2. Wenn ja, rufe ein Notification-Package auf:
* Schreibe ein Interface `Notifier` mit Methode `Notify(product, price)`.
* Implementiere ein `DiscordNotifier` oder `SlackNotifier`, das im Erfolgsfall eine Webhook-Message an dein Handy/Discord schickt.



### Schritt 4.2: Graceful Shutdown

1. Fange OS-Signale (`SIGINT`, `SIGTERM`) in deiner `main.go` ab.
2. Fahre den HTTP-Server mit `server.Shutdown(ctx)` geordnet herunter.
3. Schließe die Channels des Worker-Pools und warte mit einem `sync.WaitGroup`, bis alle laufenden Scrapes beendet wurden, bevor die App stoppt.

### Schritt 4.3: Deployment auf Neon + Render

1. Pushe deinen Code nach GitHub.
2. Setze die DB auf **Neon.tech** auf und führe die Migrationen aus.
3. Setze den Web Service auf **Render.com** auf:
* Verknüpfe dein GitHub-Repo.
* Setze Umgebungsvariablen (`DATABASE_URL`, `PORT`).
* Deploye die Anwendung!


4. Teste deine lokale CLI (`priceradar list`), indem du sie gegen deine Render-Cloud-URL verbindest.