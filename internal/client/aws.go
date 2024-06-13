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
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cloudbase/garm-provider-aws/config"
	"github.com/cloudbase/garm-provider-aws/internal/spec"
	"github.com/cloudbase/garm-provider-aws/internal/util"

	"github.com/cloudbase/garm-provider-common/errors"
)

func NewAwsCli(ctx context.Context, cfg *config.Config) (*AwsCli, error) {
	cliCfg, err := cfg.GetAWSConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS cli context: %w", err)
	}

	client := ec2.NewFromConfig(cliCfg)
	awsCli := &AwsCli{
		cfg:    cfg,
		client: client,
	}

	return awsCli, nil
}

type ClientInterface interface {
	StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
}

type AwsCli struct {
	cfg *config.Config

	client ClientInterface
}

func (a *AwsCli) Config() *config.Config {
	return a.cfg
}

func (a *AwsCli) Client() ClientInterface {
	return a.client
}

func (a *AwsCli) SetConfig(cfg *config.Config) {
	a.cfg = cfg
}

func (a *AwsCli) SetClient(client ClientInterface) {
	a.client = client
}

func (a *AwsCli) StartInstance(ctx context.Context, vmName string) error {
	_, err := a.client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{vmName},
	})
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	return nil
}

func (a *AwsCli) StopInstance(ctx context.Context, vmName string) error {
	_, err := a.client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{vmName},
	})
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	return nil
}

func (a *AwsCli) FindInstances(ctx context.Context, controllerID, instanceName string) ([]types.Instance, error) {
	resp, err := a.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:GARM_CONTROLLER_ID"),
				Values: []string{controllerID},
			},
			{
				Name:   aws.String("tag:Name"),
				Values: []string{instanceName},
			},
			{
				//   - instance-state-name - The state of the instance ( pending | running |
				//   shutting-down | terminated | stopping | stopped ).
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running", "stopping", "stopped"},
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find instances by tags: %w", err)
	}

	var instances []types.Instance
	for _, reserv := range resp.Reservations {
		instances = append(instances, reserv.Instances...)
	}

	return instances, nil
}

func (a *AwsCli) FindOneInstance(ctx context.Context, controllerID, instanceName string) (types.Instance, error) {
	if strings.HasPrefix(instanceName, "i-") {
		resp, err := a.GetInstance(ctx, instanceName)
		if err != nil {
			return types.Instance{}, fmt.Errorf("failed to get instance %s: %w", instanceName, err)
		}
		return resp, nil
	}
	resp, err := a.FindInstances(ctx, controllerID, instanceName)
	if err != nil {
		return types.Instance{}, fmt.Errorf("failed to find instance %s: %w", instanceName, errors.ErrNotFound)
	}

	if len(resp) > 1 {
		return types.Instance{}, fmt.Errorf("found more than one instance with name %s", instanceName)
	}

	if len(resp) == 0 {
		return types.Instance{}, fmt.Errorf("no such instance %s: %w", instanceName, errors.ErrNotFound)
	}

	return resp[0], nil

}

// Describes the specified instances or all instances. If you specify instance
// IDs, the output includes information for only the specified instances. If you
// specify filters, the output includes information for only those instances that
// meet the filter criteria.
func (a *AwsCli) GetInstance(ctx context.Context, instance string) (types.Instance, error) {
	resp, err := a.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instance},
		Filters: []types.Filter{
			{
				//   - instance-state-name - The state of the instance ( pending | running |
				//   shutting-down | terminated | stopping | stopped ).
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running", "stopping", "stopped"},
			},
		},
	})
	if err != nil {
		return types.Instance{}, fmt.Errorf("failed to get instance: %w", err)
	}

	var instances []types.Instance
	for _, reserv := range resp.Reservations {
		instances = append(instances, reserv.Instances...)
	}

	if len(instances) == 0 {
		return types.Instance{}, fmt.Errorf("no such instance %s: %w", instance, errors.ErrNotFound)
	}

	return instances[0], nil
}

// You can stop, start, and terminate EBS-backed instances. You can only terminate instance store-backed instances.
// What happens to an instance differs if you stop it or terminate it. For example, when you stop an instance,
// the root device and any other devices attached to the instance persist. When you terminate an instance,
// any attached EBS volumes with the DeleteOnTermination block device mapping parameter set to true are
// automatically deleted.
func (a *AwsCli) TerminateInstance(ctx context.Context, vmName string) error {
	_, err := a.client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{vmName},
	})
	if err != nil {
		if util.IsEC2NotFoundErr(err) {
			return nil
		}
		return fmt.Errorf("failed to terminate instance: %w", err)
	}

	return nil
}

func (a *AwsCli) ListDescribedInstances(ctx context.Context, poolID string) ([]types.Instance, error) {
	resp, err := a.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
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
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	var instances []types.Instance
	for _, reserv := range resp.Reservations {
		instances = append(instances, reserv.Instances...)
	}

	return instances, nil
}

func (a *AwsCli) CreateRunningInstance(ctx context.Context, spec *spec.RunnerSpec) (string, error) {

	if spec == nil {
		return "", fmt.Errorf("invalid nil runner spec")
	}

	udata, err := spec.ComposeUserData()
	if err != nil {
		return "", fmt.Errorf("failed to compose user data: %w", err)
	}

	resp, err := a.client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String(spec.BootstrapParams.Image),
		InstanceType: types.InstanceType(spec.BootstrapParams.Flavor),
		MaxCount:     aws.Int32(1),
		MinCount:     aws.Int32(1),
		SubnetId:     aws.String(spec.SubnetID),
		UserData:     aws.String(udata),
		KeyName:      spec.SSHKeyName,
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(spec.BootstrapParams.Name),
					},
					{
						Key:   aws.String("GARM_POOL_ID"),
						Value: aws.String(spec.BootstrapParams.PoolID),
					},
					{
						Key:   aws.String("OSType"),
						Value: aws.String(string(spec.BootstrapParams.OSType)),
					},
					{
						Key:   aws.String("OSArch"),
						Value: aws.String(string(spec.BootstrapParams.OSArch)),
					},
					{
						Key:   aws.String("GARM_CONTROLLER_ID"),
						Value: aws.String(spec.ControllerID),
					},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create instance: %w", err)
	}

	return *resp.Instances[0].InstanceId, nil
}
