# Release Guide

Dieses Dokument beschreibt den Prozess zum Erstellen einer neuen Version von `goConfy`.

## Versionierungsschema

Wir verwenden [SemVer](https://semver.org/).

- **v0.x.x**: Experimentelle Phase. Breaking Changes an der API sind möglich.
- **v1.x.x**: Stabile API. Breaking Changes erfordern eine neue Major-Version.

## Schritte für ein Release

1. **Vorbereitung**:
    - Stelle sicher, dass alle Tests lokal bestehen: `make test test-race`.
    - Führe den Linter aus: `make lint`.
    - Überprüfe auf Sicherheitslücken: `make vulncheck`.
    - Aktualisiere ggf. `CHANGELOG.md`.

2. **Tagging**:
    - Erstelle einen neuen Git-Tag (Beispiel für `v0.1.0`):
      ```bash
      git tag -a v0.1.0 -m "Release v0.1.0"
      ```
    - Pushe den Tag zu GitHub:
      ```bash
      git push origin v0.1.0
      ```

3. **GitHub Release**:
    - Navigiere zu "Releases" im GitHub-Repository.
    - Wähle "Draft a new release".
    - Wähle den soeben gepushten Tag aus.
    - Füge die Release-Notes aus dem `CHANGELOG.md` ein.
    - Veröffentliche das Release.
