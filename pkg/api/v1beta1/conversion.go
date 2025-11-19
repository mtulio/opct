// Package v1beta1 provides conversion functions between internal and API types.
package v1beta1

import (
	"strings"

	"github.com/redhat-openshift-ecosystem/opct/internal/opct/archive"
	"github.com/redhat-openshift-ecosystem/opct/internal/opct/plugin"
	"github.com/redhat-openshift-ecosystem/opct/internal/opct/summary"
	"github.com/redhat-openshift-ecosystem/opct/internal/openshift/ci/sippy"
	"github.com/redhat-openshift-ecosystem/opct/internal/openshift/mustgather"
	"github.com/redhat-openshift-ecosystem/opct/internal/report"
	"github.com/vmware-tanzu/sonobuoy/pkg/discovery"
)

// ConvertFromInternal converts internal report data to v1beta1 API format
func ConvertFromInternal(internal *report.ReportData) *ReportData {
	if internal == nil {
		return nil
	}

	apiData := &ReportData{
		APIVersion: APIVersion,
		Kind:       "OPCTReport",
		Summary:    convertReportSummary(internal.Summary),
		Provider:   convertReportResult(internal.Provider),
		Baseline:   convertReportResult(internal.Baseline),
		Checks:     convertReportChecks(internal.Checks),
		Setup:      convertReportSetup(internal.Setup),
	}

	return apiData
}

func convertReportSummary(internal *report.ReportSummary) *ReportSummary {
	if internal == nil {
		return nil
	}

	return &ReportSummary{
		Tests:    convertReportSummaryTests(internal.Tests),
		Alerts:   convertReportSummaryAlerts(internal.Alerts),
		Runtime:  convertReportSummaryRuntime(internal.Runtime),
		Headline: internal.Headline,
		Features: ReportSummaryFeatures{
			HasCAMGI:         internal.Features.HasCAMGI,
			HasMetricsData:   internal.Features.HasMetricsData,
			HasInstallConfig: internal.Features.HasInstallConfig,
		},
	}
}

func convertReportSummaryTests(internal *report.ReportSummaryTests) *ReportSummaryTests {
	if internal == nil {
		return nil
	}

	return &ReportSummaryTests{
		Archive:     internal.Archive,
		ArchiveDiff: internal.ArchiveDiff,
	}
}

func convertReportSummaryAlerts(internal *report.ReportSummaryAlerts) *ReportSummaryAlerts {
	if internal == nil {
		return nil
	}

	return &ReportSummaryAlerts{
		PluginK8S:             internal.PluginK8S,
		PluginK8SMessage:      internal.PluginK8SMessage,
		PluginOCP:             internal.PluginOCP,
		PluginOCPMessage:      internal.PluginOCPMessage,
		SuiteErrors:           internal.SuiteErrors,
		SuiteErrorsMessage:    internal.SuiteErrorsMessage,
		WorkloadErrors:        internal.WorkloadErrors,
		WorkloadErrorsMessage: internal.WorkloadErrorsMessage,
		Checks:                internal.Checks,
		ChecksMessage:         internal.ChecksMessage,
	}
}

func convertReportSummaryRuntime(internal *report.ReportSummaryRuntime) *ReportSummaryRuntime {
	if internal == nil {
		return nil
	}

	// Convert timers from metrics.Timers to map[string]TimerInfo
	timers := make(map[string]TimerInfo)
	if internal.Timers != nil && internal.Timers.Timers != nil {
		for name, timer := range internal.Timers.Timers {
			timers[name] = TimerInfo{
				Seconds: timer.Total,
			}
		}
	}

	return &ReportSummaryRuntime{
		Timers:        timers,
		Plugins:       internal.Plugins,
		ExecutionTime: internal.ExecutionTime,
	}
}

func convertReportResult(internal *report.ReportResult) *ReportResult {
	if internal == nil {
		return nil
	}

	return &ReportResult{
		Version:          convertReportVersion(internal.Version),
		Infra:            convertReportInfra(internal.Infra),
		ClusterOperators: convertReportClusterOperators(internal.ClusterOperators),
		ClusterHealth:    convertReportClusterHealth(internal.ClusterHealth),
		Plugins:          convertReportPlugins(internal.Plugins),
		HasValidBaseline: internal.HasValidBaseline,
		MustGatherInfo:   convertMustGatherInfo(internal.MustGatherInfo),
		ErrorCounters:    convertErrorCounter(internal.ErrorCounters),
		Runtime:          convertReportRuntime(internal.Runtime),
		Nodes:            convertNodes(internal.Nodes),
		InstallInvoker:   internal.InstallInvoker,
	}
}

