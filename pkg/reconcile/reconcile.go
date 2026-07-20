package reconcile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/oceanc80/gh2jira/pkg/gh"
	"github.com/oceanc80/gh2jira/pkg/jira"
	"github.com/oceanc80/gh2jira/pkg/workflow"
)

const unassigned_issue string = "unassigned"

type IssueStatus struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Assignee string `json:"assignee"`
}

type PairResult struct {
	Jira IssueStatus `json:"jira"`
	Git  IssueStatus `json:"github"`
}

type PairResults []PairResult
type TypeResults struct {
	Matches    PairResults `json:"matches"`
	Mismatches PairResults `json:"mismatches"`
}

type Outcome string

const (
	OutcomeMatch    Outcome = "MATCH"
	OutcomeMismatch Outcome = "MISMATCH"
)

func Reconcile(ctx context.Context, jql string, jc *jira.Connection, gc *gh.Connection, wf io.Reader) (*TypeResults, error) {
	results := &TypeResults{
		Matches:    make(PairResults, 0),
		Mismatches: make(PairResults, 0),
	}

	if jc == nil || gc == nil {
		return nil, errors.New("nil connection")
	}

	r, err := regexp.Compile(".*/github.com/.*/issues/([0-9]+)")
	if err != nil {
		return nil, err
	}

	jiraIssues, err := jc.SearchIssues(ctx, jql)
	if err != nil {
		return nil, err
	}

	jiraIssues = slices.DeleteFunc(jiraIssues, func(issue jira.Issue) bool {
		rlinks, err := jc.GetRemoteLinks(ctx, issue.Key)
		if err != nil {
			return false
		}
		found := false
		for _, rlink := range rlinks {
			if rlink.Object != nil && r.MatchString(rlink.Object.URL) {
				found = true
			}
		}
		return !found
	})

	err = workflow.ReadWorkflows(wf)
	if err != nil {
		return nil, err
	}

	for _, ji := range jiraIssues {
		jstat := ji.Fields.Status.Name
		rlinks, err := jc.GetRemoteLinks(ctx, ji.Key)
		if err != nil {
			return nil, err
		}
		for _, rlink := range rlinks {
			if rlink.Object != nil && r.MatchString(rlink.Object.URL) {
				project, issue, err := splitIssueRef(rlink.Object.URL)
				if err != nil {
					return nil, err
				}
				gi, err := gc.GetIssue(issue, gh.WithProject(project))
				if err != nil {
					return nil, err
				}
				stateMatch, err := workflow.ValidateState(gi.GetState(), jstat)
				if err != nil {
					return nil, err
				}
				var ghAssignee string = unassigned_issue
				if gi.GetAssignee() != nil {
					ghAssignee = *gi.GetAssignee().Login
				}
				var jiAssignee string = unassigned_issue
				if ji.Fields.Assignee != nil {
					jiAssignee = ji.Fields.Assignee.DisplayName
				}

				pair := PairResult{
					Jira: IssueStatus{Name: ji.Key, Status: jstat, Assignee: jiAssignee},
					Git:  IssueStatus{Name: fmt.Sprintf("%s/%d", project, gi.GetNumber()), Status: gi.GetState(), Assignee: ghAssignee},
				}
				if stateMatch {
					results.Matches = append(results.Matches, pair)
				} else {
					results.Mismatches = append(results.Mismatches, pair)
				}
			}
		}
	}

	return results, nil
}

func splitIssueRef(ref string) (string, int, error) {
	s := strings.Split(ref, "/")
	if len(s) <= 4 {
		return "", 0, fmt.Errorf("unable to extract issue attributes from URL: %v", ref)
	}

	project := fmt.Sprintf("%s/%s", s[len(s)-4], s[len(s)-3])
	num, err := strconv.Atoi(s[len(s)-1])
	if err != nil {
		return "", 0, err
	}

	return project, num, nil
}
