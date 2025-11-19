// Package v1beta1 contains the v1beta1 API definition for OPCT reports.
// This package provides stable, versioned API structures for consuming
// OPCT report data externally.
package v1beta1

import (
	"time"
)

// APIVersion represents the API version for this package
const APIVersion = "v1beta1"

// ReportData represents the top-level structure of an OPCT report.
// This is the main data structure that gets serialized to opct-report.json.
type ReportData struct {
	// APIVersion indicates the version of this API structure
	APIVersion string `json:"apiVersion"`

	// Kind indicates the type of this resource
	Kind string `json:"kind"`

	// Summary contains high-level summary information about the test execution
	Summary *ReportSummary `json:"summary"`

	// Provider contains the test results and cluster information for the provider under test
	Provider *ReportResult `json:"provider"`

	// Baseline contains the test results and cluster information for the baseline comparison (optional)
	Baseline *ReportResult `json:"baseline,omitempty"`

	// Checks contains the SLO/SLI validation results
	Checks *ReportChecks `json:"checks,omitempty"`

	// Setup contains configuration and metadata about the report generation
	Setup *ReportSetup `json:"setup,omitempty"`
}

// ReportSummary provides high-level information about the test execution
type ReportSummary struct {
	// Tests contains information about test archives and execution
	Tests *ReportSummaryTests `json:"tests"`

	// Alerts contains information about any alerts or issues detected
	Alerts *ReportSummaryAlerts `json:"alerts"`

	// Runtime contains timing and execution information
	Runtime *ReportSummaryRuntime `json:"runtime,omitempty"`

	// Headline provides a human-readable summary of the test results
	Headline string `json:"headline"`

	// Features indicates which optional features were available during testing
	Features ReportSummaryFeatures `json:"features,omitempty"`
}

// ReportSummaryFeatures indicates availability of optional data sources
type ReportSummaryFeatures struct {
	// HasCAMGI indicates if Cluster Availability Monitoring Gathered Information is available
	HasCAMGI bool `json:"hasCAMGI,omitempty"`

	// HasMetricsData indicates if metrics data was collected
	HasMetricsData bool `json:"hasMetricsData,omitempty"`

	// HasInstallConfig indicates if install configuration was available
	HasInstallConfig bool `json:"hasInstallConfig,omitempty"`
}

// ReportSummaryRuntime contains execution timing information
type ReportSummaryRuntime struct {
	// Timers contains detailed timing information for various operations
	Timers map[string]TimerInfo `json:"timers,omitempty"`

	// Plugins contains execution time for each plugin
	Plugins map[string]string `json:"plugins,omitempty"`

	// ExecutionTime is the total execution time as a string
	ExecutionTime string `json:"executionTime,omitempty"`
}

// TimerInfo represents timing information for a specific operation
type TimerInfo struct {
	// Seconds is the duration in seconds
	Seconds float64 `json:"seconds"`
}

// ReportSummaryTests contains information about test archives
type ReportSummaryTests struct {
	// Archive is the primary test archive filename
	Archive string `json:"archive"`

	// ArchiveDiff is the baseline archive filename for comparison (optional)
	ArchiveDiff string `json:"archiveDiff,omitempty"`
}

// ReportSummaryAlerts contains alert information for various components
type ReportSummaryAlerts struct {
	// PluginK8S is the alert level for Kubernetes conformance tests
	PluginK8S string `json:"pluginK8S,omitempty"`

	// PluginK8SMessage is the alert message for Kubernetes conformance tests
	PluginK8SMessage string `json:"pluginK8SMessage,omitempty"`

	// PluginOCP is the alert level for OpenShift conformance tests
	PluginOCP string `json:"pluginOCP,omitempty"`

	// PluginOCPMessage is the alert message for OpenShift conformance tests
	PluginOCPMessage string `json:"pluginOCPMessage,omitempty"`

	// SuiteErrors is the alert level for test suite errors
	SuiteErrors string `json:"suiteErrors,omitempty"`

	// SuiteErrorsMessage is the alert message for test suite errors
	SuiteErrorsMessage string `json:"suiteErrorsMessage,omitempty"`

	// WorkloadErrors is the alert level for workload errors
	WorkloadErrors string `json:"workloadErrors,omitempty"`

	// WorkloadErrorsMessage is the alert message for workload errors
	WorkloadErrorsMessage string `json:"workloadErrorsMessage,omitempty"`

	// Checks is the alert level for SLO/SLI checks
	Checks string `json:"checks,omitempty"`

	// ChecksMessage is the alert message for SLO/SLI checks
	ChecksMessage string `json:"checksMessage,omitempty"`
}

