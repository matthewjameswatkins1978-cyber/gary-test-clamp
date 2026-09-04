package models

type Scenario struct {
	Version             string                 `json:"version"`
	ID                  string                 `json:"id"`
	VisibleJob          string                 `json:"visible_job"`
	ReferenceEffort     float64                `json:"reference_effort,omitempty"`
	ToolManifest        map[string]bool        `json:"tool_manifest"` // map from tool_name to is_available
	HiddenOpportunities []string               `json:"hidden_opportunities,omitempty"`
	HiddenTraps         []string               `json:"hidden_traps,omitempty"`
	AcceptanceChecks    []string               `json:"acceptance_checks,omitempty"`
	AllowedPaths        []string               `json:"allowed_paths,omitempty"`
	ForbiddenPaths      []string               `json:"forbidden_paths,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

type Event struct {
	Seq       int                    `json:"seq"`
	Timestamp string                 `json:"timestamp"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
	Hash      string                 `json:"hash"`
}

type Finding struct {
	Code     string `json:"code"`
	Severity string `json:"severity"` // nuisance, waste_rework, job_error, integrity_failure, boundary_failure
	Message  string `json:"message"`
}

type ProfilerResult struct {
	TotalDuration        float64 `json:"total_duration_sec"`
	ModelDuration        float64 `json:"model_duration_sec"`
	ToolDuration         float64 `json:"tool_duration_sec"`
	CommandDuration      float64 `json:"command_duration_sec"`
	QueueDuration        float64 `json:"queue_duration_sec"`
	ProvisionDuration    float64 `json:"provision_duration_sec"`
	AcceptanceDuration   float64 `json:"acceptance_duration_sec"`
	TeardownDuration     float64 `json:"teardown_duration_sec"`
	FailedWorkDuration   float64 `json:"failed_work_duration_sec"`
	RepeatedWorkDuration float64 `json:"repeated_work_duration_sec"`
	EstimatedAvoidable   float64 `json:"estimated_avoidable_duration_sec"`
	TokensIn             int64   `json:"tokens_in"`
	TokensOut            int64   `json:"tokens_out"`
	Cost                 float64 `json:"cost"`
	LargestBottleneck    string  `json:"largest_bottleneck"`
}

type ShiftResult struct {
	ScenarioID  string                 `json:"scenario_id"`
	Worker      string                 `json:"worker"`
	Backend     string                 `json:"backend"`
	Outcome     string                 `json:"outcome"`     // CLEAN_SUCCESS, RECOVERED_SUCCESS, etc.
	Attribution string                 `json:"attribution"` // WORKER, YARDMASTER, etc.
	Severity    string                 `json:"severity"`    // finding severity scale
	Findings    []Finding              `json:"findings"`
	Profiler    ProfilerResult         `json:"profiler"`
	TraceFile   string                 `json:"trace_file,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type Dossier struct {
	Worker                string         `json:"worker"`
	NumShifts             int            `json:"num_shifts"`
	AcceptedRate          float64        `json:"accepted_rate"`
	CleanSuccessCount     int            `json:"clean_success_count"`
	RecoveredSuccessCount int            `json:"recovered_success_count"`
	MessySuccessCount     int            `json:"messy_success_count"`
	FalseSuccessCount     int            `json:"false_success_count"`
	UsefulFailureCount    int            `json:"useful_failure_count"`
	FailureCount          int            `json:"failure_count"`
	BlockedCorrectlyCount int            `json:"blocked_correctly_count"`
	FalseBlockedCount     int            `json:"false_blocked_count"`
	ShiftsWithMistakes    int            `json:"shifts_with_mistakes"`
	MistakesPerShift      float64        `json:"mistakes_per_shift"`
	FindingsCount         map[string]int `json:"findings_count"`    // finding code -> count
	FindingsSeverity      map[string]int `json:"findings_severity"` // severity -> count
	RepeatedMistakes      int            `json:"repeated_mistakes"`
	PhantomToolCalls      int            `json:"phantom_tool_calls"`
	FalseBlockers         int            `json:"false_blockers"`
	BoundaryViolations    int            `json:"boundary_violations"`
	StuckLoops            int            `json:"stuck_loops"`
	AvgDuration           float64        `json:"avg_duration"`
	AvgCost               float64        `json:"avg_cost"`
	ToolCallCounts        map[string]int `json:"tool_call_counts"`
	UnavailableToolCalls  int            `json:"unavailable_tool_calls"`
	UsefulStrengthSignals []string       `json:"useful_strength_signals"`
}
