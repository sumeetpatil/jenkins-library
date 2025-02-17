//go:build unit
// +build unit

package cmd

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/stretchr/testify/assert"
)

var executionLogStringClone string
var apiManager abaputils.SoftwareComponentApiManagerInterface

func init() {
	executionLog := abaputils.LogProtocolResults{
		Results: []abaputils.LogProtocol{
			{
				ProtocolLine:  1,
				OverviewIndex: 1,
				Type:          "LogEntry",
				Description:   "S",
				Timestamp:     "/Date(1644332299000+0000)/",
			},
		},
		Count: "1",
	}
	executionLogResponse, _ := json.Marshal(executionLog)
	executionLogStringClone = string(executionLogResponse)

}

func TestCloneStep(t *testing.T) {
	t.Run("Run Step - Successful with repositories.yml", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		body := `---
repositories:
- name: /DMO/REPO_A
  tag: v-1.0.1-build-0001
  branch: branchA
  version: 1.0.1
- name: /DMO/REPO_B
  tag: rel-2.1.1-build-0001
  branch: branchB
  version: 2.1.1
`
		file, _ := os.Create("filename.yaml")
		_, err := file.Write([]byte(body))
		assert.NoError(t, err)

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "filename.yaml",
		}

		logResultSuccess := `{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Success", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringClone + `}`,
				logResultSuccess,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "sc_name" : "/DMO/REPO_B", "avail_on_instance" : false, "active_branch": "branchB" } }`,
				`{"d" : [] }`,
				`{"d" : ` + executionLogStringClone + `}`,
				logResultSuccess,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "sc_name" : "/DMO/REPO_A", "avail_on_instance" : true, "active_branch": "branchA" } }`,
				`{"d" : [] }`,
			},
			Token: "myToken",
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond}
		err = runAbapEnvironmentCloneGitRepo(&config, &autils, apiManager)
		assert.NoError(t, err, "Did not expect error")
		assert.Equal(t, 0, len(client.BodyList), "Not all requests were done")
	})

	t.Run("Run Step - Successful with repositoryName", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
		}

		logResultSuccess := `{"d": { "sc_name": "testRepo1", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Success", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringClone + `}`,
				logResultSuccess,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "sc_name" : "testRepo1", "avail_on_instance" : false, "active_branch": "testBranch1" } }`,
				`{"d" : [] }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond}
		err := runAbapEnvironmentCloneGitRepo(&config, &autils, apiManager)
		assert.NoError(t, err, "Did not expect error")
		assert.Equal(t, 0, len(client.BodyList), "Not all requests were done")
	})

	t.Run("Run Step - failing", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "testBranch1",
		}

		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : {} }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "sc_name" : "testRepo1", "avail_on_instance" : true, "active_branch": "testBranch1" } }`,
				`{"d" : [] }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond}
		err := runAbapEnvironmentCloneGitRepo(&config, &autils, apiManager)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Clone of repository / software component 'testRepo1', branch 'testBranch1' failed on the ABAP system: Request to ABAP System not successful", err.Error(), "Expected different error message")
		}

	})
}

func TestCloneStepErrorMessages(t *testing.T) {
	t.Run("Status Error", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		body := `---
repositories:
- name: /DMO/REPO_A
  tag: v-1.0.1-build-0001
  branch: branchA
  version: 1.0.1
  commitID: ABCD1234
`
		file, _ := os.Create("filename.yaml")
		_, err := file.Write([]byte(body))
		assert.NoError(t, err)

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "filename.yaml",
		}

		logResultError := `{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Error", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringClone + `}`,
				logResultError,
				`{"d" : { "EntitySets" : [ "LogOverviews" ] } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}
		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond}
		err = runAbapEnvironmentCloneGitRepo(&config, &autils, apiManager)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Clone of repository / software component '/DMO/REPO_A', branch 'branchA', commit 'ABCD1234' failed on the ABAP system: Request to ABAP System not successful", err.Error(), "Expected different error message")
		}
	})

	t.Run("Poll Request Error", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "testBranch1",
		}

		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : {  } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}
		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond}
		err := runAbapEnvironmentCloneGitRepo(&config, &autils, apiManager)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Clone of repository / software component 'testRepo1', branch 'testBranch1' failed on the ABAP system: Request to ABAP System not successful", err.Error(), "Expected different error message")
		}
	})

	t.Run("Trigger Clone Error", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			RepositoryName:    "testRepo1",
			BranchName:        "testBranch1",
		}

		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : {  } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}
		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond}
		err := runAbapEnvironmentCloneGitRepo(&config, &autils, apiManager)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Clone of repository / software component 'testRepo1', branch 'testBranch1' failed on the ABAP system: Request to ABAP System not successful", err.Error(), "Expected different error message")
		}
	})

	t.Run("Missing file error", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "filename.yaml",
		}

		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : {} }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}
		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond}
		err := runAbapEnvironmentCloneGitRepo(&config, &autils, apiManager)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "Could not read repositories: Could not find filename.yaml", err.Error(), "Expected different error message")
		}

	})

	t.Run("Config overload", func(t *testing.T) {
		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
			Repositories:      "filename.yaml",
			RepositoryName:    "/DMO/REPO",
			BranchName:        "Branch",
		}

		logResultError := `{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Error", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
		client := &abaputils.ClientMock{
			BodyList: []string{
				logResultError,
				`{"d" : { "EntitySets" : [ "LogOverviews" ] } }`,
				`{"d" : { "status" : "E" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}
		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond}
		err := runAbapEnvironmentCloneGitRepo(&config, &autils, apiManager)
		if assert.Error(t, err, "Expected error") {
			assert.Equal(t, "The provided configuration is not allowed: It is not allowed to configure the parameters `repositories`and `repositoryName` at the same time", err.Error(), "Expected different error message")
		}
	})
}

