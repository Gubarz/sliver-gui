package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bishopfox/sliver/client/assets"
	consts "github.com/bishopfox/sliver/client/constants"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/commonpb"
	"github.com/google/uuid"
	"github.com/grafana/sobek"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"google.golang.org/protobuf/proto"
)

const (
	automationHistoryLimit = 500
	beaconPollInterval     = 5 * time.Second
)

type AutomationFilter struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

type AutomationRule struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	Enabled         bool             `json:"enabled"`
	Trigger         string           `json:"trigger"`
	TargetKind      string           `json:"targetKind"`
	Filter          AutomationFilter `json:"filter"`
	ExecutionMode   string           `json:"executionMode"`
	Commands        []string         `json:"commands"`
	Script          string           `json:"script"`
	TimeoutSeconds  int              `json:"timeoutSeconds"`
	ContinueOnError bool             `json:"continueOnError"`
	DelaySeconds    int              `json:"delaySeconds"`
	CooldownSeconds int              `json:"cooldownSeconds"`
	IntervalSeconds int              `json:"intervalSeconds"`
	MaxRuns         int              `json:"maxRuns"`
	RunCount        int              `json:"runCount"`
	CreatedAt       int64            `json:"createdAt"`
	UpdatedAt       int64            `json:"updatedAt"`
}

type AutomationRun struct {
	ID         string   `json:"id"`
	RuleID     string   `json:"ruleId"`
	RuleName   string   `json:"ruleName"`
	Trigger    string   `json:"trigger"`
	TargetID   string   `json:"targetId"`
	TargetName string   `json:"targetName"`
	TargetKind string   `json:"targetKind"`
	Commands   []string `json:"commands"`
	Output     string   `json:"output"`
	Error      string   `json:"error"`
	Status     string   `json:"status"`
	StartedAt  int64    `json:"startedAt"`
	FinishedAt int64    `json:"finishedAt"`
}

type automationTarget struct {
	ID       string
	Name     string
	Hostname string
	Username string
	OS       string
	Arch     string
	Kind     string
}

type automationState struct {
	Rules   []AutomationRule `json:"rules"`
	History []AutomationRun  `json:"history"`
}

type automationEngine struct {
	app  *App
	path string

	mu             sync.RWMutex
	rules          []AutomationRule
	history        []AutomationRun
	running        map[string]bool
	activeByRule   map[string]int
	lastRun        map[string]time.Time
	lastInterval   map[string]time.Time
	beaconCheckins map[string]int64
	beaconsPrimed  bool
}

func newAutomationEngine(app *App) *automationEngine {
	engine := &automationEngine{
		app:            app,
		path:           filepath.Join(assets.GetRootAppDir(), "gui-automation.json"),
		running:        map[string]bool{},
		activeByRule:   map[string]int{},
		lastRun:        map[string]time.Time{},
		lastInterval:   map[string]time.Time{},
		beaconCheckins: map[string]int64{},
	}
	if err := engine.load(); err != nil {
		log.Printf("automation: could not load state: %v", err)
	}
	return engine
}

func (e *automationEngine) start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(beaconPollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				e.tick(now)
			}
		}
	}()
}

func (e *automationEngine) load() error {
	data, err := os.ReadFile(e.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var state automationState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("decode %s: %w", e.path, err)
	}
	e.rules = state.Rules
	e.history = state.History
	return nil
}