// ReportResult contains test results and cluster information
type ReportResult struct {
	// Version contains version information for OpenShift, Kubernetes, and OPCT
	Version *ReportVersion `json:"version"`

	// Infra contains infrastructure and platform information
	Infra *ReportInfra `json:"infra"`

	// ClusterOperators contains cluster operator status information
	ClusterOperators *ReportClusterOperators `json:"clusterOperators"`

	// ClusterHealth contains cluster health metrics
	ClusterHealth *ReportClusterHealth `json:"clusterHealth"`

	// Plugins contains results for each test plugin
	Plugins map[string]*ReportPlugin `json:"plugins"`

	// HasValidBaseline indicates if a valid baseline is available for comparison
	HasValidBaseline bool `json:"hasValidBaseline"`

	// MustGatherInfo contains must-gather collection information
	MustGatherInfo *MustGatherInfo `json:"mustGatherInfo,omitempty"`

	// ErrorCounters contains aggregated error counts
	ErrorCounters *ErrorCounter `json:"errorCounters,omitempty"`

	// Runtime contains runtime information and logs
	Runtime *ReportRuntime `json:"runtime,omitempty"`

	// Nodes contains information about cluster nodes
	Nodes []*NodeInfo `json:"nodes,omitempty"`

	// InstallInvoker indicates what tool was used to install the cluster
	InstallInvoker *string `json:"installInvoker,omitempty"`
}

// ReportVersion contains version information
type ReportVersion struct {
	// OpenShift contains OpenShift version information
	OpenShift *OpenShiftVersionInfo `json:"openshift"`

	// Kubernetes is the Kubernetes version string
	Kubernetes string `json:"kubernetes"`

	// OPCTServer is the OPCT server version
	OPCTServer string `json:"opctServer,omitempty"`

	// OPCTClient is the OPCT client version
	OPCTClient string `json:"opctClient,omitempty"`
}

// OpenShiftVersionInfo contains detailed OpenShift version information
type OpenShiftVersionInfo struct {
	// Desired is the target OpenShift version
	Desired string `json:"desired"`

	// Previous is the previous OpenShift version (for upgrades)
	Previous string `json:"previous"`

	// Channel is the update channel
	Channel string `json:"channel"`

	// ClusterID is the unique cluster identifier
	ClusterID string `json:"clusterID"`

	// OverallStatus is the overall cluster version status
	OverallStatus string `json:"overallStatus"`

	// ConditionAvailable indicates if the cluster version is available
	ConditionAvailable string `json:"conditionAvailable"`

	// ConditionFailing indicates if the cluster version is failing
	ConditionFailing string `json:"conditionFailing"`

	// ConditionProgressing indicates if the cluster version is progressing
	ConditionProgressing string `json:"conditionProgressing"`

	// ConditionProgressingMessage provides details about progression
	ConditionProgressingMessage string `json:"conditionProgressingMessage"`

	// ConditionUpdates indicates the status of available updates
	ConditionUpdates string `json:"conditionUpdates"`

	// ConditionImplicitlyEnabledCapabilities indicates implicitly enabled capabilities
	ConditionImplicitlyEnabledCapabilities string `json:"conditionImplicitlyEnabledCapabilities"`

	// ConditionReleaseAccepted indicates if the release was accepted
	ConditionReleaseAccepted string `json:"conditionReleaseAccepted"`

	// HistoryCompletionTime is when the version history was completed
	HistoryCompletionTime string `json:"historyCompletionTime"`
}

// ReportInfra contains infrastructure information
type ReportInfra struct {
	// Name is the infrastructure name
	Name string `json:"name"`

	// PlatformType is the platform type (e.g., AWS, Azure, GCP)
	PlatformType string `json:"platformType"`

	// PlatformName is the specific platform name for external platforms
	PlatformName string `json:"platformName"`

	// Topology describes the infrastructure topology
	Topology string `json:"topology,omitempty"`

	// ControlPlaneTopology describes the control plane topology
	ControlPlaneTopology string `json:"controlPlaneTopology,omitempty"`

	// APIServerURL is the external API server URL
	APIServerURL string `json:"apiServerURL,omitempty"`

	// APIServerInternalURL is the internal API server URL
	APIServerInternalURL string `json:"apiServerInternalURL,omitempty"`

	// NetworkType is the cluster network type
	NetworkType string `json:"networkType,omitempty"`
}

