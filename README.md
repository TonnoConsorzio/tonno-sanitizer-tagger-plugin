# Vikunja Tonno Sanitizer Tagger Plugin

Un plugin nativo in Go per l'ecosistema [Vikunja](https://vikunja.io/) progettato per l'esecuzione tramite interprete **Yaegi**.
Il plugin intercetta automaticamente la creazione e l'aggiornamento dei task (tramite l'Event Broker interno di Vikunja) per auto-compilarli in base al titolo, applicando in automatico etichette (tag), priorità e un assegnatario predefinito.

## Funzionalità

* **Auto-Tagging Intelligente**: Ricerca parole chiave nel titolo del task (es. "fix", "spesa", "meeting") e assegna dinamicamente al task le etichette di Vikunja corrispondenti.
* **Auto-Assegnazione**: Assegna automaticamente il task all'utente che l'ha creato, per evitare task orfani. (Disattivabile)
* **Descrizione e Priorità di Fallback**: Inserisce un testo di default (es. "Bo...") nella descrizione se lasciata vuota, e imposta una priorità predefinita se non dichiarata.
* **Idempotenza Assoluta**: Non crea doppioni. Controlla nativamente il database XORM per assicurarsi di non duplicare le etichette.
* **Plug & Play Senza ID Hardcoded**: Usa direttamente il **nome dell'etichetta** per cercarla sul database. Nessun bisogno di recuperare astrusi ID numerici!

---

## Installazione

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

---

## Configurazione (`config.json`)

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
    "Spesa": ["comprare", "spesa", "prendere"]
  }
}
```

* `"tag_mappings"`: La chiave (es. `"Fix o Bug"`) **deve corrispondere esattamente al titolo dell'etichetta già creata all'interno del tuo Vikunja**. I valori nell'array sono le parole chiave che, se trovate nel titolo del task, scateneranno l'assegnazione automatica di quell'etichetta.

---

## Licenza

Questo progetto è rilasciato sotto licenza MIT. Vedi il file `LICENSE` per i dettagli.
