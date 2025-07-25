// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package root

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/oceanc80/gh2jira/pkg/config"
	"github.com/oceanc80/gh2jira/pkg/gh"
	"github.com/oceanc80/gh2jira/pkg/jira"
	"github.com/oceanc80/gh2jira/pkg/reconcile"
	"github.com/oceanc80/gh2jira/pkg/util"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

const (
	greenStart          string = "\033[32m"
	yellowStart         string = "\033[33m"
	redStart            string = "\033[31m"
	colorReset          string = "\033[0m"
	defaultWorkflowFile string = "workflows.yaml"
)

func NewReconcileCmd() *cobra.Command {
	var (
		output       string = "json"
		workflowFile string
	)
	runCmd := &cobra.Command{
		Use:   "reconcile",
		Short: "reconcile github and jira issues",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {

			var outputFunc func(data interface{}) ([]byte, error)

			switch output {
			case "yaml":
				outputFunc = yamlOutput
			case "json":
				outputFunc = jsonOutput
			case "table":
				fallthrough
			default:
				outputFunc = tableOutput
			}

			if workflowFile == "" {
				workflowFile = defaultWorkflowFile
			}

			wfReader, err := os.Open(workflowFile)
			if err != nil {
				return fmt.Errorf("failed to open workflow file %q: %w", workflowFile, err)
			}
			defer wfReader.Close()

			ff, err := util.NewFlagFeeder(cmd)
			if err != nil {
				return err
			}

			config := config.NewConfig(ff)
			err = config.Read()
			if err != nil {
				return err
			}

			if config.JiraProject == "" {
				return fmt.Errorf("must specify jira project")
			}
			jql := fmt.Sprintf("project=%s and status != Closed", config.JiraProject)

			gc, err := gh.NewConnection(gh.WithContext(cmd.Context()), gh.WithToken(config.Tokens.GithubToken))
			if err != nil {
				return err
			}
			err = gc.Connect()
			if err != nil {
				return err
			}

			if output != "yaml" && output != "json" {
				return fmt.Errorf("invalid output format %q (accepted formats are 'yaml', 'json')", output)
			}

			jc, err := jira.NewConnection(
				jira.WithBaseURI(config.JiraBaseUrl),
				jira.WithAuthToken(config.Tokens.JiraToken),
			)
			if err != nil {
				return err
			}

			err = jc.Connect()
			if err != nil {
				return err
			}

			results, err := reconcile.Reconcile(cmd.Context(), jql, jc, gc, wfReader)
			if err != nil {
				return err
			}

			b, err := outputFunc(results)
			if err != nil {
				return err
			}
			os.Stdout.Write(b)

			return nil
		},
	}

	runCmd.Flags().StringVarP(&output, "output", "o", "json", "output format (json, yaml, table)")
	runCmd.Flags().StringVar(&workflowFile, "workflow-file", "", "file containing the workflow definitions (if not using the default workflow)")

	return runCmd
}

func yamlOutput(data interface{}) ([]byte, error) {
	b, _ := json.MarshalIndent(data, "", "  ")
	yamlData, err := yaml.JSONToYAML(b)
	if err != nil {
		return nil, err
	}
	yamlData = append([]byte("---\n"), yamlData...)
	return yamlData, nil
}

func jsonOutput(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

func tableOutput(data interface{}) ([]byte, error) {
	results, ok := data.(*reconcile.TypeResults)
	if !ok {
		return nil, fmt.Errorf("expected TypeResults, got %T", data)
	}
	buf := new(bytes.Buffer)
	tw := tabwriter.NewWriter(buf, 0, 4, 1, '\t', 0)

	if len(results.Matches) == 0 && len(results.Mismatches) == 0 {
		fmt.Fprintln(tw, "no issues found")
	} else {
		fmt.Fprintf(tw, "found %v mismatch / %v match issues\n", len(results.Mismatches), len(results.Matches))
	}

	for _, pair := range results.Mismatches {
		var result string = "MISMATCH"
		var resultColor string = redStart
		fmt.Fprintf(tw, "%s%s|(%s)%s\n\tstatus (%q\t| %q)\t%s%s%s assignees(%q\t| %q)\n",
			yellowStart, pair.Jira.Name, pair.Git.Name, colorReset, pair.Jira.Status, pair.Git.Status, resultColor, result, colorReset, pair.Jira.Assignee, pair.Git.Assignee)
	}
	for _, pair := range results.Matches {
		var result string = "MATCH"
		var resultColor string = greenStart
		fmt.Fprintf(tw, "%s%s|(%s)%s\n\tstatus (%q\t| %q)\t%s%s%s assignees(%q\t| %q)\n",
			yellowStart, pair.Jira.Name, pair.Git.Name, colorReset, pair.Jira.Status, pair.Git.Status, resultColor, result, colorReset, pair.Jira.Assignee, pair.Git.Assignee)
	}
	tw.Flush()

	return buf.Bytes(), nil
}
