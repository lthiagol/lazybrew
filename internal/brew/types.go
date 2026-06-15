package brew

import (
	"time"
)

// Formula represents an installed Homebrew formula
type Formula struct {
	Name           string    `json:"name"`
	FullName       string    `json:"full_name"` // e.g., "homebrew/core/neovim"
	Tap            string    `json:"tap"`
	Version        string    `json:"version"`
	Description    string    `json:"desc"`
	Homepage       string    `json:"homepage"`
	License        string    `json:"license"`
	Pinned         bool      `json:"pinned"`
	Outdated       bool      `json:"outdated"`
	NewVersion     string    `json:"new_version"` // populated when outdated
	InstalledOn    time.Time `json:"installed_on"`
	Dependencies   []string  `json:"dependencies"`
	BuildDeps      []string  `json:"build_dependencies"`
	Caveats        string    `json:"caveats"`
	KegOnly        bool      `json:"keg_only"`
	Bottled        bool      `json:"bottled"`
	InstalledOnReq bool      `json:"installed_on_request"` // installed on request vs as dep
	InstalledAsDep bool      `json:"installed_as_dependency"`
	InstallPath    string    `json:"install_path"`
	Size           int64     `json:"size"` // bytes
	// 6.0.0 additions
	Aliases             []string `json:"aliases,omitempty"`
	Binaries            []string `json:"binaries,omitempty"`             // executables installed by this formula
	InstalledDependents []string `json:"installed_dependents,omitempty"` // reverse deps (installed only)
	ListVersions        []string `json:"list_versions,omitempty"`        // all installed version history
	Revision            string   `json:"revision,omitempty"`             // formula revision
	Shadowed            bool     `json:"shadowed,omitempty"`             // PATH shadowing warning
}

// Cask represents an installed Homebrew cask
type Cask struct {
	Name        string   `json:"name"`
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
	Pinned        bool     `json:"pinned,omitempty"`         // cask pinning
	Sha256        string   `json:"sha256,omitempty"`         // checksum
	URL           string   `json:"url,omitempty"`            // download URL
	DependsOn     []string `json:"depends_on,omitempty"`     // cask dependencies
	ConflictsWith []string `json:"conflicts_with,omitempty"` // conflicting casks
}

// Tap represents a Homebrew tap repository
type Tap struct {
	Name         string `json:"name"`
	Remote       string `json:"remote"`
	IsOfficial   bool   `json:"is_official"`
	FormulaCount int    `json:"formula_count"`
	CaskCount    int    `json:"cask_count"`
	CommandCount int    `json:"command_count"`
	LastCommit   string `json:"last_commit"`
	Installed    bool   `json:"installed"`
	IsAPI        bool   `json:"is_api"` // API-sourced vs git clone
	// 6.0.0 additions
	Trusted      bool     `json:"trusted,omitempty"`       // trust status from tap-info
	FormulaNames []string `json:"formula_names,omitempty"` // formulae in this tap
	CaskNames    []string `json:"cask_names,omitempty"`    // casks in this tap
}

// TrustStatus represents a tap/formula/cask trust state
type TrustStatus int

const (
	TrustUnknown   TrustStatus = iota
	TrustOfficial              // homebrew/* — always trusted
	TrustTrusted               // explicitly trusted via `brew trust`
	TrustUntrusted             // default for third-party
)

// String returns the string representation of TrustStatus
func (t TrustStatus) String() string {
	switch t {
	case TrustOfficial:
		return "official"
	case TrustTrusted:
		return "trusted"
	case TrustUntrusted:
		return "untrusted"
	default:
		return "unknown"
	}
}

// ParseTrustStatus converts a string to a TrustStatus
func ParseTrustStatus(s string) TrustStatus {
	switch s {
	case "official":
		return TrustOfficial
	case "trusted":
		return TrustTrusted
	case "untrusted":
		return TrustUntrusted
	default:
		return TrustUnknown
	}
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

// ServiceStatus represents the status of a service
type ServiceStatus string

const (
	ServiceStarted ServiceStatus = "started"
	ServiceStopped ServiceStatus = "stopped"
	ServiceError   ServiceStatus = "error"
	ServiceNone    ServiceStatus = "none"
)

// String returns the string representation of ServiceStatus
func (s ServiceStatus) String() string {
	return string(s)
}

// ParseServiceStatus converts a string to a ServiceStatus
func ParseServiceStatus(s string) ServiceStatus {
	switch s {
	case "started":
		return ServiceStarted
	case "stopped":
		return ServiceStopped
	case "error":
		return ServiceError
	case "none":
		return ServiceNone
	default:
		return ServiceNone
	}
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
	Arch            string `json:"arch"`
}
