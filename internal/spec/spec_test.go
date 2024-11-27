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
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cloudbase/garm-provider-aws/config"
	"github.com/cloudbase/garm-provider-common/cloudconfig"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/stretchr/testify/require"
)

func TestExtraSpecsFromBootstrapData(t *testing.T) {
	tests := []struct {
		name           string
		input          params.BootstrapInstance
		expectedOutput *extraSpecs
		errString      string
	}{
		{
			name: "valid bootstrap data",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"subnet_id": "subnet-0a0a0a0a0a0a0a0a0", "ssh_key_name": "ssh_key_name", "security_group_ids": ["sg-018c35963edfb1cce", "sg-018c35963edfb1cee"], "iops": 3000, "throughput": 200, "volume_size": 50, "volume_type": "gp3", "disable_updates": true, "enable_boot_debug": true, "extra_packages": ["package1", "package2"], "runner_install_template": "IyEvYmluL2Jhc2gKZWNobyBJbnN0YWxsaW5nIHJ1bm5lci4uLg==", "pre_install_scripts": {"setup.sh": "IyEvYmluL2Jhc2gKZWNobyBTZXR1cCBzY3JpcHQuLi4="}, "extra_context": {"key": "value"}}`),
			},
			expectedOutput: &extraSpecs{
				SubnetID:         aws.String("subnet-0a0a0a0a0a0a0a0a0"),
				Iops:             aws.Int32(3000),
				Throughput:       aws.Int32(200),
				VolumeSize:       aws.Int32(50),
				VolumeType:       types.VolumeTypeGp3,
				SSHKeyName:       aws.String("ssh_key_name"),
				SecurityGroupIds: []string{"sg-018c35963edfb1cce", "sg-018c35963edfb1cee"},
				DisableUpdates:   aws.Bool(true),
				EnableBootDebug:  aws.Bool(true),
				ExtraPackages:    []string{"package1", "package2"},
				CloudConfigSpec: cloudconfig.CloudConfigSpec{
					RunnerInstallTemplate: []byte("#!/bin/bash\necho Installing runner..."),
					PreInstallScripts: map[string][]byte{
						"setup.sh": []byte("#!/bin/bash\necho Setup script..."),
					},
					ExtraContext: map[string]string{"key": "value"},
				},
			},
			errString: "",
		},
		{
			name: "specs just with subnet_id",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"subnet_id": "subnet-0a0a0a0a0a0a0a0a0"}`),
			},
			expectedOutput: &extraSpecs{
				SubnetID: aws.String("subnet-0a0a0a0a0a0a0a0a0"),
			},
			errString: "",
		},
		{
			name: "specs just with ssh_key_name",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"ssh_key_name": "ssh_key_name"}`),
			},
			expectedOutput: &extraSpecs{
				SSHKeyName: aws.String("ssh_key_name"),
			},
			errString: "",
		},
		{
			name: "specs just with security_group_ids",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"security_group_ids": ["sg-018c35963edfb1cce", "sg-018c35963edfb1cee"]}`),
			},
			expectedOutput: &extraSpecs{
				SecurityGroupIds: []string{"sg-018c35963edfb1cce", "sg-018c35963edfb1cee"},
			},
		},
		{
			name: "specs just with disable_updates",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"disable_updates": true}`),
			},
			expectedOutput: &extraSpecs{
				DisableUpdates: aws.Bool(true),
			},
			errString: "",
		},
		{
			name: "specs just with enable_boot_debug",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"enable_boot_debug": true}`),
			},
			expectedOutput: &extraSpecs{
				EnableBootDebug: aws.Bool(true),
			},
			errString: "",
		},
		{
			name: "specs just with extra_packages",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"extra_packages": ["package1", "package2"]}`),
			},
			expectedOutput: &extraSpecs{
				ExtraPackages: []string{"package1", "package2"},
			},
			errString: "",
		},
		{
			name: "spec just with RunnerInstallTemplate",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"runner_install_template": "IyEvYmluL2Jhc2gKZWNobyBJbnN0YWxsaW5nIHJ1bm5lci4uLg=="}`),
			},
			expectedOutput: &extraSpecs{
				CloudConfigSpec: cloudconfig.CloudConfigSpec{
					RunnerInstallTemplate: []byte("#!/bin/bash\necho Installing runner..."),
				},
			},
			errString: "",
		},
		{
			name: "spec just with PreInstallScripts",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"pre_install_scripts": {"setup.sh": "IyEvYmluL2Jhc2gKZWNobyBTZXR1cCBzY3JpcHQuLi4="}}`),
			},
			expectedOutput: &extraSpecs{
				CloudConfigSpec: cloudconfig.CloudConfigSpec{
					PreInstallScripts: map[string][]byte{
						"setup.sh": []byte("#!/bin/bash\necho Setup script..."),
					},
				},
			},
			errString: "",
		},
		{
			name: "spec just with ExtraContext",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"extra_context": {"key": "value"}}`),
			},
			expectedOutput: &extraSpecs{
				CloudConfigSpec: cloudconfig.CloudConfigSpec{
					ExtraContext: map[string]string{"key": "value"},
				},
			},
			errString: "",
		},
		{
			name: "missing extra specs",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{}`),
			},
			expectedOutput: &extraSpecs{},
			errString:      "",
		},
		{
			name: "invalid format for subnet_id",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"subnet_id": "subnet-1"}`),
			},
			expectedOutput: nil,
			errString:      "subnet_id: Does not match pattern '^subnet-[0-9a-fA-F]{17}$'",
		},
		{
			name: "invalid type for subnet_id",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"subnet_id": true}`),
			},
			expectedOutput: nil,
			errString:      "subnet_id: Invalid type. Expected: string, given: boolean",
		},
		{
			name: "invalid type for ssh_key_name",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"ssh_key_name": 123}`),
			},
			expectedOutput: nil,
			errString:      "ssh_key_name: Invalid type. Expected: string, given: integer",
		},
		{
			name: "invalid type for security_group_ids",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"security_group_ids": "sg-018c35963edfb1cce"}`),
			},
			expectedOutput: nil,
			errString:      "security_group_ids: Invalid type. Expected: array, given: string",
		},
		{
			name: "invalid type for disable_updates",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"disable_updates": "true"}`),
			},
			expectedOutput: nil,
			errString:      "disable_updates: Invalid type. Expected: boolean, given: string",
		},
		{
			name: "invalid type for enable_boot_debug",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"enable_boot_debug": "true"}`),
			},
			expectedOutput: nil,
			errString:      "enable_boot_debug: Invalid type. Expected: boolean, given: string",
		},
		{
			name: "invalid type for extra_packages",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"extra_packages": "package1"}`),
			},
			expectedOutput: nil,
			errString:      "extra_packages: Invalid type. Expected: array, given: string",
		},
		{
			name: "invalid type for runner_install_template",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"runner_install_template": 123}`),
			},
			expectedOutput: nil,
			errString:      "runner_install_template: Invalid type. Expected: string, given: integer",
		},
		{
			name: "invalid type for pre_install_scripts",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"pre_install_scripts": "setup.sh"}`),
			},
			expectedOutput: nil,
			errString:      "pre_install_scripts: Invalid type. Expected: object, given: string",
		},
		{
			name: "invalid type for extra_context",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"extra_context": 123}`),
			},
			expectedOutput: nil,
			errString:      "extra_context: Invalid type. Expected: object, given: integer",
		},
		{
			name: "invalid input - additional property",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"invalid": "invalid"}`),
			},
			expectedOutput: nil,
			errString:      "Additional property invalid is not allowed",
		},
		{
			name: "invalid input - invalid json",
			input: params.BootstrapInstance{
				ExtraSpecs: json.RawMessage(`{"disable_updates": }`),
			},
			expectedOutput: nil,
			errString:      "failed to validate extra specs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := newExtraSpecsFromBootstrapData(tt.input)
			if tt.errString == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			}
			require.Equal(t, tt.expectedOutput, output)
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
		Name:       "mock-name",
		ExtraSpecs: json.RawMessage(`{"subnet_id": "subnet-0a0a0a0a0a0a0a0a0", "ssh_key_name": "ssh_key_name", "security_group_ids": ["sg-018c35963edfb1cce", "sg-018c35963edfb1cee"], "iops": 3000, "throughput": 200, "volume_size": 50, "volume_type": "gp3", "disable_updates": true, "enable_boot_debug": true, "extra_packages": ["package1", "package2"], "runner_install_template": "IyEvYmluL2Jhc2gKZWNobyBJbnN0YWxsaW5nIHJ1bm5lci4uLg==", "pre_install_scripts": {"setup.sh": "IyEvYmluL2Jhc2gKZWNobyBTZXR1cCBzY3JpcHQuLi4="}, "extra_context": {"key": "value"}}`),
	}

	config := &config.Config{
		Credentials: config.Credentials{
			CredentialType: config.AWSCredentialTypeStatic,
			StaticCredentials: config.StaticCredentials{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
			},
		},
		SubnetID: "subnet_id",
		Region:   "region",
	}
	expectedRunnerSpec := &RunnerSpec{
		Region:           "region",
		DisableUpdates:   true,
		ExtraPackages:    []string{"package1", "package2"},
		EnableBootDebug:  true,
		SubnetID:         "subnet-0a0a0a0a0a0a0a0a0",
		Tools:            Mocktools,
		ControllerID:     "controller_id",
		BootstrapParams:  data,
		SSHKeyName:       aws.String("ssh_key_name"),
		SecurityGroupIds: []string{"sg-018c35963edfb1cce", "sg-018c35963edfb1cee"},
		Iops:             aws.Int32(3000),
		VolumeType:       types.VolumeTypeGp3,
		Throughput:       aws.Int32(200),
		VolumeSize:       aws.Int32(50),
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
				Iops:       aws.Int32(3000),
				VolumeType: types.VolumeTypeGp3,
				Throughput: aws.Int32(200),
				VolumeSize: aws.Int32(50),
			},
		},
		{
			name: "valid runner spec with io2 volume type",
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
				Iops:       aws.Int32(3000),
				VolumeType: types.VolumeTypeIo2,
				VolumeSize: aws.Int32(50),
			},
		},
		{
			name: "Bad runner config just with iops",
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
				Iops: aws.Int32(3000),
			},
			errString: "EBS iops is only valid for volume types io1, io2 and gp3",
		},
		{
			name: "Bad runner config just with throughput",
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
				Throughput: aws.Int32(200),
			},
			errString: "EBS throughput is only valid for volume type gp3",
		},
		{
			name: "Bad runner config with both iops and throughput",
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
				Iops:       aws.Int32(3000),
				Throughput: aws.Int32(200),
			},
			errString: "EBS iops is only valid for volume types io1, io2 and gp3",
		},
		{
			name: "Bad runner config with invalid iops for io1",
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
				Iops:       aws.Int32(50),
				VolumeType: types.VolumeTypeIo1,
			},
			errString: "EBS iops for volume type io1 must be between 100 and 64000",
		},
		{
			name: "Bad runner config with invalid iops for io2",
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
				Iops:       aws.Int32(50),
				VolumeType: types.VolumeTypeIo2,
			},
			errString: "EBS iops for volume type io2 must be between 100 and 256000",
		},
		{
			name: "Bad runner config with invalid iops for gp3",
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
				Iops:       aws.Int32(50),
				VolumeType: types.VolumeTypeGp3,
			},
			errString: "EBS iops for volume type gp3 must be between 3000 and 16000",
		},
		{
			name: "Bad runner config with invalid volume size for gp3",
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
				VolumeSize: aws.Int32(173482),
				VolumeType: types.VolumeTypeGp3,
			},
			errString: "EBS volume size for volume type gp3 must be between 1 and 16384",
		},
		{
			name: "Bad runner config with invalid volume size for io1",
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
				VolumeSize: aws.Int32(2),
				VolumeType: types.VolumeTypeIo1,
			},
			errString: "EBS volume size for volume type io1 must be between 4 and 16384",
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
				SubnetID:         aws.String("subnet-0a0a0a0a0a0a0a0a0"),
				SSHKeyName:       aws.String("ssh_key_name"),
				SecurityGroupIds: []string{"sg-018c35963edfb1cce", "sg-018c35963edfb1cee"},
				Iops:             aws.Int32(3000),
				Throughput:       aws.Int32(200),
				VolumeSize:       aws.Int32(50),
				VolumeType:       types.VolumeTypeGp3,
				DisableUpdates:   aws.Bool(true),
				EnableBootDebug:  aws.Bool(true),
				ExtraPackages:    []string{"package1", "package2"},
			},
			expected: &RunnerSpec{
				SubnetID:         "subnet-0a0a0a0a0a0a0a0a0",
				SSHKeyName:       aws.String("ssh_key_name"),
				SecurityGroupIds: []string{"sg-018c35963edfb1cce", "sg-018c35963edfb1cee"},
				Iops:             aws.Int32(3000),
				Throughput:       aws.Int32(200),
				VolumeSize:       aws.Int32(50),
				VolumeType:       types.VolumeTypeGp3,
				DisableUpdates:   true,
				EnableBootDebug:  true,
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
