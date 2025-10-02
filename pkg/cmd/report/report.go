package report

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/redhat-openshift-ecosystem/opct/internal/opct/metrics"
	"github.com/redhat-openshift-ecosystem/opct/internal/opct/summary"
	"github.com/redhat-openshift-ecosystem/opct/internal/report"
	log "github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/sonobuoy/pkg/errlog"
)

type Input struct {
	archive         string
	archiveBase     string
	saveTo          string
	serverAddress   string
	serverSkip      bool
	embedData       bool
	saveOnly        bool
	verbose         bool
	json            bool
	skipBaselineAPI bool
	force           bool

	// insights flag to extract insights from the report. It requires additional setup.
	insights bool
}

func NewCmdReport() *cobra.Command {
	data := Input{}
	cmd := &cobra.Command{
		Use:   "report archive.tar.gz",
		Short: "Create a report from results.",
		Run: func(cmd *cobra.Command, args []string) {
			data.archive = args[0]
			checkFlags(&data)
			if err := processResult(&data); err != nil {
				errlog.LogError(errors.Wrapf(err, "could not process archive: %v", args[0]))
				os.Exit(1)
			}
		},
		Args: cobra.ExactArgs(1),
	}

	// TODO: Basline/Diff from CLI must be removed v0.6+ when the
	// report API is totally validated, introduced in v0.5.
	// report API is a serverless service storing CI results in S3, serving
	// summarized information through HTTP endpoint (CloudFront), it is consumed
	// in the filter pipeline while processing the report, preventing any additional
	// step from user to download a specific archive.
	cmd.Flags().StringVarP(
		&data.archiveBase, "baseline", "b", "",
		"[DEPRECATED] Baseline result archive file. Example: -b file.tar.gz",
	)
	cmd.Flags().StringVarP(
		&data.archiveBase, "diff", "d", "",
		"[DEPRECATED] Diff results from a baseline archive file. Example: --diff file.tar.gz",
	)

	cmd.Flags().StringVarP(
		&data.saveTo, "save-to", "s", "",
		"Extract and Save Results to disk. Default: /tmp/opct-tmp-results-<archive-name>. Example: -s ./results",
	)
	cmd.Flags().StringVar(
		&data.serverAddress, "server-address", "0.0.0.0:9090",
		"HTTP server address to serve files when --save-to is used. Example: --server-address 0.0.0.0:9090",
	)
	cmd.Flags().BoolVar(
		&data.serverSkip, "skip-server", false,
		"HTTP server address to serve files when --save-to is used. Example: --server-address 0.0.0.0:9090",
	)
	cmd.Flags().BoolVar(
		&data.embedData, "embed-data", false,
		"Force to embed the data into HTML report, allwoing the use of file protocol/CORS in the browser.",
	)
	cmd.Flags().BoolVar(
		&data.saveOnly, "save-only", false,
		"Save data and exit. Requires --save-to. Example: -s ./results --save-only",
	)
	cmd.Flags().BoolVarP(
		&data.verbose, "verbose", "v", false,
		"Show test details of test failures",
	)
	cmd.Flags().BoolVar(
		&data.json, "json", false,
		"Show report in json format",
	)
	cmd.Flags().BoolVar(
		&data.skipBaselineAPI, "skip-baseline-api", false,
		"Set to disable the BsaelineAPI call to get the baseline results injected in the failure filter pipeline.",
	)
	cmd.Flags().BoolVarP(
		&data.force, "force", "f", false,
		"Force to continue the execution, skipping deprecation warnings.",
	)
	cmd.Flags().BoolVarP(
		&data.insights, "insights", "i", false,
		"Show insights and next steps in the report.",
	)
	return cmd
}

// checkFlags checks the flags and set the default values.
func checkFlags(input *Input) {
	if input.embedData {
		log.Warnf("--embed-data is set to true, forcing --server-skip to true.")
		input.serverSkip = true
	}
}

// processResult reads the artifacts and show it as an report format.
func processResult(input *Input) error {
	log.Println("Creating report...")
	timers := metrics.NewTimers()
	timers.Add("report-total")

	if input.skipBaselineAPI {
		log.Warnf("THIS IS NOT RECOMMENDED: detected flag --skip-baseline-api, setting OPCT_DISABLE_FILTER_BASELINE=1 to skip the failure filter in the pipeline")
		os.Setenv("OPCT_DISABLE_FILTER_BASELINE", "1")
	}

	// Show deprecation warnings when using --baseline.
	if input.archiveBase != "" {
		log.Warnf(`DEPRECATED: --baseline/--diff flag should not be used and will be removed soon.
Baseline are now discovered and applied to the filter pipeline automatically.
Please remove the --baseline/--diff flags from the command.
Additionally, if you want to skip the BaselineAPI filter, use --skip-baseline-api=true.`)
		if !input.force {
			log.Warnf("Aborting execution: --force flag is not set, set it if you want continue with warnings.")
			os.Exit(1)
		}
	}

	reportDir := input.saveTo
	if reportDir == "" {
		reportDir = "/tmp/opct-tmp-results-" + filepath.Base(input.archive)
	}

	cs := summary.NewConsolidatedSummary(&summary.ConsolidatedSummaryInput{
		Verbose:     input.verbose,
		Timers:      timers,
		Archive:     input.archive,
		ArchiveBase: input.archiveBase,
		SaveTo:      reportDir,
	})

	log.Debug("Processing results")
	if err := cs.Process(); err != nil {
		return fmt.Errorf("error processing results: %v", err)
	}

	re := report.NewReportData(input.embedData)
	log.Debug("Processing report")
	if err := re.Populate(cs); err != nil {
		return fmt.Errorf("error populating report: %v", err)
	}

	// show report in CLI
	if err := showReportCLI(re, input.verbose); err != nil {
		return fmt.Errorf("error showing aggregated summary: %v", err)
	}

	if len(reportDir) == 0 {
		log.Infof("No report directory specified, skipping saving results.")
		os.Exit(0)
	}

	// Generate the consolidated summary and report results
	if err := cs.SaveResults(reportDir); err != nil {
		return fmt.Errorf("error saving consolidated summary results: %v", err)
	}
	timers.Add("report-total")
	if err := re.SaveResults(reportDir); err != nil {
		return fmt.Errorf("error saving report results: %v", err)
	}
	if input.saveOnly {
		os.Exit(0)
	}

	if input.serverSkip {
		log.Infof("The report server is not enabled (--server-skip=true)., you'll need to navigate it locallly")
		log.Infof("To read the report open your browser and navigate to the path file://%s", reportDir)
		log.Infof("To get started open the report file://%s/index.html.", reportDir)
		os.Exit(0)
	}

	// start http server to serve static report
	server, err := newReportServer(input.serverAddress, reportDir)
	if err != nil {
		log.Fatalf("Unable to initialize the report server at address %s: %v", input.serverAddress, err)
	}
	if err := server.Start(); err != nil {
		log.Fatalf("Unable to start the report server at address %s: %v", input.serverAddress, err)
	}

	return nil
}
