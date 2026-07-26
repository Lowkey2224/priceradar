# Deployment Plan
**Ja, das ist zu 100 % kostenlos möglich!** Für Lern-, Test- und Portfolio-Zwecke gibt es im modernen Cloud-Ökosystem hervorragende Anbieter.

Der Trick bei einer verteilten Architektur besteht darin, die Dienste dort zu hosten, wo die Free-Tiers am großzügigsten sind.

---

## ☁️ Wo wird was gehostet?

| Komponente | Hosting-Plattform | Grund für die Wahl |
| --- | --- | --- |
| **PostgreSQL DB** | **Neon.tech** oder **Supabase** | Managed PostgreSQL mit ca. 500 MB–3 GB Speicher. Im Gegensatz zu Render-Free-DBs werden deine Daten hier **nicht** nach 30 Tagen gelöscht. |
| **REST API** | **Render.com** *(oder Koyeb)* | Kostenlose Web-Services direkt aus deinem GitHub-Repository. Baut den Go-Code automatisch. |
| **Worker (Scraper)** | **Render.com** *(oder in API integriert)* | Kann als eigener Hintergrund-Dienst oder (für den Free Tier ideal) als Goroutine direkt im API-Prozess gestartet werden. |
| **CLI-Tool** | **Kein Cloud-Hosting** | Wird lokal auf deinem Rechner gebaut (`go build`) und greift über das Internet auf deine Live-API zu. |

---

## 🚀 Der Schritt-für-Schritt Deployment-Plan

### Schritt 1: Datenbank in der Cloud anlegen (Neon.tech)

1. Erstelle einen kostenlosen Account auf **[Neon.tech](https://neon.tech)**.
2. Erstelle ein neues Projekt (z. B. `priceradar-db`) und wähle als Region **Frankfurt (Europe)**.
3. Neon gibt dir sofort eine `DATABASE_URL` (Connection String) im Format:
`postgres://user:password@ep-xyz.eu-central-1.aws.neon.tech/neondb?sslmode=require`
4. Führe deine SQL-Migrationen gegen diesen DB-String aus (z. B. mit `golang-migrate`).

---

### Schritt 2: Anwendung auf Cloud-Readiness vorbereiten

Damit Go-Anwendungen in der Cloud laufen, müssen zwei Voraussetzungen erfüllt sein:

1. **Konfiguration über Umgebungsvariablen (`ENV`):**
Dein Go-Code sollte die Datenbank-URL und den Server-Port dynamisch einlesen:
```go
dbURL := os.Getenv("DATABASE_URL")
port := os.Getenv("PORT") // Render vergibt den Port dynamisch
if port == "" {
    port = "8080"
}

```


2. **Worker und API bündeln (Free-Tier Optimierung):**
Auf kostenlosen Hostern steht dir oft nur *ein* dauerhafter Web-Service zur Verfügung.
* **Praxis-Tipp:** Starte in deiner `cmd/api/main.go` den Scraper-Worker einfach in einer separaten Goroutine im Hintergrund, bevor das HTTP-Routing startet.



---

### Schritt 3: API & Worker auf Render.com deployen

1. Pushe deinen Code in ein öffentliches oder privates **GitHub-Repository**.
2. Registriere dich auf **[Render.com](https://render.com)**.
3. Klicke auf **New +** $\rightarrow$ **Web Service** und verbinde dein GitHub-Repo.
4. Stelle folgende Parameter ein:
* **Runtime:** `Go`
* **Build Command:** `go build -o main ./cmd/api`
* **Start Command:** `./main`
* **Environment Variables:**
* `DATABASE_URL` = *(Dein Neon.tech-Connection String)*




5. Klicke auf **Create Web Service**. Render baut das Go-Projekt und stellt dir in wenigen Minuten eine HTTPS-URL bereit (z. B. `[https://priceradar-api.onrender.com](https://priceradar-api.onrender.com)`).

---

### Schritt 4: CLI lokal kompilieren & mit Cloud-API verbinden

Das CLI-Tool musst du nicht in der Cloud hosten. Du erstellst daraus eine native ausführbare Datei auf deinem Rechner:

1. Füge deiner CLI eine Konfiguration für die Basis-URL hinzu (z. B. als Flag oder `.env`-Datei).
2. Baue die Binary:
```bash
go build -o priceradar ./cmd/cli

```


3. Teste deine Cloud-API vom lokalen Terminal aus:
```bash
./priceradar add https://amazon.de/dp/B0XYZ --target 350.00 --api https://priceradar-api.onrender.com

```



---

## 💡 Wichtige Dinge, die du beim Free-Tier beachten musst

* **Cold Starts (Kaltstarts):** Auf kostenlosen Web-Services wie Render schläft der Server ein, wenn 15 Minuten lang keine HTTP-Anfrage eingeht. Wenn du danach per CLI eine Anfrage schickst, dauert die erste Antwort 20–30 Sekunden, weil der Server erst aufwachen muss. Das ist für ein Lernprojekt völlig normal.
* **IP-Blocking beim Scraping:** Da Cloud-Anbieter (AWS, Render, GCP) oft bekannte IP-Bereiche nutzen, blockieren manche Online-Shops Anfragen von dort schneller als von deinem Heimanschluss. Schreibe deinen Scraper so, dass er einen echten `User-Agent`-Header mitsendet.