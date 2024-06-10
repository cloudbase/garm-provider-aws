// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Cloudbase Solutions SRL
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package spec

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/cloudbase/garm-provider-aws/config"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/stretchr/testify/require"
)

func TestJsonSchemaValidation(t *testing.T) {
	tests := []struct {
		name      string
		schema    json.RawMessage
		errString string
	}{
		{
			name:      "valid schema",
			schema:    json.RawMessage(`{"subnet_id": "subnet-0a0a0a0a0a0a0a0a0", "ssh_key_name": "ssh_key_name", "disable_updates": true, "enable_boot_debug": true, "extra_packages": ["package1", "package2"], "runner_install_template": "runner_install_template", "extra_context": {"key": "value"}, "pre_install_scripts":{"script1": "script1", "script2": "script2"}}`),
			errString: "",
		},
		{
			name:      "invalid schema",
			schema:    json.RawMessage(`{"subnet_id": "invalid_subnet_id"}`),
			errString: "schema validation failed: [subnet_id: Does not match pattern '^subnet-[0-9a-fA-F]{17}$']",
		},
		{
			name:      "extra argument schema",
			schema:    json.RawMessage(`{"invalid_key": "invalid_value"}`),
			errString: "Additional property invalid_key is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jsonSchemaValidation(tt.schema)
			if tt.errString == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}

func TestExtraSpecsFromBootstrapData(t *testing.T) {
	tests := []struct {
		name      string
		bootstrap params.BootstrapInstance
		errString string
	}{
		{
			name: "valid bootstrap data",
			bootstrap: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"subnet_id": "subnet-0a0a0a0a0a0a0a0a0", "ssh_key_name": "ssh_key_name", "disable_updates": true, "enable_boot_debug": true, "extra_packages": ["package1", "package2"], "runner_install_template": "runner_install_template", "extra_context": {"key": "value"}, "pre_install_scripts":{"script1": "script1", "script2": "script2"}}`),
			},
			errString: "",
		},
		{
			name: "invalid bootstrap data",
			bootstrap: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"subnet_id": "invalid_subnet_id"}`),
			},
			errString: "schema validation failed: [subnet_id: Does not match pattern '^subnet-[0-9a-fA-F]{17}$']",
		},
		{
			name: "missing extra specs",
			bootstrap: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{}`),
			},
			errString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newExtraSpecsFromBootstrapData(tt.bootstrap)
			if tt.errString == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}

func TestGetRunnerSpecFromBootstrapParams(t *testing.T) {
	Mocktools := params.RunnerApplicationDownload{
		OS:           aws.String("linux"),
		Architecture: aws.String("amd64"),
		DownloadURL:  aws.String("MockURL"),
		Filename:     aws.String("garm-runner"),
	}
	DefaultToolFetch = func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error) {
		return Mocktools, nil
	}

	data := params.BootstrapInstance{
		ExtraSpecs: json.RawMessage(`{"subnet_id": "subnet-0a0a0a0a0a0a0a0a0", "ssh_key_name": "ssh_key_name", "disable_updates": true, "enable_boot_debug": true, "extra_packages": ["package1", "package2"], "runner_install_template": "runner_install_template", "extra_context": {"key": "value"}, "pre_install_scripts":{"script1": "script1", "script2": "script2"}}`),
	}

	config := &config.Config{
		Credentials: config.Credentials{
			CredentialType: config.AWSCredentialTypeAccessKey,
			AccessKey: config.AccessKeyCredentials{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
			},
		},
		SubnetID: "subnet_id",
		Region:   "region",
	}
	expectedRunnerSpec := &RunnerSpec{
		Region:          "region",
		DisableUpdates:  true,
		ExtraPackages:   []string{"package1", "package2"},
		EnableBootDebug: true,
		SubnetID:        "subnet-0a0a0a0a0a0a0a0a0",
		Tools:           Mocktools,
		ControllerID:    "controller_id",
		BootstrapParams: data,
		SSHKeyName:      aws.String("ssh_key_name"),
	}

	runnerSpec, err := GetRunnerSpecFromBootstrapParams(config, data, "controller_id")
	require.NoError(t, err)
	require.Equal(t, expectedRunnerSpec, runnerSpec)
}

func TestRunnerSpecValidate(t *testing.T) {
	tests := []struct {
		name      string
		spec      *RunnerSpec
		errString string
	}{
		{
			name:      "empty runner spec",
			spec:      &RunnerSpec{},
			errString: "missing region",
		},
		{
			name: "missing bootstrap params",
			spec: &RunnerSpec{
				Region:       "region",
				SubnetID:     "subnet_id",
				ControllerID: "controller_id",
			},
			errString: "missing bootstrap params",
		},
		{
			name: "valid runner spec",
			spec: &RunnerSpec{
				Region:          "region",
				DisableUpdates:  true,
				ExtraPackages:   []string{"package1", "package2"},
				EnableBootDebug: true,
				Tools: params.RunnerApplicationDownload{
					OS:           aws.String("linux"),
					Architecture: aws.String("amd64"),
					DownloadURL:  aws.String("MockURL"),
					Filename:     aws.String("garm-runner"),
				},
				SubnetID:     "subnet_id",
				ControllerID: "controller_id",
				BootstrapParams: params.BootstrapInstance{
					Name: "name",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.errString == "" {
				require.Nil(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			}
		})
	}
}

func TestMergeExtraSpecs(t *testing.T) {
	tests := []struct {
		name     string
		spec     *RunnerSpec
		extra    *extraSpecs
		expected *RunnerSpec
	}{
		{
			name: "empty extra specs",
			spec: &RunnerSpec{
				SubnetID: "subnet_id",
			},
			extra:    &extraSpecs{},
			expected: &RunnerSpec{SubnetID: "subnet_id"},
		},
		{
			name: "valid extra specs",
			spec: &RunnerSpec{
				SubnetID: "subnet_id",
			},
			extra: &extraSpecs{
				SubnetID:        aws.String("subnet-0a0a0a0a0a0a0a0a0"),
				SSHKeyName:      aws.String("ssh_key_name"),
				DisableUpdates:  aws.Bool(true),
				EnableBootDebug: aws.Bool(true),
				ExtraPackages:   []string{"package1", "package2"},
			},
			expected: &RunnerSpec{
				SubnetID:        "subnet-0a0a0a0a0a0a0a0a0",
				SSHKeyName:      aws.String("ssh_key_name"),
				DisableUpdates:  true,
				EnableBootDebug: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.spec.MergeExtraSpecs(tt.extra)
			require.Equal(t, tt.expected, tt.spec)
		})
	}
}
