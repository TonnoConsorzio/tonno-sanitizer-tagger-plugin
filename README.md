# Vikunja Tonno Sanitizer Tagger Plugin

*Read this in other languages: [English](#english), [Italiano](#italiano).*

---

<a name="english"></a>
## 🇬🇧 English

A native Go plugin for the [Vikunja](https://vikunja.io/) ecosystem, designed to be executed via the **Yaegi** interpreter.
The plugin automatically intercepts task creation and updates (via Vikunja's internal Event Broker) to auto-fill them based on their title, automatically applying tags (labels), priorities, and a default assignee.

### Features

* **Smart Auto-Tagging**: Searches for keywords in the task title (e.g., "fix", "expense", "meeting") and dynamically assigns the corresponding Vikunja labels to the task.
* **Auto-Assignment**: Automatically assigns the task to the user who created it, preventing orphaned tasks. (Can be disabled)
* **Fallback Description and Priority**: Inserts a default text (e.g., "Bo...") into the description if left empty, and sets a default priority if none is provided.
* **Absolute Idempotency**: Prevents duplicates. Natively checks the XORM database to ensure labels are not duplicated on the same task.
* **Plug & Play Without Hardcoded IDs**: Uses the **label name** directly to search for it in the database. No need to retrieve obscure numeric IDs!

### Installation

1. Make sure the plugin system is enabled in your Vikunja `config.yml`:
   ```yaml
   plugins:
     enabled: true
     dir: "plugins"
     loader: "yaegi"
   ```
2. Clone this repository and move the `tonno_sanitizer_tagger.go` file and the `TonnoSanitizerTagger/` directory into your instance's plugins folder:
   ```text
   plugins/
   ├── tonno_sanitizer_tagger.go
   └── TonnoSanitizerTagger/
       └── config.json
   ```
3. (Optional) Rename the `config.example.json` file to `config.json` and customize your keywords and options.
4. Restart the Vikunja backend.

### Configuration (`config.json`)

The plugin features dynamic capabilities. The `config.json` file is read **at every event**! This means you can change rules *on the fly* without ever having to restart Vikunja.

```json
{
  "features": {
    "enable_auto_assign": true,
    "default_description": "Bo...",
    "default_priority": 3
  },
  "tag_mappings": {
    "Fix o Bug": ["bug", "error", "fix", "crash", "problem"],
    "Spesa": ["buy", "expense", "get", "order"]
  }
}
```

* `"tag_mappings"`: The key (e.g., `"Fix o Bug"`) **must exactly match the title of the label already created within your Vikunja instance**. The values in the array are the keywords that, if found in the task title, will trigger the automatic assignment of that label.

### License

This project is licensed under the MIT License. See the `LICENSE` file for details.

---

<a name="italiano"></a>
## 🇮🇹 Italiano

Un plugin nativo in Go per l'ecosistema [Vikunja](https://vikunja.io/) progettato per l'esecuzione tramite interprete **Yaegi**.
Il plugin intercetta automaticamente la creazione e l'aggiornamento dei task (tramite l'Event Broker interno di Vikunja) per auto-compilarli in base al titolo, applicando in automatico etichette (tag), priorità e un assegnatario predefinito.

### Funzionalità

* **Auto-Tagging Intelligente**: Ricerca parole chiave nel titolo del task (es. "fix", "spesa", "meeting") e assegna dinamicamente al task le etichette di Vikunja corrispondenti.
* **Auto-Assegnazione**: Assegna automaticamente il task all'utente che l'ha creato, per evitare task orfani. (Disattivabile)
* **Descrizione e Priorità di Fallback**: Inserisce un testo di default (es. "Bo...") nella descrizione se lasciata vuota, e imposta una priorità predefinita se non dichiarata.
* **Idempotenza Assoluta**: Non crea doppioni. Controlla nativamente il database XORM per assicurarsi di non duplicare le etichette.
* **Plug & Play Senza ID Hardcoded**: Usa direttamente il **nome dell'etichetta** per cercarla sul database. Nessun bisogno di recuperare astrusi ID numerici!

### Installazione

1. Assicurati che nel tuo `config.yml` di Vikunja sia abilitato il sistema di plugin:
   ```yaml
   plugins:
     enabled: true
     dir: "plugins"
     loader: "yaegi"
   ```
2. Clona questo repository e sposta il file `tonno_sanitizer_tagger.go` e la directory `TonnoSanitizerTagger/` all'interno della cartella dei plugin della tua istanza:
   ```text
   plugins/
   ├── tonno_sanitizer_tagger.go
   └── TonnoSanitizerTagger/
       └── config.json
   ```
3. (Opzionale) Rinomina il file `config.example.json` in `config.json` e personalizza le tue parole chiave e opzioni.
4. Riavvia il backend di Vikunja.

### Configurazione (`config.json`)

Il plugin è dotato di super poteri dinamici. Viene letto il file `config.json` **a ogni evento**! Questo significa che puoi cambiare le regole *al volo* senza dover mai riavviare Vikunja.

```json
{
  "features": {
    "enable_auto_assign": true,
    "default_description": "Bo...",
    "default_priority": 3
  },
  "tag_mappings": {
    "Fix o Bug": ["bug", "errore", "fix", "crash", "problema"],
    "Spesa": ["comprare", "spesa", "prendere", "acquistare", "ordinare"]
  }
}
```

* `"tag_mappings"`: La chiave (es. `"Fix o Bug"`) **deve corrispondere esattamente al titolo dell'etichetta già creata all'interno del tuo Vikunja**. I valori nell'array sono le parole chiave che, se trovate nel titolo del task, scateneranno l'assegnazione automatica di quell'etichetta.

### Licenza

Questo progetto è rilasciato sotto licenza MIT. Vedi il file `LICENSE` per i dettagli.
