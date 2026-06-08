package main

import (
	"strings"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/plugin"
	"code.vikunja.io/api/pkg/user"
)

// Hardcoded configuration instead of external JSON file
var config = struct {
	EnableAutoAssign   bool
	DefaultDescription string
	DefaultPriority    int64
	TagMappings        map[string][]string
}{
	EnableAutoAssign:   true,
	DefaultDescription: "Bo...",
	DefaultPriority:    3,
	TagMappings: map[string][]string{
		"Fix o Bug": {"bug", "error", "fix", "crash", "problem", "errore", "problema"},
		"Spesa":     {"buy", "expense", "get", "order", "comprare", "spesa", "prendere", "acquistare", "ordinare"},
	},
}

// TonnoSanitizerTaggerPlugin implements the required Vikunja extension interfaces.
type TonnoSanitizerTaggerPlugin struct{}

// Compile-time interface assertions to enforce correctness.
var (
	_ plugin.Plugin = (*TonnoSanitizerTaggerPlugin)(nil)
)

// Singleton instance deployment to prevent state duplication.
var singleton = &TonnoSanitizerTaggerPlugin{}

// Factory Entrypoint for the Vikunja Loader System.
func NewPlugin() plugin.Plugin { return singleton }

func (p *TonnoSanitizerTaggerPlugin) Name() string    { return "tonno-sanitizer-tagger-plugin" }
func (p *TonnoSanitizerTaggerPlugin) Version() string { return "1.0.0" }

func (p *TonnoSanitizerTaggerPlugin) Init() error {
	log.Infof("Initializing %s version %s", p.Name(), p.Version())

	// Native subscription to the task creation event
	events.RegisterListener(func(event *models.TaskCreatedEvent) error {
		return handleTaskChange(event.Task, event.Doer)
	})

	// Subscription to task updates
	events.RegisterListener(func(event *models.TaskUpdatedEvent) error {
		return handleTaskChange(event.Task, event.Doer)
	})

	return nil
}

func (p *TonnoSanitizerTaggerPlugin) Shutdown() error {
	log.Infof("Shutting down %s", p.Name())
	return nil
}

func handleTaskChange(task *models.Task, doer *user.User) error {
	if task == nil {
		return nil
	}

	// Rule 1: Normalization (lowercase and simple punctuation removal)
	title := strings.ToLower(task.Title)
	punctuation := []string{"!", "?", ".", ",", ":", ";"}
	for _, p := range punctuation {
		title = strings.ReplaceAll(title, p, "")
	}

	// Rule 2 & 3: Mapping from configuration
	if config.TagMappings != nil {
		for labelTitle, keywords := range config.TagMappings {
			for _, keyword := range keywords {
				if strings.Contains(title, keyword) {
					// Check if label already exists in task.Labels to maintain idempotency
					alreadyExists := false
					for _, l := range task.Labels {
						if l != nil && strings.ToLower(l.Title) == strings.ToLower(labelTitle) {
							alreadyExists = true
							break
						}
					}
					if !alreadyExists {
						// Add the label in-memory. Vikunja core will automatically link it.
						task.Labels = append(task.Labels, &models.Label{Title: labelTitle})
					}
					break // Keyword found, no need to search for others for this label
				}
			}
		}
	}

	// Rule 4: Fallback defaults based on configuration
	defDesc := config.DefaultDescription
	if defDesc == "" {
		defDesc = "Bo..."
	}

	if strings.TrimSpace(task.Description) == "" {
		task.Description = defDesc
	}

	defPrio := config.DefaultPriority
	if defPrio == 0 {
		defPrio = 3
	}

	if task.Priority == 0 {
		task.Priority = defPrio
	}

	// Auto Assignment if enabled
	if config.EnableAutoAssign {
		// If creator is valid and no assignees, auto-assign
		if len(task.Assignees) == 0 && doer != nil && doer.ID > 0 {
			task.Assignees = append(task.Assignees, doer)
		}
	}

	return nil
}