func convertReportVersion(internal *report.ReportVersion) *ReportVersion {
	if internal == nil {
		return nil
	}

	return &ReportVersion{
		OpenShift:  convertOpenShiftVersionInfo(internal.OpenShift),
		Kubernetes: internal.Kubernetes,
		OPCTServer: internal.OPCTServer,
		OPCTClient: internal.OPCTClient,
	}
}

func convertOpenShiftVersionInfo(internal *summary.SummaryClusterVersionOutput) *OpenShiftVersionInfo {
	if internal == nil {
		return nil
	}

	return &OpenShiftVersionInfo{
		Desired:                                internal.Desired,
		Previous:                               internal.Previous,
		Channel:                                internal.Channel,
		ClusterID:                              internal.ClusterID,
		OverallStatus:                          internal.OverallStatus,
		ConditionAvailable:                     internal.CondAvailable,
		ConditionFailing:                       internal.CondFailing,
		ConditionProgressing:                   internal.CondProgressing,
		ConditionProgressingMessage:            internal.CondProgressingMessage,
		ConditionUpdates:                       internal.CondRetrievedUpdates,
		ConditionImplicitlyEnabledCapabilities: internal.CondImplicitlyEnabledCapabilities,
		ConditionReleaseAccepted:               internal.CondReleaseAccepted,
		HistoryCompletionTime:                  internal.HistoryCompletionTime,
	}
}

func convertReportInfra(internal *report.ReportInfra) *ReportInfra {
	if internal == nil {
		return nil
	}

	return &ReportInfra{
		Name:                 internal.Name,
		PlatformType:         internal.PlatformType,
		PlatformName:         internal.PlatformName,
		Topology:             internal.Topology,
		ControlPlaneTopology: internal.ControlPlaneTopology,
		APIServerURL:         internal.APIServerURL,
		APIServerInternalURL: internal.APIServerInternalURL,
		NetworkType:          internal.NetworkType,
	}
}

func convertReportClusterOperators(internal *report.ReportClusterOperators) *ReportClusterOperators {
	if internal == nil {
		return nil
	}

	return &ReportClusterOperators{
		CountAvailable:   internal.CountAvailable,
		CountProgressing: internal.CountProgressing,
		CountDegraded:    internal.CountDegraded,
	}
}

func convertReportClusterHealth(internal *report.ReportClusterHealth) *ReportClusterHealth {
	if internal == nil {
		return nil
	}

	// Convert pod health details
	var podHealthDetails []PodHealthDetail
	for _, detail := range internal.PodHealthDetails {
		podHealthDetails = append(podHealthDetails, convertPodHealthDetail(detail))
	}

	return &ReportClusterHealth{
		NodeHealthTotal:  internal.NodeHealthTotal,
		NodeHealthy:      internal.NodeHealthy,
		NodeHealthPerc:   internal.NodeHealthPerc,
		PodHealthTotal:   internal.PodHealthTotal,
		PodHealthy:       internal.PodHealthy,
		PodHealthPerc:    internal.PodHealthPerc,
		PodHealthDetails: podHealthDetails,
	}
}

func convertPodHealthDetail(internal discovery.HealthInfoDetails) PodHealthDetail {
	return PodHealthDetail{
		Name:      internal.Name,
		Namespace: internal.Namespace,
		Status:    "Unknown", // Status field not directly available
		Healthy:   internal.Healthy,
		Message:   internal.Message,
	}
}

func convertReportPlugins(internal map[string]*report.ReportPlugin) map[string]*ReportPlugin {
	if internal == nil {
		return nil
	}

	plugins := make(map[string]*ReportPlugin)
	for name, plugin := range internal {
		plugins[name] = convertReportPlugin(plugin)
	}
	return plugins
}

func convertReportPlugin(internal *report.ReportPlugin) *ReportPlugin {
	if internal == nil {
		return nil
	}

	return &ReportPlugin{
		ID:             internal.ID,
		Title:          internal.Title,
		Name:           internal.Name,
		Definition:     convertPluginDefinition(internal.Definition),
		Stat:           convertReportPluginStat(internal.Stat),
		ErrorCounters:  convertErrorCounter(internal.ErrorCounters),
		Suite:          convertTestSuiteInfo(internal.Suite),
		Tests:          convertTestItems(internal.Tests),
		FailedFilter1:  convertReportTestFailures(internal.FailedFilter1),
		TagsFilter1:    internal.TagsFilter1,
		FailedFilter2:  convertReportTestFailures(internal.FailedFilter2),
		TagsFilter2:    internal.TagsFilter2,
		FailedFilter3:  convertReportTestFailures(internal.FailedFilter3),
		TagsFilter3:    internal.TagsFilter3,
		FailedFilter4:  convertReportTestFailures(internal.FailedFilter4),
		TagsFilter4:    internal.TagsFilter4,
		FailedFilter5:  convertReportTestFailures(internal.FailedFilter5),
		TagsFilter5:    internal.TagsFilter5,
		FailedFilter6:  convertReportTestFailures(internal.FailedFilter6),
		TagsFilter6:    internal.TagsFilter6,
		FailedFiltered: convertReportTestFailures(internal.FailedFiltered),
		TagsFiltered:   internal.TagsFiltered,
	}
}

