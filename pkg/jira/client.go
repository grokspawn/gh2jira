package jira

import "context"

type JiraClient interface {
	SearchIssues(ctx context.Context, jql string) ([]Issue, error)
	CreateIssue(ctx context.Context, issue *Issue) (*Issue, error)
	AddRemoteLink(ctx context.Context, issueID string, link *RemoteLink) error
	GetRemoteLinks(ctx context.Context, issueKey string) ([]RemoteLink, error)
}
