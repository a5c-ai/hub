package actions

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// WorkflowDefinition represents the complete YAML workflow structure
type WorkflowDefinition struct {
	Name        string                   `yaml:"name"`
	On          TriggerConfig            `yaml:"on"`
	Env         map[string]string        `yaml:"env,omitempty"`
	Defaults    *Defaults                `yaml:"defaults,omitempty"`
	Jobs        map[string]JobDefinition `yaml:"jobs"`
	Permissions interface{}              `yaml:"permissions,omitempty"`
}

// TriggerConfig represents the trigger configuration
type TriggerConfig struct {
	Push             *PushTrigger          `yaml:"push,omitempty"`
	PullRequest      *PullRequestTrigger   `yaml:"pull_request,omitempty"`
	Schedule         []ScheduleTrigger     `yaml:"schedule,omitempty"`
	WorkflowDispatch *WorkflowDispatch     `yaml:"workflow_dispatch,omitempty"`
	Release          *ReleaseTrigger       `yaml:"release,omitempty"`
	Issues           *IssuesTrigger        `yaml:"issues,omitempty"`
	IssueComment     *IssueCommentTrigger  `yaml:"issue_comment,omitempty"`
	Watch            *WatchTrigger         `yaml:"watch,omitempty"`
	Fork             *ForkTrigger          `yaml:"fork,omitempty"`
	Create           *CreateTrigger        `yaml:"create,omitempty"`
	Delete           *DeleteTrigger        `yaml:"delete,omitempty"`
	Repository       *RepositoryTrigger    `yaml:"repository,omitempty"`
}

// PushTrigger represents push event configuration
type PushTrigger struct {
	Branches     []string `yaml:"branches,omitempty"`
	BranchesIgnore []string `yaml:"branches-ignore,omitempty"`
	Tags         []string `yaml:"tags,omitempty"`
	TagsIgnore   []string `yaml:"tags-ignore,omitempty"`
	Paths        []string `yaml:"paths,omitempty"`
	PathsIgnore  []string `yaml:"paths-ignore,omitempty"`
}

// PullRequestTrigger represents pull request event configuration
type PullRequestTrigger struct {
	Types          []string `yaml:"types,omitempty"`
	Branches       []string `yaml:"branches,omitempty"`
	BranchesIgnore []string `yaml:"branches-ignore,omitempty"`
	Paths          []string `yaml:"paths,omitempty"`
	PathsIgnore    []string `yaml:"paths-ignore,omitempty"`
}

// ScheduleTrigger represents scheduled event configuration
type ScheduleTrigger struct {
	Cron string `yaml:"cron"`
}

// WorkflowDispatch represents manual trigger configuration
type WorkflowDispatch struct {
	Inputs map[string]WorkflowInput `yaml:"inputs,omitempty"`
}

// WorkflowInput represents an input for workflow dispatch
type WorkflowInput struct {
	Description string      `yaml:"description,omitempty"`
	Required    bool        `yaml:"required,omitempty"`
	Default     interface{} `yaml:"default,omitempty"`
	Type        string      `yaml:"type,omitempty"`
	Options     []string    `yaml:"options,omitempty"`
}

// ReleaseTrigger, IssuesTrigger, etc. - simplified for now
type ReleaseTrigger struct {
	Types []string `yaml:"types,omitempty"`
}

type IssuesTrigger struct {
	Types []string `yaml:"types,omitempty"`
}

type IssueCommentTrigger struct {
	Types []string `yaml:"types,omitempty"`
}

type WatchTrigger struct {
	Types []string `yaml:"types,omitempty"`
}

type ForkTrigger struct{}

type CreateTrigger struct{}

type DeleteTrigger struct{}

type RepositoryTrigger struct {
	Types []string `yaml:"types,omitempty"`
}

// Defaults represents default configuration
type Defaults struct {
	Run *RunDefaults `yaml:"run,omitempty"`
}

// RunDefaults represents default run configuration
type RunDefaults struct {
	Shell            string `yaml:"shell,omitempty"`
	WorkingDirectory string `yaml:"working-directory,omitempty"`
}

