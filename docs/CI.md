# GitHub Actions CI

Dieses Projekt nutzt GitHub Actions für die kontinuierliche Integration. Die Workflows stellen sicher, dass der Code stabil, sicher und konsistent bleibt.

## Checks in der CI

Der Workflow `ci.yml` führt folgende Schritte aus:

1.  **Binary Guard**: Stellt sicher, dass keine Binärdateien oder unerwartete ausführbare Dateien ins Repository committet wurden.
2.  **Go Matrix**: Testet gegen die in `go.mod` definierte Mindestversion von Go (1.22) sowie die aktuelle `stable` Version.
3.  **Formatting**: Überprüft mit `gofmt`, ob der Code korrekt formatiert ist.
4.  **Dependency Check**: Stellt sicher, dass `go mod tidy` ausgeführt wurde und `go.mod`/`go.sum` aktuell sind.
5.  **Tests**: Führt alle Tests im Root-Modul und im `tools/` Modul aus.
6.  **Race Detector**: Führt Tests mit dem Go Race Detector aus (`-race`).
7.  **Vulnerability Check**: Nutzt `govulncheck`, um auf bekannte Sicherheitslücken in Abhängigkeiten zu prüfen.

## Lokale Ausführung

Um die gleichen Checks lokal auszuführen, kann das `Makefile` verwendet werden:

- **Alle Checks**: `make verify`
- **Tests**: `make test`
- **Race Detector**: `make test-race`
- **Format-Check**: `make fmt-check`
- **Vulnerability Check**: `make vulncheck`

## Caching

Die CI nutzt die Caching-Funktionen von `actions/setup-go`, um Builds und Dependency-Downloads zu beschleunigen. Der Cache basiert auf den `go.sum` Dateien der Module.
