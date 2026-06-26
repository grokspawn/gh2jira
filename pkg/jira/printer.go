package jira

import (
	"context"
	"fmt"
)

func PrintJiraIssue(ctx context.Context, jc *Connection, jiraIssue Issue) {
	fmt.Printf("%s (%s/%s): %+v -> %s\n", jiraIssue.Key, jiraIssue.Fields.Type.Name, jiraIssue.Fields.Priority.Name, jiraIssue.Fields.Summary, jiraIssue.Fields.Status.Name)
	if jiraIssue.Fields.Assignee != nil {
		fmt.Printf("\tAssignee : %v\n", jiraIssue.Fields.Assignee.DisplayName)
	} else {
		fmt.Printf("\tAssignee : Unassigned\n")
	}
	fmt.Printf("\tReporter: %v\n", jiraIssue.Fields.Reporter.DisplayName)
	fmt.Printf("\tSummary: %s\n", jiraIssue.Fields.Summary)
	rlinks, err := jc.GetRemoteLinks(ctx, jiraIssue.Key)
	if err != nil {
		return
	}
	fmt.Printf("\tLinks:\n")
	for _, rlink := range rlinks {
		if rlink.Object != nil {
			fmt.Printf("\t\t%s\n", rlink.Object.URL)
		}
	}
	fmt.Println("")
}

func PrintJiraIssues(ctx context.Context, jc *Connection, jiraIssues []Issue) {
	for _, ji := range jiraIssues {
		PrintJiraIssue(ctx, jc, ji)
	}
}