// JobDefinition represents a job definition
type JobDefinition struct {
	Name            string                 `yaml:"name,omitempty"`
	RunsOn          interface{}            `yaml:"runs-on"` // string or array
	Environment     interface{}            `yaml:"environment,omitempty"` // string or object
	ConcurrencyGroup string                `yaml:"concurrency,omitempty"`
	Outputs         map[string]string      `yaml:"outputs,omitempty"`
	Env             map[string]string      `yaml:"env,omitempty"`
	Defaults        *Defaults              `yaml:"defaults,omitempty"`
	If              string                 `yaml:"if,omitempty"`
	Steps           []StepDefinition       `yaml:"steps"`
	TimeoutMinutes  int                    `yaml:"timeout-minutes,omitempty"`
	Strategy        *Strategy              `yaml:"strategy,omitempty"`
	ContinueOnError bool                   `yaml:"continue-on-error,omitempty"`
	Container       interface{}            `yaml:"container,omitempty"`
	Services        map[string]interface{} `yaml:"services,omitempty"`
	Needs           interface{}            `yaml:"needs,omitempty"` // string or array
	Permissions     interface{}            `yaml:"permissions,omitempty"`
}

// Strategy represents job strategy (matrix builds, etc.)
type Strategy struct {
	Matrix          interface{} `yaml:"matrix,omitempty"`
	FailFast        *bool       `yaml:"fail-fast,omitempty"`
	MaxParallel     int         `yaml:"max-parallel,omitempty"`
}

// StepDefinition represents a step definition
type StepDefinition struct {
	ID              string            `yaml:"id,omitempty"`
	If              string            `yaml:"if,omitempty"`
	Name            string            `yaml:"name,omitempty"`
	Uses            string            `yaml:"uses,omitempty"`
	Run             string            `yaml:"run,omitempty"`
	With            map[string]interface{} `yaml:"with,omitempty"`
	Env             map[string]string `yaml:"env,omitempty"`
	ContinueOnError bool              `yaml:"continue-on-error,omitempty"`
	TimeoutMinutes  int               `yaml:"timeout-minutes,omitempty"`
	Shell           string            `yaml:"shell,omitempty"`
	WorkingDirectory string           `yaml:"working-directory,omitempty"`
}

// WorkflowParser parses GitHub Actions-compatible YAML workflows
type WorkflowParser struct{}

// NewWorkflowParser creates a new workflow parser
func NewWorkflowParser() *WorkflowParser {
	return &WorkflowParser{}
}

