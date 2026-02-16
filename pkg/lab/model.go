package lab

import "time"

type Status string

const (
	StatusOK          Status = "ok"
	StatusDegraded    Status = "degraded"
	StatusUnavailable Status = "unavailable"
)

type Overview struct {
	UpdatedAt time.Time
	Machines  []Machine
	Agents    AgentSummary
	Training  TrainingSummary
	Models    []HFModel
	Commits   []Commit
	Errors    map[string]string
}

type Machine struct {
	Name       string
	Host       string
	Status     Status
	Load1      float64
	MemUsedPct float64
	Containers []Container
	// Extended stats
	CPUCores     int
	MemTotalGB   float64
	MemUsedGB    float64
	DiskTotalGB  float64
	DiskUsedGB   float64
	DiskUsedPct  float64
	GPUName      string
	GPUVRAMTotal float64 // GB, 0 if not applicable
	GPUVRAMUsed  float64
	GPUVRAMPct   float64
	GPUTemp      int // Celsius, 0 if unavailable
	Uptime       string
}

type Container struct {
	Name    string
	Status  string
	Image   string
	Uptime  string
	Created time.Time
}

type AgentSummary struct {
	Available       bool
	RegisteredTotal int
	QueuePending    int
	TasksCompleted  int
	TasksFailed     int
	Capabilities    int
	HeartbeatAge    float64
	ExporterUp      bool
}

type TrainingSummary struct {
	GoldGenerated  int
	GoldTarget     int
	GoldPercent    float64
	GoldAvailable  bool
	InterceptCount int
	SessionCount   int
	LastIntercept  time.Time
	GGUFCount      int
	GGUFFiles      []string
	AdapterCount   int
}

type HFModel struct {
	ModelID      string    `json:"modelId"`
	Author       string    `json:"author"`
	Downloads    int       `json:"downloads"`
	Likes        int       `json:"likes"`
	Tags         []string  `json:"tags"`
	PipelineTag  string    `json:"pipeline_tag"`
	CreatedAt    time.Time `json:"createdAt"`
	LastModified time.Time `json:"lastModified"`
}

type Commit struct {
	SHA       string
	Message   string
	Author    string
	Repo      string
	Timestamp time.Time
}

type Service struct {
	Name     string
	URL      string
	Category string
	Machine  string
	Icon     string
	Status   string // ok, degraded, unavailable, unchecked
}

// Dataset stats from DuckDB (pushed to InfluxDB as dataset_stats).

type DatasetTable struct {
	Name string
	Rows int
}

type DatasetSummary struct {
	Available bool
	Tables    []DatasetTable
	UpdatedAt time.Time
}

// Golden set data explorer types.

type GoldenSetSummary struct {
	Available        bool
	TotalExamples    int
	TargetTotal      int
	CompletionPct    float64
	Domains          int
	Voices           int
	AvgGenTime       float64
	AvgResponseChars float64
	DomainStats      []DomainStat
	VoiceStats       []VoiceStat
	Workers          []WorkerStat
	UpdatedAt        time.Time
}

type WorkerStat struct {
	Worker   string
	Count    int
	LastSeen time.Time
}

type DomainStat struct {
	Domain     string
	Count      int
	AvgGenTime float64
}

type VoiceStat struct {
	Voice      string
	Count      int
	AvgChars   float64
	AvgGenTime float64
}

// Live training run status (from InfluxDB training_status measurement).

type TrainingRunStatus struct {
	Model      string
	RunID      string
	Status     string  // training, fusing, complete, failed
	Iteration  int
	TotalIters int
	Pct        float64
	LastLoss   float64 // most recent train loss
	ValLoss    float64 // most recent val loss
	TokensSec  float64 // most recent tokens/sec
}

// Benchmark data types for training run viewer.

type BenchmarkRun struct {
	RunID string
	Model string
	Type  string // "content", "capability", "training"
}

type LossPoint struct {
	Iteration    int
	Loss         float64
	LossType     string // "val" or "train"
	LearningRate float64
	TokensPerSec float64
}

type ContentPoint struct {
	Label     string
	Dimension string
	Score     float64
	Iteration int
	HasKernel bool
}

type CapabilityPoint struct {
	Label    string
	Category string
	Accuracy float64
	Correct  int
	Total    int
	Iteration int
}

type CapabilityJudgePoint struct {
	Label       string
	ProbeID     string
	Category    string
	Reasoning   float64
	Correctness float64
	Clarity     float64
	Avg         float64
	Iteration   int
}

type BenchmarkData struct {
	Runs            []BenchmarkRun
	Loss            map[string][]LossPoint
	Content         map[string][]ContentPoint
	Capability      map[string][]CapabilityPoint
	CapabilityJudge map[string][]CapabilityJudgePoint
	UpdatedAt       time.Time
}