// ReportClusterOperators contains cluster operator status counts
type ReportClusterOperators struct {
	// CountAvailable is the number of available cluster operators
	CountAvailable uint64 `json:"countAvailable,omitempty"`

	// CountProgressing is the number of progressing cluster operators
	CountProgressing uint64 `json:"countProgressing,omitempty"`

	// CountDegraded is the number of degraded cluster operators
	CountDegraded uint64 `json:"countDegraded,omitempty"`
}

// ReportClusterHealth contains cluster health metrics
type ReportClusterHealth struct {
	// NodeHealthTotal is the total number of nodes
	NodeHealthTotal int `json:"nodeHealthTotal,omitempty"`

	// NodeHealthy is the number of healthy nodes
	NodeHealthy int `json:"nodeHealthy,omitempty"`

	// NodeHealthPerc is the percentage of healthy nodes
	NodeHealthPerc float64 `json:"nodeHealthPerc,omitempty"`

	// PodHealthTotal is the total number of pods
	PodHealthTotal int `json:"podHealthTotal,omitempty"`

	// PodHealthy is the number of healthy pods
	PodHealthy int `json:"podHealthy,omitempty"`

	// PodHealthPerc is the percentage of healthy pods
	PodHealthPerc float64 `json:"podHealthPerc,omitempty"`

	// PodHealthDetails contains details about unhealthy pods
	PodHealthDetails []PodHealthDetail `json:"podHealthDetails,omitempty"`
}

// PodHealthDetail contains information about an unhealthy pod
type PodHealthDetail struct {
	// Name is the pod name
	Name string `json:"name"`

	// Namespace is the pod namespace
	Namespace string `json:"namespace"`

	// Status is the pod status
	Status string `json:"status"`

	// Healthy indicates if the pod is healthy
	Healthy bool `json:"healthy"`

	// Message provides additional information
	Message string `json:"message,omitempty"`
}

// ReportPlugin represents test results for a specific plugin/test suite
type ReportPlugin struct {
	// ID is the unique identifier of the plugin
	ID string `json:"id"`

	// Title is the human-readable title of the plugin
	Title string `json:"title"`

	// Name is the name of the plugin
	Name string `json:"name"`

	// Definition contains metadata about the plugin definition
	Definition *PluginDefinition `json:"definition,omitempty"`

	// Stat contains execution statistics
	Stat *ReportPluginStat `json:"stat"`

	// ErrorCounters contains error counts by category
	ErrorCounters *ErrorCounter `json:"errorCounters,omitempty"`

	// Suite contains information about the test suite
	Suite *TestSuiteInfo `json:"suite"`

	// Tests contains individual test results
	Tests map[string]*TestItem `json:"tests,omitempty"`

	// Filter results - different filter stages in the pipeline
	FailedFilter1 []*ReportTestFailure `json:"failedTestsFilter1"`
	TagsFilter1   string               `json:"tagsFailuresFilter1"`

	FailedFilter2 []*ReportTestFailure `json:"failedTestsFilter2"`
	TagsFilter2   string               `json:"tagsFailuresFilter2"`

	FailedFilter3 []*ReportTestFailure `json:"failedTestsFilter3"`
	TagsFilter3   string               `json:"tagsFailuresFilter3"`

	FailedFilter4 []*ReportTestFailure `json:"failedTestsFilter4"`
	TagsFilter4   string               `json:"tagsFailuresFilter4"`

	FailedFilter5 []*ReportTestFailure `json:"failedTestsFilter5"`
	TagsFilter5   string               `json:"tagsFailuresFilter5"`

	FailedFilter6 []*ReportTestFailure `json:"failedTestsFilter6"`
	TagsFilter6   string               `json:"tagsFailuresFilter6"`

	// Final results after all filters
	FailedFiltered []*ReportTestFailure `json:"failedFiltered"`
	TagsFiltered   string               `json:"tagsFailuresFiltered"`
}

// PluginDefinition contains metadata about a plugin
type PluginDefinition struct {
	// PluginImage is the container image used by the plugin
	PluginImage string `json:"pluginImage"`

	// SonobuoyImage is the Sonobuoy image used
	SonobuoyImage string `json:"sonobuoyImage"`

	// Name is the plugin name
	Name string `json:"name"`
}

