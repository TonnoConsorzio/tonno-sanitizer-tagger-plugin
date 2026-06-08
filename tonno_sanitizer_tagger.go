package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/vikunja/vikunja/pkg/db"
	"github.com/vikunja/vikunja/pkg/events"
	"github.com/vikunja/vikunja/pkg/log"
	"github.com/vikunja/vikunja/pkg/models"
	"github.com/vikunja/vikunja/pkg/plugins"
)

// Struttura della configurazione JSON
type PluginConfig struct {
	Features struct {
		EnableAutoAssign   bool   `json:"enable_auto_assign"`
		DefaultDescription string `json:"default_description"`
		DefaultPriority    int64  `json:"default_priority"`
	} `json:"features"`
	TagMappings map[string][]string `json:"tag_mappings"`
}

// Lettura dinamica del file config
func loadConfig() *PluginConfig {
	// Di default Vikunja viene eseguito dalla root, i plugin sono in "plugins/nome"
	configPaths := []string{
		"plugins/TonnoSanitizerTagger/config.json",
		"plugins/Tonno Sanitizer Tagger/config.json",
		"TonnoSanitizerTagger/config.json",
		"config.json",
	}

	var data []byte
	var err error
	for _, p := range configPaths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}

	config := &PluginConfig{}
	if err != nil {
		log.Errorf("Impossibile leggere config.json. Usa i default. Errore: %v", err)
		return config
	}

	if err := json.Unmarshal(data, config); err != nil {
		log.Errorf("Errore nel parsing di config.json: %v", err)
	}
	return config
}

// TonnoSanitizerTaggerPlugin implements the required Vikunja extension interfaces.
type TonnoSanitizerTaggerPlugin struct{}

// Compile-time interface assertions to enforce correctness.
var (
	_ plugins.Plugin = (*TonnoSanitizerTaggerPlugin)(nil)
)

// Singleton instance deployment to prevent state duplication.
var singleton = &TonnoSanitizerTaggerPlugin{}

// Factory Entrypoints for the Vikunja Loader System.
func NewPlugin() plugins.Plugin { return singleton }

func (p *TonnoSanitizerTaggerPlugin) Name() string    { return "tonno-sanitizer-tagger-plugin" }
func (p *TonnoSanitizerTaggerPlugin) Version() string { return "1.0.0" }

func (p *TonnoSanitizerTaggerPlugin) Init() error {
	log.Infof("Initializing %s version %s", p.Name(), p.Version())

	// Sottoscrizione nativa all'evento di creazione task (Watermill / Observer Pattern)
	events.RegisterListener(func(event *models.TaskCreatedEvent) error {
		return handleTaskChange(event.Task, event.Doer)
	})

	// Sottoscrizione anche all'aggiornamento (utile se modifichi un titolo esistente)
	events.RegisterListener(func(event *models.TaskUpdatedEvent) error {
		return handleTaskChange(event.Task, event.Doer)
	})

	return nil
}

func (p *TonnoSanitizerTaggerPlugin) Shutdown() error {
	log.Infof("Shutting down %s", p.Name())
	return nil
}

// Strutture interne per l'interazione XORM senza dipendere da funzioni complesse non-Yaegi
type TaskUpdateStruct struct {
	Description string `xorm:"description"`
	Priority    int64  `xorm:"priority"`
}

type TaskTag struct {
	ID        int64 `xorm:"pk autoincr"`
	TaskID    int64 `xorm:"task_id"`
	LabelID   int64 `xorm:"label_id"` // In Vikunja, i tag sono generalmente chiamati "labels" a DB
}

type TaskAssignee struct {
	ID     int64 `xorm:"pk autoincr"`
	TaskID int64 `xorm:"task_id"`
	UserID int64 `xorm:"user_id"`
}