func TestALreadyCloned(t *testing.T) {
	t.Run("Already cloned, switch branch and pull instead", func(t *testing.T) {

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.Host = "example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}

		logResultSuccess := `{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Success", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringClone + `}`,
				logResultSuccess,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : [] }`,
				`{"d" : ` + executionLogStringClone + `}`,
				logResultSuccess,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "sc_name" : "testRepo1", "avail_on_inst" : true, "active_branch": "testBranch1" } }`,
				`{"d" : [] }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		repo := abaputils.Repository{
			Name:     "testRepo1",
			Branch:   "inactie_branch",
			CommitID: "abcd1234",
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond}
		err := cloneSingleRepo(apiManager, autils.ReturnedConnectionDetailsHTTP, repo, &config, &autils)
		assert.NoError(t, err, "Did not expect error")
	})

	t.Run("Already cloned, branch is already checked out, pull instead", func(t *testing.T) {

		var autils = abaputils.AUtilsMock{}
		defer autils.Cleanup()
		autils.ReturnedConnectionDetailsHTTP.Password = "password"
		autils.ReturnedConnectionDetailsHTTP.User = "user"
		autils.ReturnedConnectionDetailsHTTP.URL = "https://example.com"
		autils.ReturnedConnectionDetailsHTTP.Host = "example.com"
		autils.ReturnedConnectionDetailsHTTP.XCsrfToken = "xcsrftoken"

		config := abapEnvironmentCloneGitRepoOptions{
			CfAPIEndpoint:     "https://api.endpoint.com",
			CfOrg:             "testOrg",
			CfSpace:           "testSpace",
			CfServiceInstance: "testInstance",
			CfServiceKeyName:  "testServiceKey",
			Username:          "testUser",
			Password:          "testPassword",
		}

		logResultSuccess := `{"d": { "sc_name": "/DMO/SWC", "status": "S", "to_Log_Overview": { "results": [ { "log_index": 1, "log_name": "Main Import", "type_of_found_issues": "Success", "timestamp": "/Date(1644332299000+0000)/", "to_Log_Protocol": { "results": [ { "log_index": 1, "index_no": "1", "log_name": "", "type": "Info", "descr": "Main import", "timestamp": null, "criticality": 0 } ] } } ] } } }`
		client := &abaputils.ClientMock{
			BodyList: []string{
				`{"d" : ` + executionLogStringClone + `}`,
				logResultSuccess,
				`{"d" : { "status" : "S" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "status" : "R" } }`,
				`{"d" : { "sc_name" : "testRepo1", "avail_on_inst" : true, "active_branch": "testBranch1" } }`,
				`{"d" : [] }`,
			},
			Token:      "myToken",
			StatusCode: 200,
		}

		repo := abaputils.Repository{
			Name:     "testRepo1",
			Branch:   "testBranch1",
			CommitID: "abcd1234",
		}

		apiManager = &abaputils.SoftwareComponentApiManager{Client: client, PollIntervall: 1 * time.Nanosecond}
		err := cloneSingleRepo(apiManager, autils.ReturnedConnectionDetailsHTTP, repo, &config, &autils)
		assert.NoError(t, err, "Did not expect error")
	})

}