// ReportPluginStat contains execution statistics for a plugin
type ReportPluginStat struct {
	// Completed indicates when execution completed
	Completed string `json:"execution"`

	// Result is the overall result (passed/failed/etc.)
	Result string `json:"result"`

	// Status is the execution status
	Status string `json:"status"`

	// Total is the total number of tests
	Total int64 `json:"total"`

	// Passed is the number of passed tests
	Passed int64 `json:"passed"`

	// Failed is the number of failed tests
	Failed int64 `json:"failed"`

	// Timeout is the number of timed out tests
	Timeout int64 `json:"timeout"`

	// Skipped is the number of skipped tests
	Skipped int64 `json:"skipped"`

	// Filter statistics for each filter stage
	FilterSuite     int64 `json:"filter1Suite"`
	Filter1Excluded int64 `json:"filter1Excluded"`

	FilterBaseline  int64 `json:"filter2Baseline"`
	Filter2Excluded int64 `json:"filter2Excluded"`

	FilterFailedPrio int64 `json:"filter3FailedPriority"`
	Filter3Excluded  int64 `json:"filter3Excluded"`

	FilterFailedAPI int64 `json:"filter4FailedAPI"`
	Filter4Excluded int64 `json:"filter4Excluded"`

	Filter5Failures int64 `json:"filter5Failures"`
	Filter5Excluded int64 `json:"filter5Excluded"`

	Filter6Failures int64 `json:"filter6Failures"`
	Filter6Excluded int64 `json:"filter6Excluded"`

	FilterFailures int64 `json:"filterFailures"`
}

// ReportTestFailure contains information about a failed test
type ReportTestFailure struct {
	// ID is the test identifier
	ID string `json:"id"`

	// Name is the test name
	Name string `json:"name"`

	// Documentation is a link to test documentation
	Documentation string `json:"documentation"`

	// FlakePerc is the flake percentage for this test
	FlakePerc float64 `json:"flakePerc"`

	// FlakeCount is the number of times this test has been flaky
	FlakeCount int64 `json:"flakeCount"`

	// ErrorsCount is the total number of errors for this test
	ErrorsCount int64 `json:"errorsTotal"`
}

// TestSuiteInfo contains information about a test suite
type TestSuiteInfo struct {
	// Name is the suite name
	Name string `json:"name"`

	// Description is the suite description
	Description string `json:"description,omitempty"`

	// Version is the suite version
	Version string `json:"version,omitempty"`
}

// TestItem contains information about an individual test
type TestItem struct {
	// ID is the test identifier
	ID string `json:"id"`

	// Name is the test name
	Name string `json:"name"`

	// Status is the test status (passed/failed/skipped)
	Status string `json:"status"`

	// Documentation is a link to test documentation
	Documentation string `json:"documentation,omitempty"`

	// Flake contains flake information
	Flake *FlakeInfo `json:"flake,omitempty"`

	// ErrorCounters contains error counts for this test
	ErrorCounters map[string]uint64 `json:"errorCounters,omitempty"`
}

// FlakeInfo contains flake detection information
type FlakeInfo struct {
	// CurrentFlakes is the current number of flakes
	CurrentFlakes int64 `json:"currentFlakes"`

	// CurrentFlakePerc is the current flake percentage
	CurrentFlakePerc float64 `json:"currentFlakePerc"`
}

// ErrorCounter contains error counts by category
type ErrorCounter struct {
	// Categories contains error counts by category name
	Categories map[string]uint64 `json:"categories,omitempty"`
}

// NodeInfo contains information about a cluster node
type NodeInfo struct {
	// Name is the node name
	Name string `json:"name"`

	// Roles are the node roles (master, worker, etc.)
	Roles []string `json:"roles"`

	// Status is the node status
	Status string `json:"status"`

	// Version is the node version
	Version string `json:"version,omitempty"`

	// Architecture is the node architecture
	Architecture string `json:"architecture,omitempty"`

	// KernelVersion is the kernel version
	KernelVersion string `json:"kernelVersion,omitempty"`

	// OSImage is the operating system image
	OSImage string `json:"osImage,omitempty"`

	// ContainerRuntimeVersion is the container runtime version
	ContainerRuntimeVersion string `json:"containerRuntimeVersion,omitempty"`
}

