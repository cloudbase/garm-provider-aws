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

package client

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cloudbase/garm-provider-aws/config"
	"github.com/cloudbase/garm-provider-aws/internal/spec"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStartInstance(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Region:   "us-west-2",
		SubnetID: "subnet-1234567890abcdef0",
		Credentials: config.Credentials{
			CredentialType: config.AWSCredentialTypeStatic,
			StaticCredentials: config.StaticCredentials{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
			},
		},
	}
	mockClient := new(MockComputeClient)
	instanceId := "i-1234567890abcdef0"
	awsCli := &AwsCli{
		cfg:    cfg,
		client: mockClient,
	}
	mockClient.On("StartInstances", ctx, mock.MatchedBy(func(input *ec2.StartInstancesInput) bool {
		return len(input.InstanceIds) == 1 && input.InstanceIds[0] == instanceId
	}), mock.Anything).Return(&ec2.StartInstancesOutput{}, nil)

	err := awsCli.StartInstance(ctx, instanceId)
	require.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestStopInstance(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Region:   "us-west-2",
		SubnetID: "subnet-1234567890abcdef0",
		Credentials: config.Credentials{
			CredentialType: config.AWSCredentialTypeStatic,
			StaticCredentials: config.StaticCredentials{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
			},
		},
	}
	mockClient := new(MockComputeClient)
	instanceId := "i-1234567890abcdef0"
	awsCli := &AwsCli{
		cfg:    cfg,
		client: mockClient,
	}
	mockClient.On("StopInstances", ctx, mock.MatchedBy(func(input *ec2.StopInstancesInput) bool {
		return len(input.InstanceIds) == 1 && input.InstanceIds[0] == instanceId
	}), mock.Anything).Return(&ec2.StopInstancesOutput{}, nil)

	err := awsCli.StopInstance(ctx, instanceId)
	require.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestFindInstances(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Region:   "us-west-2",
		SubnetID: "subnet-1234567890abcdef0",
		Credentials: config.Credentials{
			CredentialType: config.AWSCredentialTypeStatic,
			StaticCredentials: config.StaticCredentials{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
			},
		},
	}
	mockClient := new(MockComputeClient)
	awsCli := &AwsCli{
		cfg:    cfg,
		client: mockClient,
	}
	instanceName := "instance-name"
	controllerID := "controllerID"
	tags := []types.Tag{
		{
			Key:   aws.String("tag:GARM_CONTROLLER_ID"),
			Value: &controllerID,
		},
		{
			Key:   aws.String("tag:Name"),
			Value: &instanceName,
		},
	}
	mockClient.On("DescribeInstances", ctx, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
		return len(input.Filters) == 3
	}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId: aws.String("i-1234567890abcdef0"),
						Tags:       tags,
					},
					{
						InstanceId: aws.String("i-1234567890abcdef1"),
						Tags:       tags,
					},
				},
			},
		},
	}, nil)

	instances, err := awsCli.FindInstances(ctx, controllerID, instanceName)
	require.NoError(t, err)
	require.Len(t, instances, 2)
	require.Equal(t, tags, instances[0].Tags)
	require.Equal(t, tags, instances[1].Tags)

	mockClient.AssertExpectations(t)
}

func TestFindOneInstanceWithName(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Region:   "us-west-2",
		SubnetID: "subnet-1234567890abcdef0",
		Credentials: config.Credentials{
			CredentialType: config.AWSCredentialTypeStatic,
			StaticCredentials: config.StaticCredentials{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
			},
		},
	}
	mockClient := new(MockComputeClient)
	awsCli := &AwsCli{
		cfg:    cfg,
		client: mockClient,
	}
	instanceName := "instance-name"
	controllerID := "controllerID"
	tags := []types.Tag{
		{
			Key:   aws.String("tag:GARM_CONTROLLER_ID"),
			Value: &controllerID,
		},
		{
			Key:   aws.String("tag:Name"),
			Value: &instanceName,
		},
	}
	mockClient.On("DescribeInstances", ctx, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
		return len(input.Filters) == 3
	}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId: aws.String("i-1234567890abcdef0"),
						Tags:       tags,
					},
				},
			},
		},
	}, nil)

	instance, err := awsCli.FindOneInstance(ctx, controllerID, instanceName)
	require.NoError(t, err)
	require.Equal(t, tags, instance.Tags)

	mockClient.AssertExpectations(t)
}