func (e *automationEngine) persistLocked() error {
	state := automationState{Rules: e.rules, History: e.history}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	temp := e.path + ".tmp"
	if err := os.WriteFile(temp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(temp, e.path)
}

func (e *automationEngine) tick(now time.Time) {
	if !e.app.connected {
		return
	}
	e.pollBeaconCheckins()

	e.mu.RLock()
	rules := append([]AutomationRule(nil), e.rules...)
	e.mu.RUnlock()
	for _, rule := range rules {
		if !rule.Enabled || rule.Trigger != "interval" {
			continue
		}
		interval := time.Duration(rule.IntervalSeconds) * time.Second
		if interval < 10*time.Second {
			interval = 10 * time.Second
		}
		e.mu.Lock()
		last := e.lastInterval[rule.ID]
		if last.IsZero() {
			e.lastInterval[rule.ID] = now
			e.mu.Unlock()
			continue
		}
		if now.Sub(last) < interval {
			e.mu.Unlock()
			continue
		}
		e.lastInterval[rule.ID] = now
		e.mu.Unlock()
		e.dispatchRule(rule, "interval", nil)
	}
}

func (e *automationEngine) pollBeaconCheckins() {
	beacons, err := e.app.rpcClient.GetBeacons(context.Background(), &commonpb.Empty{})
	if err != nil {
		return
	}
	e.mu.RLock()
	previous := make(map[string]int64, len(e.beaconCheckins))
	for id, checkin := range e.beaconCheckins {
		previous[id] = checkin
	}
	primed := e.beaconsPrimed
	e.mu.RUnlock()

	current := make(map[string]int64, len(beacons.Beacons))
	for _, beacon := range beacons.Beacons {
		current[beacon.ID] = beacon.LastCheckin
		if primed && previous[beacon.ID] != 0 && beacon.LastCheckin > previous[beacon.ID] {
			e.dispatchTrigger("beacon-checkin", targetFromBeacon(beacon))
		}
	}
	e.mu.Lock()
	e.beaconCheckins = current
	e.beaconsPrimed = true
	e.mu.Unlock()
}

func (e *automationEngine) handleSliverEvent(event *clientpb.Event) {
	switch event.EventType {
	case consts.SessionOpenedEvent:
		if event.Session != nil {
			e.dispatchTrigger("session-connected", targetFromSession(event.Session))
		}
	case consts.BeaconRegisteredEvent:
		beacon := &clientpb.Beacon{}
		if len(event.Data) > 0 && proto.Unmarshal(event.Data, beacon) == nil && beacon.ID != "" {
			e.mu.Lock()
			e.beaconCheckins[beacon.ID] = beacon.LastCheckin
			e.mu.Unlock()
			e.dispatchTrigger("beacon-registered", targetFromBeacon(beacon))
		}
	}
}

func (e *automationEngine) dispatchTrigger(trigger string, target automationTarget) {
	e.mu.RLock()
	rules := append([]AutomationRule(nil), e.rules...)
	e.mu.RUnlock()
	for _, rule := range rules {
		if rule.Enabled && rule.Trigger == trigger {
			e.dispatchRule(rule, trigger, &target)
		}
	}
}

func (e *automationEngine) dispatchRule(rule AutomationRule, trigger string, target *automationTarget) {
	if rule.MaxRuns > 0 && rule.RunCount >= rule.MaxRuns {
		return
	}
	if target != nil {
		if matchesAutomationRule(rule, *target) {
			e.queueRun(rule, trigger, *target)
		}
		return
	}
	targets, err := e.currentTargets()
	if err != nil {
		e.recordDispatchError(rule, trigger, err)
		return
	}
	for _, candidate := range targets {
		if matchesAutomationRule(rule, candidate) {
			e.queueRun(rule, trigger, candidate)
		}
	}
}

func (e *automationEngine) currentTargets() ([]automationTarget, error) {
	if !e.app.connected {
		return nil, fmt.Errorf("not connected")
	}
	ctx := context.Background()
	sessions, sessionErr := e.app.rpcClient.GetSessions(ctx, &commonpb.Empty{})
	beacons, beaconErr := e.app.rpcClient.GetBeacons(ctx, &commonpb.Empty{})
	if sessionErr != nil && beaconErr != nil {
		return nil, fmt.Errorf("list targets: sessions: %v; beacons: %v", sessionErr, beaconErr)
	}
	var targets []automationTarget
	if sessionErr == nil {
		for _, session := range sessions.Sessions {
			targets = append(targets, targetFromSession(session))
		}
	}
	if beaconErr == nil {
		for _, beacon := range beacons.Beacons {
			targets = append(targets, targetFromBeacon(beacon))
		}
	}
	return targets, nil
}

func (e *automationEngine) queueRun(rule AutomationRule, trigger string, target automationTarget) {
	key := rule.ID + ":" + target.ID
	now := time.Now()
	e.mu.Lock()
	if e.running[key] {
		e.mu.Unlock()
		return
	}
	if cooldown := time.Duration(rule.CooldownSeconds) * time.Second; cooldown > 0 && now.Sub(e.lastRun[key]) < cooldown {
		e.mu.Unlock()
		return
	}
	current := e.ruleByIDLocked(rule.ID)
	if current == nil || (current.MaxRuns > 0 && current.RunCount+e.activeByRule[rule.ID] >= current.MaxRuns) {
		e.mu.Unlock()
		return
	}
	e.running[key] = true
	e.activeByRule[rule.ID]++
	e.lastRun[key] = now
	e.mu.Unlock()
	go e.execute(rule, trigger, target, key)
}

func (e *automationEngine) execute(rule AutomationRule, trigger string, target automationTarget, key string) {
	run := AutomationRun{
		ID:         uuid.NewString(),
		RuleID:     rule.ID,
		RuleName:   rule.Name,
		Trigger:    trigger,
		TargetID:   target.ID,
		TargetName: displayTargetName(target),
		TargetKind: target.Kind,
		Status:     "running",
		StartedAt:  time.Now().UnixMilli(),
	}
	e.storeRun(run)

	var output string
	var runErr error
	if automationExecutionMode(rule) == "javascript" {
		output, run.Commands, runErr = e.executeJavaScript(rule, trigger, target)
	} else {
		output, run.Commands, runErr = e.executeCommands(rule, target)
	}

	run.Output = output
	run.FinishedAt = time.Now().UnixMilli()
	if runErr != nil {
		run.Status = "failed"
		run.Error = runErr.Error()
	} else {
		run.Status = "completed"
	}

	e.mu.Lock()
	delete(e.running, key)
	if e.activeByRule[rule.ID] > 0 {
		e.activeByRule[rule.ID]--
	}
	if current := e.ruleByIDLocked(rule.ID); current != nil {
		current.RunCount++
		current.UpdatedAt = time.Now().UnixMilli()
	}
	e.replaceRunLocked(run)
	if err := e.persistLocked(); err != nil {
		log.Printf("automation: persist run: %v", err)
	}
	e.mu.Unlock()
	e.emit("automation-run", run)
}

func (e *automationEngine) storeRun(run AutomationRun) {
	e.mu.Lock()
	e.history = append([]AutomationRun{run}, e.history...)
	if len(e.history) > automationHistoryLimit {
		e.history = e.history[:automationHistoryLimit]
	}
	if err := e.persistLocked(); err != nil {
		log.Printf("automation: persist started run: %v", err)
	}
	e.mu.Unlock()
	e.emit("automation-run", run)
}

func (e *automationEngine) replaceRunLocked(run AutomationRun) {
	for index := range e.history {
		if e.history[index].ID == run.ID {
			e.history[index] = run
			return
		}
	}
	e.history = append([]AutomationRun{run}, e.history...)
}

func (e *automationEngine) recordDispatchError(rule AutomationRule, trigger string, err error) {
	now := time.Now().UnixMilli()
	run := AutomationRun{
		ID: uuid.NewString(), RuleID: rule.ID, RuleName: rule.Name, Trigger: trigger,
		Status: "failed", Error: err.Error(), StartedAt: now, FinishedAt: now,
	}
	e.storeRun(run)
}

func (e *automationEngine) emit(name string, payload interface{}) {
	if e.app.ctx != nil {
		runtime.EventsEmit(e.app.ctx, name, payload)
	}
}

func (e *automationEngine) ruleByIDLocked(id string) *AutomationRule {
	for index := range e.rules {
		if e.rules[index].ID == id {
			return &e.rules[index]
		}
	}
	return nil
}

func targetFromSession(session *clientpb.Session) automationTarget {
	return automationTarget{
		ID: session.ID, Name: session.Name, Hostname: session.Hostname,
		Username: session.Username, OS: session.OS, Arch: session.Arch, Kind: "session",
	}
}

func targetFromBeacon(beacon *clientpb.Beacon) automationTarget {
	return automationTarget{
		ID: beacon.ID, Name: beacon.Name, Hostname: beacon.Hostname,
		Username: beacon.Username, OS: beacon.OS, Arch: beacon.Arch, Kind: "beacon",
	}
}

func displayTargetName(target automationTarget) string {
	if target.Name != "" {
		return target.Name
	}
	if target.Hostname != "" {
		return target.Hostname
	}
	return target.ID
}

func matchesAutomationRule(rule AutomationRule, target automationTarget) bool {
	if rule.TargetKind != "" && rule.TargetKind != "any" && rule.TargetKind != target.Kind {
		return false
	}
	return matchAutomationPattern(target.OS, rule.Filter.OS) &&
		matchAutomationPattern(target.Arch, rule.Filter.Arch) &&
		matchAutomationPattern(target.Hostname, rule.Filter.Hostname) &&
		matchAutomationPattern(target.Username, rule.Filter.Username) &&
		matchAutomationPattern(target.Name, rule.Filter.Name)
}

func matchAutomationPattern(value, patterns string) bool {
	patterns = strings.TrimSpace(patterns)
	if patterns == "" || patterns == "*" {
		return true
	}
	value = strings.ToLower(value)
	for _, candidate := range strings.Split(patterns, ",") {
		candidate = strings.ToLower(strings.TrimSpace(candidate))
		if candidate == "" {
			continue
		}
		if matched, err := filepath.Match(candidate, value); err == nil && matched {
			return true
		}
	}
	return false
}

func renderAutomationCommand(command string, target automationTarget) string {
	replacer := strings.NewReplacer(
		"{{id}}", target.ID,
		"{{name}}", target.Name,
		"{{hostname}}", target.Hostname,
		"{{username}}", target.Username,
		"{{os}}", target.OS,
		"{{arch}}", target.Arch,
		"{{kind}}", target.Kind,
	)
	return replacer.Replace(command)
}

func automationExecutionMode(rule AutomationRule) string {
	if rule.ExecutionMode == "javascript" {
		return "javascript"
	}
	return "commands"
}

func validateAutomationRule(rule AutomationRule) error {
	rule.Name = strings.TrimSpace(rule.Name)
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	switch rule.Trigger {
	case "session-connected", "beacon-registered", "beacon-checkin", "interval", "manual":
	default:
		return fmt.Errorf("unsupported trigger %q", rule.Trigger)
	}
	switch rule.TargetKind {
	case "", "any", "session", "beacon":
	default:
		return fmt.Errorf("unsupported target kind %q", rule.TargetKind)
	}
	switch automationExecutionMode(rule) {
	case "javascript":
		if strings.TrimSpace(rule.Script) == "" {
			return fmt.Errorf("JavaScript source is required")
		}
		if _, err := sobek.Compile(rule.Name+".js", rule.Script, true); err != nil {
			return fmt.Errorf("JavaScript: %w", err)
		}
		return nil
	case "commands":
		for _, command := range rule.Commands {
			if strings.TrimSpace(command) != "" {
				return nil
			}
		}
		return fmt.Errorf("at least one command is required")
	default:
		return fmt.Errorf("unsupported execution mode %q", rule.ExecutionMode)
	}
}

func (a *App) ListAutomationRules() ([]AutomationRule, error) {
	if a.automation == nil {
		return nil, fmt.Errorf("automation engine is unavailable")
	}
	a.automation.mu.RLock()
	defer a.automation.mu.RUnlock()
	rules := append([]AutomationRule(nil), a.automation.rules...)
	sort.SliceStable(rules, func(i, j int) bool {
		return strings.ToLower(rules[i].Name) < strings.ToLower(rules[j].Name)
	})
	return rules, nil
}

func (a *App) SaveAutomationRule(rule AutomationRule) (AutomationRule, error) {
	if a.automation == nil {
		return AutomationRule{}, fmt.Errorf("automation engine is unavailable")
	}
	rule.Name = strings.TrimSpace(rule.Name)
	rule.Description = strings.TrimSpace(rule.Description)
	rule.TargetKind = strings.TrimSpace(rule.TargetKind)
	rule.ExecutionMode = automationExecutionMode(rule)
	if rule.ExecutionMode == "commands" {
		rule.Commands = compactCommands(rule.Commands)
	}
	if err := validateAutomationRule(rule); err != nil {
		return AutomationRule{}, err
	}
	now := time.Now().UnixMilli()
	a.automation.mu.Lock()
	existing := a.automation.ruleByIDLocked(rule.ID)
	if existing == nil {
		rule.ID = uuid.NewString()
		rule.CreatedAt = now
		rule.RunCount = 0
		a.automation.rules = append(a.automation.rules, rule)
	} else {
		rule.CreatedAt = existing.CreatedAt
		rule.RunCount = existing.RunCount
		*existing = rule
	}
	rule.UpdatedAt = now
	if current := a.automation.ruleByIDLocked(rule.ID); current != nil {
		current.UpdatedAt = now
		rule = *current
	}
	if err := a.automation.persistLocked(); err != nil {
		a.automation.mu.Unlock()
		return AutomationRule{}, err
	}
	a.automation.mu.Unlock()
	a.automation.emit("automation-updated", rule)
	return rule, nil
}

func compactCommands(commands []string) []string {
	result := make([]string, 0, len(commands))
	for _, command := range commands {
		if command = strings.TrimSpace(command); command != "" {
			result = append(result, command)
		}
	}
	return result
}

func (a *App) DeleteAutomationRule(id string) error {
	if a.automation == nil {
		return fmt.Errorf("automation engine is unavailable")
	}
	a.automation.mu.Lock()
	for index := range a.automation.rules {
		if a.automation.rules[index].ID == id {
			a.automation.rules = append(a.automation.rules[:index], a.automation.rules[index+1:]...)
			err := a.automation.persistLocked()
			a.automation.mu.Unlock()
			if err == nil {
				a.automation.emit("automation-updated", map[string]string{"deleted": id})
			}
			return err
		}
	}
	a.automation.mu.Unlock()
	return fmt.Errorf("automation rule not found: %s", id)
}

func (a *App) SetAutomationRuleEnabled(id string, enabled bool) error {
	if a.automation == nil {
		return fmt.Errorf("automation engine is unavailable")
	}
	a.automation.mu.Lock()
	rule := a.automation.ruleByIDLocked(id)
	if rule == nil {
		a.automation.mu.Unlock()
		return fmt.Errorf("automation rule not found: %s", id)
	}
	rule.Enabled = enabled
	rule.UpdatedAt = time.Now().UnixMilli()
	saved := *rule
	err := a.automation.persistLocked()
	a.automation.mu.Unlock()
	if err == nil {
		a.automation.emit("automation-updated", saved)
	}
	return err
}

func (a *App) RunAutomationRule(id, targetID string) error {
	if a.automation == nil {
		return fmt.Errorf("automation engine is unavailable")
	}
	a.automation.mu.RLock()
	rule := a.automation.ruleByIDLocked(id)
	if rule == nil {
		a.automation.mu.RUnlock()
		return fmt.Errorf("automation rule not found: %s", id)
	}
	copyRule := *rule
	a.automation.mu.RUnlock()
	if targetID == "" {
		targets, err := a.automation.currentTargets()
		if err != nil {
			return err
		}
		matched := 0
		for _, target := range targets {
			if matchesAutomationRule(copyRule, target) {
				matched++
				a.automation.queueRun(copyRule, "manual", target)
			}
		}
		if matched == 0 {
			return fmt.Errorf("no current targets match this rule")
		}
		return nil
	}
	session, beacon, err := a.findTarget(targetID)
	if err != nil {
		return err
	}
	var target automationTarget
	if session != nil {
		target = targetFromSession(session)
	} else {
		target = targetFromBeacon(beacon)
	}
	if !matchesAutomationRule(copyRule, target) {
		return fmt.Errorf("target does not match this rule's filters")
	}
	a.automation.queueRun(copyRule, "manual", target)
	return nil
}

func (a *App) GetAutomationHistory() ([]AutomationRun, error) {
	if a.automation == nil {
		return nil, fmt.Errorf("automation engine is unavailable")
	}
	a.automation.mu.RLock()
	defer a.automation.mu.RUnlock()
	return append([]AutomationRun(nil), a.automation.history...), nil
}

func (a *App) ClearAutomationHistory() error {
	if a.automation == nil {
		return fmt.Errorf("automation engine is unavailable")
	}
	a.automation.mu.Lock()
	a.automation.history = nil
	err := a.automation.persistLocked()
	a.automation.mu.Unlock()
	if err == nil {
		a.automation.emit("automation-run", map[string]bool{"cleared": true})
	}
	return err
}