// Parse parses a workflow YAML content and returns a WorkflowDefinition
func (p *WorkflowParser) Parse(yamlContent string) (*WorkflowDefinition, error) {
	var workflow WorkflowDefinition
	
	if err := yaml.Unmarshal([]byte(yamlContent), &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	// Validate the workflow
	if err := p.validate(&workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	return &workflow, nil
}

// validate performs basic validation on the workflow
func (p *WorkflowParser) validate(workflow *WorkflowDefinition) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(workflow.Jobs) == 0 {
		return fmt.Errorf("workflow must have at least one job")
	}

	// Validate jobs
	for jobID, job := range workflow.Jobs {
		if err := p.validateJob(jobID, &job); err != nil {
			return fmt.Errorf("invalid job '%s': %w", jobID, err)
		}
	}

	// Validate job dependencies
	if err := p.validateJobDependencies(workflow.Jobs); err != nil {
		return fmt.Errorf("invalid job dependencies: %w", err)
	}

	return nil
}

// validateJob validates a single job
func (p *WorkflowParser) validateJob(jobID string, job *JobDefinition) error {
	if !isValidJobID(jobID) {
		return fmt.Errorf("invalid job ID '%s': must contain only alphanumeric characters, hyphens, and underscores", jobID)
	}

	if job.RunsOn == nil {
		return fmt.Errorf("runs-on is required")
	}

	if len(job.Steps) == 0 {
		return fmt.Errorf("job must have at least one step")
	}

	// Validate steps
	for i, step := range job.Steps {
		if err := p.validateStep(i, &step); err != nil {
			return fmt.Errorf("invalid step %d: %w", i+1, err)
		}
	}

	return nil
}

// validateStep validates a single step
func (p *WorkflowParser) validateStep(index int, step *StepDefinition) error {
	if step.Uses == "" && step.Run == "" {
		return fmt.Errorf("step must have either 'uses' or 'run'")
	}

	if step.Uses != "" && step.Run != "" {
		return fmt.Errorf("step cannot have both 'uses' and 'run'")
	}

	if step.Uses != "" {
		if err := p.validateActionReference(step.Uses); err != nil {
			return fmt.Errorf("invalid action reference '%s': %w", step.Uses, err)
		}
	}

	return nil
}

// validateActionReference validates an action reference (e.g., actions/checkout@v4)
func (p *WorkflowParser) validateActionReference(actionRef string) error {
	// Basic validation for action reference format
	// Format: owner/action@version or ./local-action or docker://image
	
	if strings.HasPrefix(actionRef, "./") {
		// Local action
		return nil
	}

	if strings.HasPrefix(actionRef, "docker://") {
		// Docker action
		return nil
	}

	// Standard action format: owner/action@version
	parts := strings.Split(actionRef, "@")
	if len(parts) != 2 {
		return fmt.Errorf("action reference must include version (e.g., owner/action@v1)")
	}

	actionParts := strings.Split(parts[0], "/")
	if len(actionParts) < 2 {
		return fmt.Errorf("action reference must include owner and action name (e.g., owner/action@v1)")
	}

	return nil
}

// validateJobDependencies validates job dependency graph for cycles
func (p *WorkflowParser) validateJobDependencies(jobs map[string]JobDefinition) error {
	// Build dependency graph
	dependencies := make(map[string][]string)
	
	for jobID, job := range jobs {
		var needs []string
		
		switch v := job.Needs.(type) {
		case string:
			needs = []string{v}
		case []interface{}:
			for _, need := range v {
				if needStr, ok := need.(string); ok {
					needs = append(needs, needStr)
				}
			}
		case []string:
			needs = v
		}
		
		dependencies[jobID] = needs
	}

	// Check for invalid dependencies (referencing non-existent jobs)
	for jobID, deps := range dependencies {
		for _, dep := range deps {
			if _, exists := jobs[dep]; !exists {
				return fmt.Errorf("job '%s' depends on non-existent job '%s'", jobID, dep)
			}
		}
	}

	// Check for circular dependencies using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	var hasCycle func(string) bool
	hasCycle = func(jobID string) bool {
		visited[jobID] = true
		recStack[jobID] = true
		
		for _, dep := range dependencies[jobID] {
			if !visited[dep] && hasCycle(dep) {
				return true
			} else if recStack[dep] {
				return true
			}
		}
		
		recStack[jobID] = false
		return false
	}
	
	for jobID := range jobs {
		if !visited[jobID] && hasCycle(jobID) {
			return fmt.Errorf("circular dependency detected involving job '%s'", jobID)
		}
	}

	return nil
}

// isValidJobID checks if a job ID is valid
func isValidJobID(jobID string) bool {
	// Job IDs must contain only alphanumeric characters, hyphens, and underscores
	// and cannot start with a number or hyphen
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_-]*$`, jobID)
	return matched
}

// GetTriggerEvents extracts all trigger events from the workflow
func (p *WorkflowParser) GetTriggerEvents(workflow *WorkflowDefinition) []string {
	var events []string
	
	if workflow.On.Push != nil {
		events = append(events, "push")
	}
	if workflow.On.PullRequest != nil {
		events = append(events, "pull_request")
	}
	if len(workflow.On.Schedule) > 0 {
		events = append(events, "schedule")
	}
	if workflow.On.WorkflowDispatch != nil {
		events = append(events, "workflow_dispatch")
	}
	if workflow.On.Release != nil {
		events = append(events, "release")
	}
	if workflow.On.Issues != nil {
		events = append(events, "issues")
	}
	if workflow.On.IssueComment != nil {
		events = append(events, "issue_comment")
	}
	if workflow.On.Watch != nil {
		events = append(events, "watch")
	}
	if workflow.On.Fork != nil {
		events = append(events, "fork")
	}
	if workflow.On.Create != nil {
		events = append(events, "create")
	}
	if workflow.On.Delete != nil {
		events = append(events, "delete")
	}
	if workflow.On.Repository != nil {
		events = append(events, "repository")
	}
	
	return events
}

// ShouldTrigger determines if a workflow should trigger for given event and context
func (p *WorkflowParser) ShouldTrigger(workflow *WorkflowDefinition, event string, context TriggerContext) bool {
	switch event {
	case "push":
		return p.shouldTriggerOnPush(workflow.On.Push, context)
	case "pull_request":
		return p.shouldTriggerOnPullRequest(workflow.On.PullRequest, context)
	case "schedule":
		return len(workflow.On.Schedule) > 0
	case "workflow_dispatch":
		return workflow.On.WorkflowDispatch != nil
	default:
		// For other events, just check if they're configured
		events := p.GetTriggerEvents(workflow)
		for _, e := range events {
			if e == event {
				return true
			}
		}
		return false
	}
}

// TriggerContext provides context for trigger evaluation
type TriggerContext struct {
	Branch     string
	Tag        string
	Paths      []string
	PRAction   string
	PRTargetBranch string
}

// shouldTriggerOnPush checks if workflow should trigger on push event
func (p *WorkflowParser) shouldTriggerOnPush(push *PushTrigger, context TriggerContext) bool {
	if push == nil {
		return false
	}

	// Check branches
	if len(push.Branches) > 0 {
		if !p.matchesPatterns(context.Branch, push.Branches) {
			return false
		}
	}

	if len(push.BranchesIgnore) > 0 {
		if p.matchesPatterns(context.Branch, push.BranchesIgnore) {
			return false
		}
	}

	// Check tags
	if len(push.Tags) > 0 {
		if !p.matchesPatterns(context.Tag, push.Tags) {
			return false
		}
	}

	if len(push.TagsIgnore) > 0 {
		if p.matchesPatterns(context.Tag, push.TagsIgnore) {
			return false
		}
	}

	// Check paths
	if len(push.Paths) > 0 {
		if !p.matchesAnyPath(context.Paths, push.Paths) {
			return false
		}
	}

	if len(push.PathsIgnore) > 0 {
		if p.matchesAnyPath(context.Paths, push.PathsIgnore) {
			return false
		}
	}

	return true
}

// shouldTriggerOnPullRequest checks if workflow should trigger on pull request event
func (p *WorkflowParser) shouldTriggerOnPullRequest(pr *PullRequestTrigger, context TriggerContext) bool {
	if pr == nil {
		return false
	}

	// Check PR types
	if len(pr.Types) > 0 {
		found := false
		for _, prType := range pr.Types {
			if prType == context.PRAction {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check target branches
	if len(pr.Branches) > 0 {
		if !p.matchesPatterns(context.PRTargetBranch, pr.Branches) {
			return false
		}
	}

	if len(pr.BranchesIgnore) > 0 {
		if p.matchesPatterns(context.PRTargetBranch, pr.BranchesIgnore) {
			return false
		}
	}

	// Check paths
	if len(pr.Paths) > 0 {
		if !p.matchesAnyPath(context.Paths, pr.Paths) {
			return false
		}
	}

	if len(pr.PathsIgnore) > 0 {
		if p.matchesAnyPath(context.Paths, pr.PathsIgnore) {
			return false
		}
	}

	return true
}

// matchesPatterns checks if a value matches any of the given patterns
func (p *WorkflowParser) matchesPatterns(value string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := p.matchPattern(value, pattern); matched {
			return true
		}
	}
	return false
}

// matchesAnyPath checks if any of the paths match any of the patterns
func (p *WorkflowParser) matchesAnyPath(paths []string, patterns []string) bool {
	for _, path := range paths {
		if p.matchesPatterns(path, patterns) {
			return true
		}
	}
	return false
}

// matchPattern performs glob-style pattern matching
func (p *WorkflowParser) matchPattern(value, pattern string) (bool, error) {
	// Simple glob pattern matching - in production, use a proper glob library
	if pattern == value {
		return true, nil
	}

	if strings.Contains(pattern, "*") {
		// Convert glob pattern to regex
		regexPattern := strings.ReplaceAll(pattern, "*", ".*")
		regexPattern = "^" + regexPattern + "$"
		matched, err := regexp.MatchString(regexPattern, value)
		return matched, err
	}

	return value == pattern, nil
}