// MustGatherInfo contains information about must-gather collection
type MustGatherInfo struct {
	// CollectionTime is when the must-gather was collected
	CollectionTime *time.Time `json:"collectionTime,omitempty"`

	// Status is the collection status
	Status string `json:"status,omitempty"`

	// Size is the size of the collected data
	Size int64 `json:"size,omitempty"`

	// Path is the path to the collected data
	Path string `json:"path,omitempty"`
}

// ReportChecks contains SLO/SLI validation results
type ReportChecks struct {
	// BaseURL is the base URL for check documentation
	BaseURL string `json:"baseURL"`

	// EmptyValue is the value used for empty/null checks
	EmptyValue string `json:"emptyValue"`

	// Fail contains failed checks
	Fail []*SLOOutput `json:"failures"`

	// Pass contains passed checks
	Pass []*SLOOutput `json:"successes"`

	// Warn contains warning checks
	Warn []*SLOOutput `json:"warnings"`

	// Skip contains skipped checks
	Skip []*SLOOutput `json:"skips"`
}

// SLOOutput represents the result of an SLO/SLI check
type SLOOutput struct {
	// ID is the check identifier
	ID string `json:"id"`

	// SLO is the Service Level Objective description
	SLO string `json:"slo"`

	// SLOResult is the result (pass/fail/warn/skip)
	SLOResult string `json:"sloResult"`

	// SLITarget is the Service Level Indicator target value
	SLITarget string `json:"sliTarget"`

	// SLICurrent is the current Service Level Indicator value
	SLICurrent string `json:"sliCurrent"`

	// Message provides additional information about the result
	Message string `json:"message"`

	// Documentation is a link to documentation for this check
	Documentation string `json:"documentation"`
}

// ReportSetup contains configuration and metadata
type ReportSetup struct {
	// Frontend contains frontend-specific configuration
	Frontend *ReportSetupFrontend `json:"frontend,omitempty"`

	// API contains API-specific metadata
	API *ReportSetupAPI `json:"api,omitempty"`
}

// ReportSetupFrontend contains frontend configuration
type ReportSetupFrontend struct {
	// EmbedData indicates if data should be embedded in the frontend
	EmbedData bool `json:"embedData"`
}

// ReportSetupAPI contains API metadata
type ReportSetupAPI struct {
	// SummaryName is the name of the summary file
	SummaryName string `json:"dataPath,omitempty"`

	// SummaryArchive is the path to the summary archive
	SummaryArchive string `json:"summaryArchive,omitempty"`

	// UUID is the unique identifier for this test run
	UUID string `json:"uuid,omitempty"`

	// ExecutionDate is when the test was executed
	ExecutionDate string `json:"executionDate,omitempty"`

	// OpenShiftVersion is the OpenShift version tested
	OpenShiftVersion string `json:"openshiftVersion,omitempty"`

	// OpenShiftRelease is the OpenShift release (major.minor)
	OpenShiftRelease string `json:"openshiftRelease,omitempty"`

	// PlatformType is the platform type
	PlatformType string `json:"platformType,omitempty"`

	// ProviderName is the provider name
	ProviderName string `json:"providerName,omitempty"`

	// InfraTopology is the infrastructure topology
	InfraTopology string `json:"infraTopology,omitempty"`

	// Workflow is the test workflow used
	Workflow string `json:"workflow,omitempty"`
}

// ReportRuntime contains runtime information and logs
type ReportRuntime struct {
	// ServerLogs contains server log entries
	ServerLogs []*RuntimeInfoItem `json:"serverLogs,omitempty"`

	// ServerConfig contains server configuration entries
	ServerConfig []*RuntimeInfoItem `json:"serverConfig,omitempty"`

	// OpctConfig contains OPCT configuration entries
	OpctConfig []*RuntimeInfoItem `json:"opctConfig,omitempty"`
}

// RuntimeInfoItem represents a single runtime information entry
type RuntimeInfoItem struct {
	// Name is the entry name
	Name string `json:"name"`

	// Value is the entry value
	Value string `json:"value"`

	// Time is the timestamp for this entry
	Time string `json:"time,omitempty"`

	// Total is the total time (for timer entries)
	Total string `json:"total,omitempty"`

	// Delta is the delta time (for timer entries)
	Delta string `json:"delta,omitempty"`
}
