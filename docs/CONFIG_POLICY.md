# goConfy Config-Policy & Best Practices

Diese Policy legt fest, wie Konfigurationen in goConfy-basierten Projekten definiert, dokumentiert und gepflegt werden müssen. Ziel ist eine 100%ige Abdeckung aller im Code verwendeten Konfigurationen in einer kommentierten Beispiel-Datei (`config.example.yml`) sowie die Validierung über ein Schema.

## Kernregeln

### 1. Keine versteckten Konfigurationen (No Hidden Config)
*   **Regel:** Jede Konfiguration, die im App-Code verwendet wird, **MUSS** in der `config.example.yml` dokumentiert sein.
*   **Verbot:** Kein direkter Zugriff auf `os.Getenv()` oder die Verwendung von Hardcoded-Defaults im Code, ohne dass der entsprechende Key in der Beispiel-Config existiert.
*   **Warum:** Transparenz für Ops und Entwickler; alle Stellschrauben müssen an einem Ort sichtbar sein.

### 2. Single Source of Truth
*   **Regel:** Dokumentation (Kommentare), Defaults und Struktur werden primär in der `config.example.yml` (bzw. den Go-Struct-Tags) definiert.
*   **Generatoren:** Tools wie `goconfygen` nutzen diese Informationen, um neue Konfigurationsdateien oder Schemata zu erzeugen.

### 3. Pflicht zur Beschreibung (Descriptions Required)
*   **Regel:** Jeder Key in der YAML-Datei muss mindestens eine Kommentarzeile direkt darüber haben.
*   **Inhalt:**
    *   Zweck des Feldes (Was macht es?).
    *   Wann sollte es geändert werden?
    *   Sicherheitshinweise (z.B. bei Secrets).
    *   Erwartetes Format (z.B. bei Zeitangaben wie `30s`).

### 4. Enums dokumentieren (Enums Documented)
*   **Regel:** Felder mit einer begrenzten Auswahl an Optionen (Enum-like) müssen im Kommentar alle erlaubten Werte auflisten.
*   **Beispiel:** `mode: "fast" # Erlaubte Werte: "fast", "safe", "strict". Default: "safe".`

### 5. Schema-Validierung (Schema Enforced)
*   **Regel:** Die Konfigurationsstruktur muss über ein Schema (JSON Schema) validierbar sein.
*   **Anforderung:** Unbekannte Keys führen zum Fehler (Strict Mode), Typen müssen korrekt sein und Constraints (wie `required`) müssen geprüft werden.

### 6. Environment Macros dokumentieren
*   **Regel:** Wenn Macros wie `{ENV:VAR_NAME}` oder `{FILE:path}` verwendet werden, muss im Kommentar das erwartete Format der Umgebungsvariable stehen und ob ein Default-Wert zulässig ist (z.B. `{ENV:PORT:8080}`).

---

## Beispiel-Struktur (Go-Struct Tags)

In goConfy nutzen wir Struct-Tags, um die Policy direkt im Code zu verankern:

```go
type DBConfig struct {
    // Port ist die Portnummer für die DB-Verbindung.
    // Default: 5432.
    Port int `yaml:"port" default:"5432" desc:"Database port" env:"DB_PORT"`

    // Password ist das Passwort für den DB-Zugriff.
    // Muss als Environment-Variable gesetzt werden.
    Password string `yaml:"password" secret:"true" env:"DB_PASSWORD" desc:"Database secret password"`
}
```

## Do's & Don'ts

| Do | Don't |
| :--- | :--- |
| **Do:** Nutze `desc` Tags für alle Felder. | **Don't:** Nutze `os.Getenv("MY_VAR")` direkt im Business-Logik-Code. |
| **Do:** Halte `config.example.yml` synchron zum Code (CI-Checks!). | **Don't:** Füge neue Config-Keys hinzu, ohne die Dokumentation zu aktualisieren. |
| **Do:** Nutze `{ENV:...}` Macros für sensitive Daten in Beispielen. | **Don't:** Schreibe echte Passwörter in `config.example.yml`. |
| **Do:** Nutze aussagekräftige Namen (z.B. `timeout_ms` oder `timeout: 10s`). | **Don't:** Nutze kryptische Abkürzungen ohne Erklärung. |

---

## Automatisierung & CI

Um Drift zu vermeiden, wird in der CI geprüft:
1.  Lässt sich `config.example.yml` fehlerfrei gegen die App-Struct laden?
2.  Sind alle Felder der App-Struct in der `config.example.yml` vorhanden?
3.  Sind alle Keys in der `config.example.yml` auch in der Struct definiert?
4.  Ist das JSON Schema aktuell?
