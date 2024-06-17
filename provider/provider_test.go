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

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cloudbase/garm-provider-aws/config"
	"github.com/cloudbase/garm-provider-aws/internal/client"
	"github.com/cloudbase/garm-provider-aws/internal/spec"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateInstance(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	spec.DefaultToolFetch = func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error) {
		return params.RunnerApplicationDownload{
			OS:           aws.String("linux"),
			Architecture: aws.String("amd64"),
			DownloadURL:  aws.String("MockURL"),
			Filename:     aws.String("garm-runner"),
		}, nil
	}
	bootstrapParams := params.BootstrapInstance{
		Name:   "garm-instance",
		Flavor: "t2.micro",
		Image:  "ami-12345678",
		Tools: []params.RunnerApplicationDownload{
			{
				OS:           aws.String("linux"),
				Architecture: aws.String("amd64"),
				DownloadURL:  aws.String("MockURL"),
				Filename:     aws.String("garm-runner"),
			},
		},
		OSType:     params.Linux,
		OSArch:     params.Amd64,
		PoolID:     "my-pool",
		ExtraSpecs: json.RawMessage(`{}`),
	}
	expectedInstance := params.ProviderInstance{
		ProviderID: instanceID,
		Name:       "garm-instance",
		OSType:     "linux",
		OSArch:     "amd64",
		Status:     "running",
	}
	provider := &AwsProvider{
		controllerID: "controllerID",
		awsCli:       &client.AwsCli{},
	}
	config := &config.Config{
		Region:   "us-east-1",
		SubnetID: "subnet-123456",
		Credentials: config.Credentials{
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretKey",
			SessionToken:    "token",
		},
	}
	mockComputeClient := new(client.MockComputeClient)
	provider.awsCli.SetConfig(config)
	provider.awsCli.SetClient(mockComputeClient)

	mockComputeClient.On("RunInstances", ctx, mock.Anything, mock.Anything).Return(&ec2.RunInstancesOutput{
		Instances: []types.Instance{
			{
				InstanceId: aws.String(instanceID),
			},
		},
	}, nil)
	result, err := provider.CreateInstance(ctx, bootstrapParams)
	assert.NoError(t, err)
	assert.Equal(t, expectedInstance, result)
}

func TestCreateInstanceError(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	spec.DefaultToolFetch = func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error) {
		return params.RunnerApplicationDownload{
			OS:           aws.String("linux"),
			Architecture: aws.String("amd64"),
			DownloadURL:  aws.String("MockURL"),
			Filename:     aws.String("garm-runner"),
		}, nil
	}
	bootstrapParams := params.BootstrapInstance{
		Name:   "garm-instance",
		Flavor: "t2.micro",
		Image:  "ami-12345678",
		Tools: []params.RunnerApplicationDownload{
			{
				OS:           aws.String("linux"),
				Architecture: aws.String("amd64"),
				DownloadURL:  aws.String("MockURL"),
				Filename:     aws.String("garm-runner"),
			},
		},
		OSType:     params.Linux,
		OSArch:     params.Amd64,
		PoolID:     "my-pool",
		ExtraSpecs: json.RawMessage(`{}`),
	}
	expectedInstance := params.ProviderInstance{}
	provider := &AwsProvider{
		controllerID: "controllerID",
		awsCli:       &client.AwsCli{},
	}
	config := &config.Config{
		Region:   "us-east-1",
		SubnetID: "subnet-123456",
		Credentials: config.Credentials{
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretKey",
			SessionToken:    "token",
		},
	}
	mockComputeClient := new(client.MockComputeClient)
	provider.awsCli.SetConfig(config)
	provider.awsCli.SetClient(mockComputeClient)

	mockComputeClient.On("RunInstances", ctx, mock.Anything, mock.Anything).Return(&ec2.RunInstancesOutput{
		Instances: []types.Instance{
			{
				InstanceId: aws.String(instanceID),
			},
		},
	}, fmt.Errorf("error creating instance"))
	result, err := provider.CreateInstance(ctx, bootstrapParams)
	assert.ErrorContains(t, err, "failed to create instance")
	assert.Equal(t, expectedInstance, result)
}

func TestDeleteInstanceWithID(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	provider := &AwsProvider{
		controllerID: "controllerID",
		awsCli:       &client.AwsCli{},
	}
	config := &config.Config{
		Region:   "us-east-1",
		SubnetID: "subnet-123456",
		Credentials: config.Credentials{
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretKey",
			SessionToken:    "token",
		},
	}
	mockComputeClient := new(client.MockComputeClient)
	provider.awsCli.SetConfig(config)
	provider.awsCli.SetClient(mockComputeClient)

	mockComputeClient.On("TerminateInstances", ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}, mock.Anything).Return(&ec2.TerminateInstancesOutput{}, nil)
	err := provider.DeleteInstance(ctx, instanceID)
	assert.NoError(t, err)
}