func handleTaskChange(task *models.Task, doer *models.User) error {
	if task == nil {
		return nil
	}

	// Regola 1: Normalizzazione
	title := strings.ToLower(task.Title)
	punctuation := []string{"!", "?", ".", ",", ":", ";"}
	for _, p := range punctuation {
		title = strings.ReplaceAll(title, p, "")
	}

	// Carichiamo le configurazioni utente da JSON
	config := loadConfig()

	// Regola 2 & 3: Mappatura dal config
	labelsToApply := make([]string, 0)
	if config.TagMappings != nil {
		for labelTitle, keywords := range config.TagMappings {
			for _, keyword := range keywords {
				if strings.Contains(title, keyword) {
					labelsToApply = append(labelsToApply, labelTitle)
					break
				}
			}
		}
	}

	// Apriamo una transazione DB pulita per le scritture
	s := db.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		return err
	}

	// Recuperiamo gli ID delle etichette
	// Struttura per interrogare le etichette (labels) nel DB
	type Label struct {
		ID    int64  `xorm:"pk autoincr"`
		Title string `xorm:"title"`
	}
	tagsToAdd := make([]int64, 0)
	for _, lTitle := range labelsToApply {
		var l Label
		has, err := s.Table("labels").Where("title = ?", lTitle).Get(&l)
		if err != nil {
			log.Errorf("[%s] Errore lookup etichetta '%s': %v", singleton.Name(), lTitle, err)
			continue
		}
		if has {
			tagsToAdd = append(tagsToAdd, l.ID)
		} else {
			// Se l'etichetta non esiste su Vikunja, logghiamo un warning e saltiamo
			log.Infof("[%s] Etichetta '%s' non trovata nel database. Creala in Vikunja per assegnarla.", singleton.Name(), lTitle)
		}
	}

	// Regola 4: Idempotenza. Evitiamo duplicati leggendo i tag già in DB
	var existingTags []TaskTag
	if err := s.Table("label_tasks").Where("task_id = ?", task.ID).Find(&existingTags); err != nil {
		s.Rollback()
		return err
	}

	existingMap := make(map[int64]bool)
	for _, et := range existingTags {
		existingMap[et.LabelID] = true
	}

	for _, tID := range tagsToAdd {
		if !existingMap[tID] {
			newTag := TaskTag{
				TaskID:  task.ID,
				LabelID: tID,
			}
			if _, err := s.Table("label_tasks").Insert(&newTag); err != nil {
				log.Errorf("[%s] Errore inserimento tag: %v", singleton.Name(), err)
			}
		}
	}

	// Regola 5: Fallback default basato su config
	changed := false
	updateStruct := TaskUpdateStruct{
		Description: task.Description,
		Priority:    task.Priority,
	}

	// Usa il default dal JSON oppure cade indietro su "Bo..."
	defDesc := config.Features.DefaultDescription
	if defDesc == "" {
		defDesc = "Bo..."
	}

	if strings.TrimSpace(task.Description) == "" {
		updateStruct.Description = defDesc
		changed = true
	}

	// Usa il default dal JSON oppure 3 (Media)
	defPrio := config.Features.DefaultPriority
	if defPrio == 0 {
		defPrio = 3
	}

	if task.Priority == 0 {
		updateStruct.Priority = defPrio
		changed = true
	}

	if changed {
		if _, err := s.Table("tasks").ID(task.ID).Cols("description", "priority").Update(&updateStruct); err != nil {
			log.Errorf("[%s] Errore aggiornamento fallback task: %v", singleton.Name(), err)
		}
	}

	// Assegnazione Automatica se abilitata dal config
	if config.Features.EnableAutoAssign {
		var assignees []TaskAssignee
		if err := s.Table("task_assignees").Where("task_id = ?", task.ID).Find(&assignees); err != nil {
			s.Rollback()
			return err
		}

		// Se il creatore è valido e non ci sono assegnatari, lo auto-assegniamo
		if len(assignees) == 0 && doer != nil && doer.ID > 0 {
			assignee := TaskAssignee{
				TaskID: task.ID,
				UserID: doer.ID,
			}
			if _, err := s.Table("task_assignees").Insert(&assignee); err != nil {
				log.Errorf("[%s] Errore auto-assegnazione task: %v", singleton.Name(), err)
			}
		}
	}

	return s.Commit()
}
