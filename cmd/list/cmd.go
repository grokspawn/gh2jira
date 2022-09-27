// Copyright © 2022 jesus m. rodriguez jmrodri@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package list

import (
	"github.com/jmrodri/gh2jira/internal/gh"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	lo := gh.ListerOptions{}
	lister := gh.Lister{
		Options: &lo,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Github issues",
		Long:  "List Github issues filtered by milestone, assignee, or label",
		RunE: func(cmd *cobra.Command, args []string) error {
			lister.ListIssues()
			return nil
		},
	}

	cmd.Flags().StringVar(&lo.Milestone, "milestone", "", "the milestone ID from the url, not the display name")
	cmd.Flags().StringVar(&lo.Assignee, "assignee", "", "username of the issue is assigned")
	cmd.Flags().StringVar(&lo.Project, "project", "operator-framework/operator-sdk",
		"Github project to list e.g. ORG/REPO")
	cmd.Flags().StringSliceVar(&lo.Label, "label", nil, "label i.e. --label \"documentation,bug\" or --label doc --label bug")

	return cmd
}