func TestDeleteInstanceWithName(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	instanceName := "garm-instance"
	provider := &AwsProvider{
		controllerID: "controllerID",
		awsCli:       &client.AwsCli{},
	}
	config := &config.Config{
		Region:   "us-east-1",
		SubnetID: "subnet-123456",
		Credentials: config.Credentials{
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretKey",
			SessionToken:    "token",
		},
	}
	mockComputeClient := new(client.MockComputeClient)
	provider.awsCli.SetConfig(config)
	provider.awsCli.SetClient(mockComputeClient)

	mockComputeClient.On("DescribeInstances", ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:GARM_CONTROLLER_ID"),
				Values: []string{""},
			},
			{
				Name:   aws.String("tag:Name"),
				Values: []string{instanceName},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running", "stopping", "stopped"},
			},
		},
	}, mock.Anything).Return(&ec2.DescribeInstancesOutput{
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
	mockComputeClient.On("TerminateInstances", ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}, mock.Anything).Return(&ec2.TerminateInstancesOutput{}, nil)
	err := provider.DeleteInstance(ctx, instanceName)
	assert.NoError(t, err)
}

func TestGetInstanceWithID(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	instanceName := "garm-instance"
	expectedOutput := params.ProviderInstance{
		ProviderID: instanceID,
		Name:       instanceName,
		OSType:     "linux",
		OSArch:     "amd64",
		Status:     "running",
	}
	provider := &AwsProvider{
		controllerID: "controllerID",
		awsCli:       &client.AwsCli{},
	}
	config := &config.Config{
		Region:   "us-east-1",
		SubnetID: "subnet-123456",
		Credentials: config.Credentials{
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretKey",
			SessionToken:    "token",
		},
	}
	mockComputeClient := new(client.MockComputeClient)
	provider.awsCli.SetConfig(config)
	provider.awsCli.SetClient(mockComputeClient)

	mockComputeClient.On("DescribeInstances", ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running", "stopping", "stopped"},
			},
		},
	}, mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId: aws.String(instanceID),
						Tags: []types.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String(instanceName),
							},
							{
								Key:   aws.String("OSType"),
								Value: aws.String("linux"),
							},
							{
								Key:   aws.String("OSArch"),
								Value: aws.String("amd64"),
							},
						},
						State: &types.InstanceState{
							Name: types.InstanceStateNameRunning,
						},
					},
				},
			},
		},
	}, nil)
	result, err := provider.GetInstance(ctx, instanceID)
	assert.NoError(t, err)
	assert.Equal(t, result, expectedOutput)
}

func TestGetInstanceWithName(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	instanceName := "garm-instance"
	expectedOutput := params.ProviderInstance{
		ProviderID: instanceID,
		Name:       instanceName,
		OSType:     "linux",
		OSArch:     "amd64",
		Status:     "running",
	}
	provider := &AwsProvider{
		controllerID: "controllerID",
		awsCli:       &client.AwsCli{},
	}
	config := &config.Config{
		Region:   "us-east-1",
		SubnetID: "subnet-123456",
		Credentials: config.Credentials{
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretKey",
			SessionToken:    "token",
		},
	}
	mockComputeClient := new(client.MockComputeClient)
	provider.awsCli.SetConfig(config)
	provider.awsCli.SetClient(mockComputeClient)

	mockComputeClient.On("DescribeInstances", ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:GARM_CONTROLLER_ID"),
				Values: []string{""},
			},
			{
				Name:   aws.String("tag:Name"),
				Values: []string{instanceName},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running", "stopping", "stopped"},
			},
		},
	}, mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId: aws.String(instanceID),
						Tags: []types.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String(instanceName),
							},
							{
								Key:   aws.String("OSType"),
								Value: aws.String("linux"),
							},
							{
								Key:   aws.String("OSArch"),
								Value: aws.String("amd64"),
							},
						},
						State: &types.InstanceState{
							Name: types.InstanceStateNameRunning,
						},
					},
				},
			},
		},
	}, nil)
	result, err := provider.GetInstance(ctx, instanceName)
	assert.NoError(t, err)
	assert.Equal(t, result, expectedOutput)
}

