package jira

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v60/github"
)

func getDomainFromIssueUrl(url string) string {
	if url == "" {
		return ""
	}

	parts := strings.Split(url, "/")
	parts = parts[:len(parts)-2]
	parts = parts[len(parts)-2:]
	return strings.Join(parts, "/")
}

func getIssueNumberFromIssueUrl(url string) string {
	if url == "" {
		return ""
	}

	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}

func expandDescription(body, url string) (string, error) {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	var out []string

	matcher := regexp.MustCompile(`^- \[ \] #[0-9]+`)

	for _, line := range lines {
		if matcher.FindStringIndex(line) != nil {
			parts := strings.Split(line, "#")
			elements := strings.Split(parts[1], " ")
			issue := elements[0]
			var trailer string
			if len(elements) > 1 {
				trailer = strings.Join(elements[1:], " ")
			}
			out = append(out,
				fmt.Sprintf(" * [%s|%s] %s",
					fmt.Sprintf("%s#%s", getDomainFromIssueUrl(url), issue),
					strings.Replace(url, getIssueNumberFromIssueUrl(url), issue, -1),
					trailer))
		} else {
			out = append(out, line)
		}
	}

	return strings.Join(out, "\n"), nil
}

func (conn *Connection) Clone(ctx context.Context, fromIssue *github.Issue, project string, issueType string, dryRun bool) (*Issue, error) {
	if conn.jiraClient == nil {
		if err := conn.Connect(); err != nil {
			return nil, err
		}
	}

	description, err := expandDescription(fromIssue.GetBody(), fromIssue.GetHTMLURL())
	if err != nil {
		return nil, err
	}

	ji := Issue{
		Fields: &IssueFields{
			Description: description,
			Type: IssueType{
				Name: issueType,
			},
			Project: Project{
				Key: project,
			},
			Summary: fmt.Sprintf("[UPSTREAM] %s #%d", fromIssue.GetTitle(), fromIssue.GetNumber()),
		},
	}

	var daIssue *Issue

	if dryRun {
		fmt.Println("\n############# DRY RUN MODE #############")
		fmt.Printf("Cloning issue #%d to jira project board: %s\n\n", fromIssue.GetNumber(), ji.Fields.Project.Key)
		fmt.Printf("Summary: %s\n", ji.Fields.Summary)
		fmt.Printf("Type: %s\n", ji.Fields.Type.Name)
		fmt.Println("Description:")
		fmt.Printf("%s\n", ji.Fields.Description)
		fmt.Printf("Domain: %s\n", getDomainFromIssueUrl(fromIssue.GetHTMLURL()))
		fmt.Println("\n############# DRY RUN MODE #############")
	} else {
		fmt.Printf("Cloning issue #%d to jira project board: %s\n\n", fromIssue.GetNumber(), ji.Fields.Project.Key)

		daIssue, err = conn.CreateIssue(ctx, &ji)
		if err != nil {
			fmt.Printf("Error cloning issue: %v\n", err)
			return daIssue, err
		}

		if daIssue != nil {
			fmt.Printf("Issue cloned; see %s\n",
				fmt.Sprintf(filepath.Join(conn.baseUri, "browse/%s"), daIssue.Key))
		}
		if err = conn.AddRemoteLink(ctx, daIssue.ID, &RemoteLink{
			Object: &RemoteLinkObject{
				URL:   fromIssue.GetHTMLURL(),
				Title: fmt.Sprintf("%s#%v", getDomainFromIssueUrl(fromIssue.GetHTMLURL()), fromIssue.GetNumber()),
			},
		}); err != nil {
			return nil, err
		}
	}

	return daIssue, nil
}
