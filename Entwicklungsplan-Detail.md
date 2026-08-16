# 📝 Überarbeiteter Entwicklungsplan (Ohne API- & CLI-Frameworks)

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
2. Wenn Status `429` zurückkommt, nutze `time.Sleep` mit *Exponential Backoff* (z. B. 2s -> 4s -> 8s Pause) für die betroffene Domain.

---

## 🗄️ Phase 2: Datenbank, Domain Logic & Pure Standard Library REST-API

Jetzt bringst du Ordnung, Datenhaltung und HTTP-Endpunkte ins Spiel – **rein mit `net/http` aus Go 1.22+**.

### Schritt 2.1: Datenbank-Setup & Migrationen

1. Starte lokal einen PostgreSQL-Docker-Container (oder erstelle deine DB auf Neon.tech).
2. Erstelle unter `db/migrations/` zwei SQL-Dateien mit `golang-migrate`:
* `000001_create_products_table.up.sql`: Erstellt Tabelle `products` (`id`, `url`, `title`, `target_price`, `current_price`, `created_at`).
* `000002_create_price_history_table.up.sql`: Erstellt Tabelle `price_history` (`id`, `product_id`, `price`, `fetched_at`).
3. Führe die Migrationen aus.

### Schritt 2.2: Generierung von Go-Code mit `sqlc`

1. Erstelle eine `sqlc.yaml`-Konfigurationsdatei im Root.
2. Schreibe unter `db/queries/products.sql` rohe SQL-Statements für CRUD-Operationen:
* `-- name: CreateProduct :one` -> `INSERT INTO products...`
* `-- name: ListProducts :many` -> `SELECT * FROM products...`
* `-- name: UpdateProductPrice :exec` -> `UPDATE products SET current_price = $1...`
3. Führe `sqlc generate` im Terminal aus. Du hast jetzt extrem performanten, typsicheren Go-Code unter `internal/db/` ohne ORM-Overhead!

### Schritt 2.3: REST-API mit Go's Standard `http.ServeMux` (Kein Chi!)

1. **Routing ohne externe Libs:** Ab Go 1.22 unterstützt `http.ServeMux` HTTP-Methoden und Pfad-Parameter direkt (z. B. `mux.HandleFunc("POST /products", CreateProductHandler)`).
2. Erstelle `cmd/api/main.go` und `internal/api/handlers.go`.
3. Schreibe Handler-Funktionen mit der Standard-Signatur `func(w http.ResponseWriter, r *http.Request)`:
* `CreateProductHandler`: Liest JSON aus dem Request-Body (`json.NewDecoder(r.Body)`), validiert die URL und speichert das Produkt via `sqlc` in der DB.
* `ListProductsHandler`: Holt alle Produkte aus der DB und gibt sie als JSON mit `json.NewEncoder(w).Encode(...)` zurück.
* `GetPriceHistoryHandler`: Gibt historische Preisdaten für ein Produkt aus.
4. **Eigene Middleware:** Implementiere eine einfache Middleware als Standard-Go-Funktion:

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

### Schritt 2.4: Integration von Worker und API

1. Lasse beim Starten des API-Servers (`cmd/api/main.go`) im Hintergrund über eine Goroutine den Worker-Pool laufen (`go startWorkerPool(...)`).
2. Implementiere einen Endpunkt `POST /scrape-jobs/trigger`, der alle URLs aus der Datenbank zieht und als Jobs in den Worker-Channel pusht.

---

## 💻 Phase 3: Die CLI (`cmd/cli`) mit Go's Standard `flag`-Package

Statt Cobra nutzt du das eingebaute `flag`-Package, um Subcommands und Konsolenausgaben ohne externe Bibliotheken umzusetzen.

### Schritt 3.1: Standard Subcommand-Parsing Setup

1. Initialisiere die CLI in `cmd/cli/main.go`.
2. Verwende `os.Args` und `flag.NewFlagSet`, um Subcommands zu verarbeiten:

switch os.Args[1] {
case "add":
    addCmd.Parse(os.Args[2:])
case "list":
    listCmd.Parse(os.Args[2:])
case "trigger":
    triggerCmd.Parse(os.Args[2:])
}

### Schritt 3.2: Subcommands bauen

1. **Befehl `add`:**
* Syntax: `priceradar add -url <URL> -target <preis>` (definiert via `flag.NewFlagSet("add", flag.ExitOnError)`)
* Logik: Sendet einen HTTP-`POST`-Request via Go `net/http` an deine REST-API.

2. **Befehl `list`:**
* Syntax: `priceradar list`
* Logik: Liest Produkte von der API (`GET /products`) und verwendet Go's eingebautes Package **`text/tabwriter`**, um eine perfekt formatierte Tabelle direkt in das Terminal zu drucken (keine externe Table-Bibliothek nötig):

w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
fmt.Fprintln(w, "ID\tTITLE\tPRICE\tTARGET")
// ... Schleife über Produkte ...
w.Flush()

3. **Befehl `trigger`:**
* Syntax: `priceradar trigger`
* Logik: Schickt einen HTTP-`POST`-Request an den Trigger-Endpunkt der API, um den Scrape-Vorgang sofort zu starten.

---

## 🔔 Phase 4: Business Logic, Alerts & Cloud-Deployment

Der finale Feinschliff für Produktionstauglichkeit.

### Schritt 4.1: Schwellenwert-Check & Alerts

1. Sobald der Worker einen neuen Preis speichert, prüft er: Ist `fetched_price <= target_price`?
2. Wenn ja, rufe ein Notification-Package auf:
* Schreibe ein Interface `Notifier` mit Methode `Notify(product, price)`.
* Implementiere ein `DiscordNotifier` oder `SlackNotifier` via Standard `net/http`, das im Erfolgsfall eine Webhook-JSON-Payload an deinen Discord/Slack-Kanal schickt.

### Schritt 4.2: Graceful Shutdown

1. Fange OS-Signale (`SIGINT`, `SIGTERM`) in deiner `main.go` mit `signal.Notify` ab.
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