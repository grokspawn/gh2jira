package list

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/oceanc80/gh2jira/pkg/config"
	"github.com/oceanc80/gh2jira/pkg/jira"
	"github.com/oceanc80/gh2jira/pkg/util"
)

var (
	query string
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List open Jira issues",
		Long:  "List open Jira issues filtered with optional additional JQL",
		RunE: func(cmd *cobra.Command, args []string) error {
			ff, err := util.NewFlagFeeder(cmd)
			if err != nil {
				return err
			}
			config := config.NewConfig(ff)
			err = config.Read()
			if err != nil {
				return err
			}

			jc, err := jira.NewConnection(
				jira.WithBaseURI(config.JiraBaseUrl),
				jira.WithAuth(config.Tokens.JiraAuth),
			)
			if err != nil {
				return err
			}

			err = jc.Connect()
			if err != nil {
				return err
			}

			jql := ""
			switch {
			case config.JiraProject == "" && query == "":
				return fmt.Errorf("must provide either project or query")
			case config.JiraProject != "" && query != "":
				jql = "project=" + config.JiraProject + " AND " + query
			case config.JiraProject != "":
				jql = "project=" + config.JiraProject
			default:
				jql = query
			}
			jql += " and status != Closed"

			result, err := jc.SearchIssues(cmd.Context(), jql)
			if err != nil {
				return err
			}
			for _, i := range result {
				fmt.Printf("%s (%s/%s): %+v -> %s\n", i.Key, i.Fields.Type.Name, i.Fields.Priority.Name, i.Fields.Summary, i.Fields.Status.Name)
				if i.Fields.Assignee != nil {
					fmt.Printf("Assignee : %v\n", i.Fields.Assignee.DisplayName)
				} else {
					fmt.Printf("Assignee : Unassigned\n")
				}
				fmt.Printf("Reporter: %v\n", i.Fields.Reporter.DisplayName)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "Jira query (if provided, ANDed with project)")
	return cmd
}
