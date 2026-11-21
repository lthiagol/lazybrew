package brew

import (
	"time"
)

// Formula represents an installed Homebrew formula
type Formula struct {
	Name           string
	FullName       string // e.g., "homebrew/core/neovim"
	Tap            string
	Version        string
	Description    string
	Homepage       string
	License        string
	Pinned         bool
	Outdated       bool
	NewVersion     string // populated when outdated
	InstalledOn    time.Time
	Dependencies   []string
	BuildDeps      []string
	Caveats        string
	KegOnly        bool
	Bottled        bool
	InstalledOnReq bool // installed on request vs as dep
	InstalledAsDep bool
	InstallPath    string
	// 6.0.0 additions
	Aliases             []string
	Binaries            []string // executables installed by this formula
	InstalledDependents []string // reverse deps (installed only)
	ListVersions        []string // all installed version history
	Revision            int      // formula revision
	Shadowed            bool     // PATH shadowing warning
}

// InstalledVersion is one entry of the "installed" array in
// `brew info --json=v2` payloads.
type InstalledVersion struct {
	Version string `json:"version"`
}

// Cask represents an installed Homebrew cask
type Cask struct {
	// Name is the token brew acts on (claude-code). DisplayName is the
	// human title brew lists under "name" (Claude Code).
	Name        string   `json:"name"`
	DisplayName string   `json:"-"`
	FullName    string   `json:"full_name"`
	Tap         string   `json:"tap"`
	Version     string   `json:"version"`
	Description string   `json:"desc"`
	Homepage    string   `json:"homepage"`
	Outdated    bool     `json:"outdated"`
	NewVersion  string   `json:"new_version"`
	Artifacts   []string `json:"artifacts"` // app names
	AutoUpdates bool     `json:"auto_updates"`
	// 6.0.0 additions
	Pinned bool   `json:"pinned,omitempty"` // cask pinning
	Sha256 string `json:"sha256,omitempty"` // checksum
	URL    string `json:"url,omitempty"`    // download URL
	// DependsOn names every formula and cask the cask depends on. The split
	// lists and the requirements come from brew's object shape of depends_on
	// (v2-05 AC-02.2); they are filled by Cask.UnmarshalJSON.
	DependsOn         []string `json:"depends_on,omitempty"`
	DependsOnFormulae []string `json:"-"`
	DependsOnCasks    []string `json:"-"`
	// Requirements reads like "macOS ≥ 13" or "arch arm64".
	Requirements []string `json:"-"`

	// Installed holds the "installed" entries of a `brew info --json=v2`
	// payload; a non-empty slice means the cask is installed (r1
	// finding 22 — search info used to hardcode casks as not installed).
	Installed []InstalledVersion `json:"installed,omitempty"`
}

// Tap represents a Homebrew tap repository
type Tap struct {
	Name         string `json:"name"`
	Remote       string `json:"remote"`
	IsOfficial   bool   `json:"is_official"`
	FormulaCount int    `json:"formula_count"`
	CaskCount    int    `json:"cask_count"`
	CommandCount int    `json:"command_count"`
	Installed    bool   `json:"installed"`
	IsAPI        bool   `json:"is_api"` // API-sourced vs git clone
	// 6.0.0 additions
	Trusted      bool     `json:"trusted,omitempty"`       // trust status from tap-info
	FormulaNames []string `json:"formula_names,omitempty"` // formulae in this tap
	CaskNames    []string `json:"cask_names,omitempty"`    // casks in this tap
}

// TrustType represents the type of trusted object
type TrustType string

const (
	TrustTypeTap     TrustType = "tap"
	TrustTypeFormula TrustType = "formula"
	TrustTypeCask    TrustType = "cask"
	TrustTypeCommand TrustType = "command"
)

// TrustEntry represents a single trusted item
type TrustEntry struct {
	Name string    `json:"name"` // tap, formula, or cask name
	Type TrustType `json:"type"` // tap | formula | cask | command
	Tap  string    `json:"tap"`  // parent tap
}

// Service represents a brew-managed service
type Service struct {
	Name     string        `json:"name"`
	Status   ServiceStatus `json:"status"`
	User     string        `json:"user"`
	File     string        `json:"file"`
	ExitCode int           `json:"exit_code"`
}

// ServiceInfo is one entry of `brew services info --all --json`, the detail
// the Services Info view shows. Brew reports some fields as null: a string
// that is null is "" here, a pointer or map that is null is nil. The schema is
// brew's own, in Library/Homebrew/services/formula_wrapper.rb (to_hash).
type ServiceInfo struct {
	Name         string         `json:"name"`
	ServiceName  string         `json:"service_name"`
	Running      bool           `json:"running"`
	Loaded       bool           `json:"loaded"`
	Schedulable  bool           `json:"schedulable"`
	PID          *int           `json:"pid"`
	ExitCode     *int           `json:"exit_code"`
	User         string         `json:"user"`
	Status       ServiceStatus  `json:"status"`
	File         string         `json:"file"`
	Registered   bool           `json:"registered"`
	LoadedFile   string         `json:"loaded_file"`
	Command      string         `json:"command"`
	WorkingDir   string         `json:"working_dir"`
	RootDir      string         `json:"root_dir"`
	LogPath      string         `json:"log_path"`
	ErrorLogPath string         `json:"error_log_path"`
	Interval     *int           `json:"interval"`
	Cron         map[string]any `json:"cron"`
}

// ServiceStatus represents the status of a service
type ServiceStatus string

const (
	ServiceStarted   ServiceStatus = "started"
	ServiceStopped   ServiceStatus = "stopped"
	ServiceError     ServiceStatus = "error"
	ServiceNone      ServiceStatus = "none"
	ServiceScheduled ServiceStatus = "scheduled"
)

// String returns the string representation of ServiceStatus
func (s ServiceStatus) String() string {
	return string(s)
}

// SearchResult represents a search match
type SearchResult struct {
	Name        string `json:"name"`
	IsFormula   bool   `json:"is_formula"`
	IsCask      bool   `json:"is_cask"`
	Installed   bool   `json:"installed"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// DoctorWarning represents a single brew doctor warning
type DoctorWarning struct {
	Title   string `json:"title"`
	Details string `json:"details"`
}

// BrewConfig holds system-level brew configuration
type BrewConfig struct {
	HomebrewVersion string `json:"homebrew_version"`
	Prefix          string `json:"prefix"`
	Cellar          string `json:"cellar"`
	Repository      string `json:"repository"`
	CoreTap         string `json:"core_tap"`
	OS              string `json:"os"`
}