func convertPluginDefinition(internal *plugin.PluginDefinition) *PluginDefinition {
	if internal == nil {
		return nil
	}

	return &PluginDefinition{
		PluginImage:   internal.PluginImage,
		SonobuoyImage: internal.SonobuoyImage,
		Name:          internal.Name,
	}
}

func convertReportPluginStat(internal *report.ReportPluginStat) *ReportPluginStat {
	if internal == nil {
		return nil
	}

	return &ReportPluginStat{
		Completed:        internal.Completed,
		Result:           internal.Result,
		Status:           internal.Status,
		Total:            internal.Total,
		Passed:           internal.Passed,
		Failed:           internal.Failed,
		Timeout:          internal.Timeout,
		Skipped:          internal.Skipped,
		FilterSuite:      internal.FilterSuite,
		Filter1Excluded:  internal.Filter1Excluded,
		FilterBaseline:   internal.FilterBaseline,
		Filter2Excluded:  internal.Filter2Excluded,
		FilterFailedPrio: internal.FilterFailedPrio,
		Filter3Excluded:  internal.Filter3Excluded,
		FilterFailedAPI:  internal.FilterFailedAPI,
		Filter4Excluded:  internal.Filter4Excluded,
		Filter5Failures:  internal.Filter5Failures,
		Filter5Excluded:  internal.Filter5Excluded,
		Filter6Failures:  internal.Filter6Failures,
		Filter6Excluded:  internal.Filter6Excluded,
		FilterFailures:   internal.FilterFailures,
	}
}

func convertTestSuiteInfo(internal *summary.OpenshiftTestsSuite) *TestSuiteInfo {
	if internal == nil {
		return nil
	}

	return &TestSuiteInfo{
		Name:        internal.Name,
		Description: "", // Not available in internal structure
		Version:     "", // Not available in internal structure
	}
}

func convertTestItems(internal map[string]*plugin.TestItem) map[string]*TestItem {
	if internal == nil {
		return nil
	}

	tests := make(map[string]*TestItem)
	for name, test := range internal {
		tests[name] = convertTestItem(test)
	}
	return tests
}

func convertTestItem(internal *plugin.TestItem) *TestItem {
	if internal == nil {
		return nil
	}

	return &TestItem{
		ID:            internal.ID,
		Name:          internal.Name,
		Status:        internal.Status,
		Documentation: internal.Documentation,
		Flake:         convertFlakeInfo(internal.Flake),
		ErrorCounters: convertErrorCountersMap(internal.ErrorCounters),
	}
}

func convertFlakeInfo(internal *sippy.SippyTestsResponse) *FlakeInfo {
	if internal == nil {
		return nil
	}

	return &FlakeInfo{
		CurrentFlakes:    internal.CurrentFlakes,
		CurrentFlakePerc: internal.CurrentFlakePerc,
	}
}

func convertReportTestFailures(internal []*report.ReportTestFailure) []*ReportTestFailure {
	if internal == nil {
		return nil
	}

	failures := make([]*ReportTestFailure, len(internal))
	for i, failure := range internal {
		failures[i] = convertReportTestFailure(failure)
	}
	return failures
}

func convertReportTestFailure(internal *report.ReportTestFailure) *ReportTestFailure {
	if internal == nil {
		return nil
	}

	return &ReportTestFailure{
		ID:            internal.ID,
		Name:          internal.Name,
		Documentation: internal.Documentation,
		FlakePerc:     internal.FlakePerc,
		FlakeCount:    internal.FlakeCount,
		ErrorsCount:   internal.ErrorsCount,
	}
}

func convertErrorCounter(internal *archive.ErrorCounter) *ErrorCounter {
	if internal == nil {
		return nil
	}

	// Convert the internal error counter structure
	// archive.ErrorCounter is a map[string]uint64 type
	categories := make(map[string]uint64)
	if *internal != nil {
		for key, value := range *internal {
			categories[key] = uint64(value)
		}
	}

	return &ErrorCounter{
		Categories: categories,
	}
}

func convertErrorCountersMap(internal archive.ErrorCounter) map[string]uint64 {
	if internal == nil {
		return nil
	}

	result := make(map[string]uint64)
	for key, value := range internal {
		result[key] = uint64(value)
	}
	return result
}