func TestListInstances(t *testing.T) {
	ctx := context.Background()
	poolID := "my-pool"
	expectedOutput := []params.ProviderInstance{
		{
			ProviderID: "i-1234567890abcdef0",
			Name:       "garm-instance",
			OSType:     "linux",
			OSArch:     "amd64",
			Status:     "running",
		},
		{
			ProviderID: "i-1234567890abcdef1",
			Name:       "garm-instance1",
			OSType:     "linux",
			OSArch:     "amd64",
			Status:     "running",
		},
	}
	provider := &AwsProvider{
		controllerID: "controllerID",
		awsCli:       &client.AwsCli{},
	}
	config := &config.Config{
		Region:   "us-east-1",
		SubnetID: "subnet-123456",
		Credentials: config.Credentials{
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretKey",
			SessionToken:    "token",
		},
	}
	mockComputeClient := new(client.MockComputeClient)
	provider.awsCli.SetConfig(config)
	provider.awsCli.SetClient(mockComputeClient)

	mockComputeClient.On("DescribeInstances", ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:GARM_POOL_ID"),
				Values: []string{poolID},
			},
			{
				//   - instance-state-name - The state of the instance ( pending | running |
				//   shutting-down | terminated | stopping | stopped ).
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running", "stopping", "stopped"},
			},
		},
	}, mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId: aws.String("i-1234567890abcdef0"),
						Tags: []types.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String("garm-instance"),
							},
							{
								Key:   aws.String("OSType"),
								Value: aws.String("linux"),
							},
							{
								Key:   aws.String("OSArch"),
								Value: aws.String("amd64"),
							},
						},
						State: &types.InstanceState{
							Name: types.InstanceStateNameRunning,
						},
					},
					{
						InstanceId: aws.String("i-1234567890abcdef1"),
						Tags: []types.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String("garm-instance1"),
							},
							{
								Key:   aws.String("OSType"),
								Value: aws.String("linux"),
							},
							{
								Key:   aws.String("OSArch"),
								Value: aws.String("amd64"),
							},
						},
						State: &types.InstanceState{
							Name: types.InstanceStateNameRunning,
						},
					},
				},
			},
		},
	}, nil)
	result, err := provider.ListInstances(ctx, poolID)
	assert.NoError(t, err)
	assert.Equal(t, result, expectedOutput)
}

func TestStop(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	provider := &AwsProvider{
		controllerID: "controllerID",
		awsCli:       &client.AwsCli{},
	}
	config := &config.Config{
		Region:   "us-east-1",
		SubnetID: "subnet-123456",
		Credentials: config.Credentials{
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretKey",
			SessionToken:    "token",
		},
	}
	mockComputeClient := new(client.MockComputeClient)
	provider.awsCli.SetConfig(config)
	provider.awsCli.SetClient(mockComputeClient)

	mockComputeClient.On("StopInstances", ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}, mock.Anything).Return(&ec2.StopInstancesOutput{}, nil)
	err := provider.Stop(ctx, instanceID, false)
	assert.NoError(t, err)
}

func TestStartStoppedInstance(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	provider := &AwsProvider{
		controllerID: "controllerID",
		awsCli:       &client.AwsCli{},
	}
	config := &config.Config{
		Region:   "us-east-1",
		SubnetID: "subnet-123456",
		Credentials: config.Credentials{
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretKey",
			SessionToken:    "token",
		},
	}
	mockComputeClient := new(client.MockComputeClient)
	provider.awsCli.SetConfig(config)
	provider.awsCli.SetClient(mockComputeClient)

	mockComputeClient.On("DescribeInstances", ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running", "stopping", "stopped"},
			},
		},
	}, mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId: aws.String(instanceID),
						State: &types.InstanceState{
							Name: types.InstanceStateNameStopped,
						},
					},
				},
			},
		},
	}, nil)
	mockComputeClient.On("StartInstances", ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	}, mock.Anything).Return(&ec2.StartInstancesOutput{}, nil)
	err := provider.Start(ctx, instanceID)
	assert.NoError(t, err)
}

func TestStartStoppingInstance(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	provider := &AwsProvider{
		controllerID: "controllerID",
		awsCli:       &client.AwsCli{},
	}
	config := &config.Config{
		Region:   "us-east-1",
		SubnetID: "subnet-123456",
		Credentials: config.Credentials{
			AccessKeyID:     "accessKey",
			SecretAccessKey: "secretKey",
			SessionToken:    "token",
		},
	}
	mockComputeClient := new(client.MockComputeClient)
	provider.awsCli.SetConfig(config)
	provider.awsCli.SetClient(mockComputeClient)

	mockComputeClient.On("DescribeInstances", ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running", "stopping", "stopped"},
			},
		},
	}, mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId: aws.String(instanceID),
						State: &types.InstanceState{
							Name: types.InstanceStateNameStopping,
						},
					},
				},
			},
		},
	}, nil)
	mockComputeClient.On("StartInstances", ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	}, mock.Anything).Return(&ec2.StartInstancesOutput{}, nil)
	err := provider.Start(ctx, instanceID)
	assert.Error(t, err)
	assert.Equal(t, "instance "+instanceID+" cannot be started in stopping state", err.Error())
}
