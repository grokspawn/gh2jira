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
	"os"

	"sigs.k8s.io/yaml"
)

const schemaName string = "gh2jira.tokenstore"

type JiraAuth struct {
	Token    string `json:"token,omitempty"`
	Email    string `json:"email,omitempty"`
	APIToken string `json:"apiToken,omitempty"`
}

func (j JiraAuth) IsCloud() bool {
	return j.Email != "" && j.APIToken != ""
}

type TokenPair struct {
	JiraAuth    JiraAuth `json:"jira"`
	GithubToken string   `json:"github"`
}

type TokenStore struct {
	Schema string    `json:"schema"`
	Tokens TokenPair `json:"authTokens"`
}

func ReadTokenStore(f string) (*TokenStore, error) {
	b, err := readFile(f)
	if err != nil {
		return nil, err
	}

	var c TokenStore
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}
	if c.Schema != schemaName {
		return nil, fmt.Errorf("invalid schema: %q should be %q: %v", c.Schema, schemaName, err)
	}
	if c.Tokens.GithubToken == "" {
		return nil, errors.New("missing required github token")
	}

	hasToken := c.Tokens.JiraAuth.Token != ""
	hasCloud := c.Tokens.JiraAuth.Email != "" || c.Tokens.JiraAuth.APIToken != ""
	if hasToken && hasCloud {
		return nil, errors.New("jira auth: specify either token (datacenter) or email+apiToken (cloud), not both")
	}
	if !hasToken && !hasCloud {
		return nil, errors.New("missing required jira auth: provide token (datacenter) or email+apiToken (cloud)")
	}
	if hasCloud && (c.Tokens.JiraAuth.Email == "" || c.Tokens.JiraAuth.APIToken == "") {
		return nil, errors.New("jira cloud auth requires both email and apiToken")
	}

	return &c, nil
}

// overrideable func for mocking os.ReadFile
var readFile = func(file string) ([]byte, error) {
	return os.ReadFile(file)
}