func convertNodes(internal []*summary.Node) []*NodeInfo {
	if internal == nil {
		return nil
	}

	nodes := make([]*NodeInfo, len(internal))
	for i, node := range internal {
		nodes[i] = convertNodeInfo(node)
	}
	return nodes
}

func convertNodeInfo(internal *summary.Node) *NodeInfo {
	if internal == nil {
		return nil
	}

	// Convert node roles from string to slice
	var roles []string
	if internal.NodeRoles != "" {
		roles = strings.Fields(internal.NodeRoles)
	}

	return &NodeInfo{
		Name:                    internal.Hostname,
		Roles:                   roles,
		Status:                  "Ready", // Default status as not available in internal structure
		Version:                 "",      // Not available in internal structure
		Architecture:            internal.Architecture,
		KernelVersion:           "", // Not available in internal structure
		OSImage:                 internal.OperatingSystem,
		ContainerRuntimeVersion: "", // Not available in internal structure
	}
}

func convertMustGatherInfo(internal *mustgather.MustGather) *MustGatherInfo {
	if internal == nil {
		return nil
	}

	// Convert the must-gather info - this may need adjustment based on actual structure
	return &MustGatherInfo{
		// Add conversion logic here based on the actual structure of mustgather.MustGather
		Status: "unknown", // placeholder
	}
}

func convertReportChecks(internal *report.ReportChecks) *ReportChecks {
	if internal == nil {
		return nil
	}

	return &ReportChecks{
		BaseURL:    internal.BaseURL,
		EmptyValue: internal.EmptyValue,
		Fail:       convertSLOOutputs(internal.Fail),
		Pass:       convertSLOOutputs(internal.Pass),
		Warn:       convertSLOOutputs(internal.Warn),
		Skip:       convertSLOOutputs(internal.Skip),
	}
}

func convertSLOOutputs(internal []*report.SLOOutput) []*SLOOutput {
	if internal == nil {
		return nil
	}

	outputs := make([]*SLOOutput, len(internal))
	for i, output := range internal {
		outputs[i] = convertSLOOutput(output)
	}
	return outputs
}

func convertSLOOutput(internal *report.SLOOutput) *SLOOutput {
	if internal == nil {
		return nil
	}

	return &SLOOutput{
		ID:            internal.ID,
		SLO:           internal.SLO,
		SLOResult:     internal.SLOResult,
		SLITarget:     internal.SLITarget,
		SLICurrent:    internal.SLIActual,
		Message:       internal.Message,
		Documentation: internal.Documentation,
	}
}

func convertReportSetup(internal *report.ReportSetup) *ReportSetup {
	if internal == nil {
		return nil
	}

	return &ReportSetup{
		Frontend: convertReportSetupFrontend(internal.Frontend),
		API:      convertReportSetupAPI(internal.API),
	}
}

func convertReportSetupFrontend(internal *report.ReportSetupFrontend) *ReportSetupFrontend {
	if internal == nil {
		return nil
	}

	return &ReportSetupFrontend{
		EmbedData: internal.EmbedData,
	}
}

func convertReportSetupAPI(internal *report.ReportSetupAPI) *ReportSetupAPI {
	if internal == nil {
		return nil
	}

	return &ReportSetupAPI{
		SummaryName:      internal.SummaryName,
		SummaryArchive:   internal.SummaryArchive,
		UUID:             internal.UUID,
		ExecutionDate:    internal.ExecutionDate,
		OpenShiftVersion: internal.OpenShiftVersion,
		OpenShiftRelease: internal.OpenShiftRelease,
		PlatformType:     internal.PlatformType,
		ProviderName:     internal.ProviderName,
		InfraTopology:    internal.InfraTopology,
		Workflow:         internal.Workflow,
	}
}

func convertReportRuntime(internal *report.ReportRuntime) *ReportRuntime {
	if internal == nil {
		return nil
	}

	return &ReportRuntime{
		ServerLogs:   convertRuntimeInfoItems(internal.ServerLogs),
		ServerConfig: convertRuntimeInfoItems(internal.ServerConfig),
		OpctConfig:   convertRuntimeInfoItems(internal.OpctConfig),
	}
}

func convertRuntimeInfoItems(internal []*archive.RuntimeInfoItem) []*RuntimeInfoItem {
	if internal == nil {
		return nil
	}

	items := make([]*RuntimeInfoItem, len(internal))
	for i, item := range internal {
		items[i] = convertRuntimeInfoItem(item)
	}
	return items
}

func convertRuntimeInfoItem(internal *archive.RuntimeInfoItem) *RuntimeInfoItem {
	if internal == nil {
		return nil
	}

	return &RuntimeInfoItem{
		Name:  internal.Name,
		Value: internal.Value,
		Time:  internal.Time,
		Total: internal.Total,
		Delta: internal.Delta,
	}
}