func TestFindOneInstanceWithID(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Region:   "us-west-2",
		SubnetID: "subnet-1234567890abcdef0",
		Credentials: config.Credentials{
			CredentialType: config.AWSCredentialTypeStatic,
			StaticCredentials: config.StaticCredentials{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
			},
		},
	}
	mockClient := new(MockComputeClient)
	awsCli := &AwsCli{
		cfg:    cfg,
		client: mockClient,
	}
	instanceId := "i-1234567890abcdef0"
	controllerID := "controllerID"
	mockClient.On("DescribeInstances", ctx, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
		return len(input.InstanceIds) == 1 && input.InstanceIds[0] == instanceId
	}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId: &instanceId,
					},
				},
			},
		},
	}, nil)

	instance, err := awsCli.FindOneInstance(ctx, controllerID, instanceId)
	require.NoError(t, err)
	require.Equal(t, instanceId, *instance.InstanceId)

	mockClient.AssertExpectations(t)
}

func TestGetInstance(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Region:   "us-west-2",
		SubnetID: "subnet-1234567890abcdef0",
		Credentials: config.Credentials{
			CredentialType: config.AWSCredentialTypeStatic,
			StaticCredentials: config.StaticCredentials{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
			},
		},
	}
	mockClient := new(MockComputeClient)
	awsCli := &AwsCli{
		cfg:    cfg,
		client: mockClient,
	}
	instanceID := "i-1234567890abcdef0"
	mockClient.On("DescribeInstances", ctx, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
		return len(input.InstanceIds) == 1 && input.InstanceIds[0] == instanceID
	}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId: aws.String(instanceID),
					},
				},
			},
		},
	}, nil)

	instance, err := awsCli.GetInstance(ctx, instanceID)
	require.NoError(t, err)
	require.Equal(t, instanceID, *instance.InstanceId)
}

func TestTerminateInstance(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Region:   "us-west-2",
		SubnetID: "subnet-1234567890abcdef0",
		Credentials: config.Credentials{
			CredentialType: config.AWSCredentialTypeStatic,
			StaticCredentials: config.StaticCredentials{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
			},
		},
	}
	mockClient := new(MockComputeClient)
	awsCli := &AwsCli{
		cfg:    cfg,
		client: mockClient,
	}
	poolID := "poolID"
	tags := []types.Tag{
		{
			Key:   aws.String("tag:GARM_POOL_ID"),
			Value: &poolID,
		},
	}
	mockClient.On("DescribeInstances", ctx, mock.MatchedBy(func(input *ec2.DescribeInstancesInput) bool {
		return len(input.Filters) == 2 && input.Filters[0].Name == aws.String("tag:GARM_POOL_ID")
	}), mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						Tags: tags,
					},
				},
			},
		},
	}, nil)

	mockClient.On("TerminateInstances", ctx, mock.MatchedBy(func(input *ec2.TerminateInstancesInput) bool {
		return len(input.InstanceIds) == 1
	}), mock.Anything).Return(&ec2.TerminateInstancesOutput{}, nil)

	err := awsCli.TerminateInstance(ctx, poolID)
	require.NoError(t, err)
}

func TestCreateRunningInstance(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Region:   "us-west-2",
		SubnetID: "subnet-1234567890abcdef0",
		Credentials: config.Credentials{
			CredentialType: config.AWSCredentialTypeStatic,
			StaticCredentials: config.StaticCredentials{
				AccessKeyID:     "AccessKeyID",
				SecretAccessKey: "SecretAccessKey",
				SessionToken:    "SessionToken",
			},
		},
	}
	mockClient := new(MockComputeClient)
	awsCli := &AwsCli{
		cfg:    cfg,
		client: mockClient,
	}
	instanceID := "i-1234567890abcdef0"
	spec.DefaultToolFetch = func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error) {
		return params.RunnerApplicationDownload{
			OS:           aws.String("linux"),
			Architecture: aws.String("amd64"),
			DownloadURL:  aws.String("MockURL"),
			Filename:     aws.String("garm-runner"),
		}, nil
	}
	spec := &spec.RunnerSpec{
		Region: "us-west-2",
		Tools: params.RunnerApplicationDownload{
			OS:           aws.String("linux"),
			Architecture: aws.String("amd64"),
			DownloadURL:  aws.String("MockURL"),
			Filename:     aws.String("garm-runner"),
		},
		BootstrapParams: params.BootstrapInstance{
			Name:   "instance-name",
			OSType: "linux",
			Image:  "ami-12345678",
			Flavor: "t2.micro",
			PoolID: "poolID",
		},
		SubnetID:     "subnet-1234567890abcdef0",
		SSHKeyName:   aws.String("SSHKeyName"),
		ControllerID: "controllerID",
	}
	mockClient.On("RunInstances", ctx, mock.Anything, mock.Anything).Return(&ec2.RunInstancesOutput{
		Instances: []types.Instance{
			{
				InstanceId: aws.String(instanceID),
				KeyName:    aws.String("SSHKeyName"),
			},
		},
	}, nil)

	instance, err := awsCli.CreateRunningInstance(ctx, spec)
	require.NoError(t, err)
	require.Equal(t, instanceID, instance)
}
