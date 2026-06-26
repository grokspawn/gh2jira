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

package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	expectedGhToken   string = "foo"
	expectedJiraToken string = "bar"
	expectedEmail     string = "user@example.com"
	expectedAPIToken  string = "cloud-api-token"

	mockReadFileDC = func(file string) ([]byte, error) {
		data := fmt.Sprintf(`
schema: gh2jira.tokenstore
authTokens:
  jira:
    token: %s
  github: %s
`, expectedJiraToken, expectedGhToken)
		return []byte(data), nil
	}
	mockReadFileCloud = func(file string) ([]byte, error) {
		data := fmt.Sprintf(`
schema: gh2jira.tokenstore
authTokens:
  jira:
    email: %s
    apiToken: %s
  github: %s
`, expectedEmail, expectedAPIToken, expectedGhToken)
		return []byte(data), nil
	}
	mockReadFileBadFile = func(file string) ([]byte, error) {
		return nil, errors.New("oh no!")
	}
	mockReadFileBadYaml = func(file string) ([]byte, error) {
		data := `
schema: gh2jira.tokenstore
authTokens:
jira= bar
github: foo
`
		return []byte(data), nil
	}
	mockReadFileMissingGhToken = func(file string) ([]byte, error) {
		data := `
schema: gh2jira.tokenstore
authTokens:
  jira:
    token: foo
`
		return []byte(data), nil
	}
	mockReadFileMissingJira = func(file string) ([]byte, error) {
		data := `
schema: gh2jira.tokenstore
authTokens:
  github: bar
`
		return []byte(data), nil
	}
	mockReadFileBothModes = func(file string) ([]byte, error) {
		data := `
schema: gh2jira.tokenstore
authTokens:
  jira:
    token: dc-token
    email: user@example.com
    apiToken: cloud-token
  github: bar
`
		return []byte(data), nil
	}
	mockReadFileCloudPartial = func(file string) ([]byte, error) {
		data := `
schema: gh2jira.tokenstore
authTokens:
  jira:
    email: user@example.com
  github: bar
`
		return []byte(data), nil
	}
)

func TestReadFile(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(file string) ([]byte, error)
		audit   func(t *testing.T, ts *TokenStore)
		wantErr bool
	}{
		{
			name: "DC auth with token",
			mock: mockReadFileDC,
			audit: func(t *testing.T, ts *TokenStore) {
				require.Equal(t, expectedGhToken, ts.Tokens.GithubToken)
				require.Equal(t, expectedJiraToken, ts.Tokens.JiraAuth.Token)
				require.False(t, ts.Tokens.JiraAuth.IsCloud())
			},
		},
		{
			name: "Cloud auth with email and apiToken",
			mock: mockReadFileCloud,
			audit: func(t *testing.T, ts *TokenStore) {
				require.Equal(t, expectedGhToken, ts.Tokens.GithubToken)
				require.Equal(t, expectedEmail, ts.Tokens.JiraAuth.Email)
				require.Equal(t, expectedAPIToken, ts.Tokens.JiraAuth.APIToken)
				require.True(t, ts.Tokens.JiraAuth.IsCloud())
			},
		},
		{
			name:    "bad file",
			mock:    mockReadFileBadFile,
			wantErr: true,
		},
		{
			name:    "bad yaml",
			mock:    mockReadFileBadYaml,
			wantErr: true,
		},
		{
			name:    "missing github token",
			mock:    mockReadFileMissingGhToken,
			wantErr: true,
		},
		{
			name:    "missing jira auth entirely",
			mock:    mockReadFileMissingJira,
			wantErr: true,
		},
		{
			name:    "both DC and Cloud set",
			mock:    mockReadFileBothModes,
			wantErr: true,
		},
		{
			name:    "cloud partial - email without apiToken",
			mock:    mockReadFileCloudPartial,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readFile = tt.mock
			ts, err := ReadTokenStore("")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.audit != nil {
					tt.audit(t, ts)
				}
			}
		})
	}
}
