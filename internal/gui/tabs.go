package gui

// The tab indices of each panel, in the order panelTabs lists them. Code names
// a tab by these, never by a bare number, so a reordered bar changes one list.
// The Info tab of a panel is first, so its default tab is its details.

// The tabs of the Status panel: the dashboard, the config, the doctor and the
// missing dependencies.
const (
	statusDashboardTab = iota
	statusConfigTab
	statusDoctorTab
	statusMissingTab
)

// The tabs of the Formulae panel.
const (
	formulaInfoTab = iota
	formulaDepsTab
	formulaUsesTab
	formulaCaveatsTab
	formulaFilesTab
)

// The tabs of the Casks panel.
const (
	caskInfoTab = iota
	caskDepsTab
	caskCaveatsTab
)

// The tabs of the Outdated panel: the selected item's details, and its versions.
const (
	outdatedInfoTab = iota
	outdatedVersionsTab
)

// The tabs of the Taps panel.
const (
	tapInfoTab = iota
	tapTrustTab
	tapFormulaeTab
)

// The tabs of the Services panel: the selected service's Info, and its log.
const (
	serviceInfoTab = iota
	servicesLogsTab
)
