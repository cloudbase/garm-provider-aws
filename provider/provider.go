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
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cloudbase/garm-provider-aws/config"
	"github.com/cloudbase/garm-provider-aws/internal/client"
	"github.com/cloudbase/garm-provider-aws/internal/spec"
	"github.com/cloudbase/garm-provider-aws/internal/util"
	garmErrors "github.com/cloudbase/garm-provider-common/errors"
	execution "github.com/cloudbase/garm-provider-common/execution/v0.1.0"
	"github.com/cloudbase/garm-provider-common/params"
)

var _ execution.ExternalProvider = &AwsProvider{}

var Version = "v0.0.0-unknown"

func NewAwsProvider(ctx context.Context, configPath, controllerID string) (execution.ExternalProvider, error) {
	conf, err := config.NewConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	awsCli, err := client.NewAwsCli(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS CLI: %w", err)
	}

	return &AwsProvider{
		controllerID: controllerID,
		awsCli:       awsCli,
	}, nil
}

type AwsProvider struct {
	controllerID string
	awsCli       *client.AwsCli
}

func (a *AwsProvider) CreateInstance(ctx context.Context, bootstrapParams params.BootstrapInstance) (params.ProviderInstance, error) {
	spec, err := spec.GetRunnerSpecFromBootstrapParams(a.awsCli.Config(), bootstrapParams, a.controllerID)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("failed to get runner spec: %w", err)
	}

	instanceID, err := a.awsCli.CreateRunningInstance(ctx, spec)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("failed to create instance: %w", err)
	}

	instance := params.ProviderInstance{
		ProviderID: instanceID,
		Name:       spec.BootstrapParams.Name,
		OSType:     spec.BootstrapParams.OSType,
		OSArch:     spec.BootstrapParams.OSArch,
		Status:     "running",
	}

	return instance, nil

}

func (a *AwsProvider) DeleteInstance(ctx context.Context, instance string) error {
	var inst string
	if strings.HasPrefix(instance, "i-") {
		inst = instance
	} else {
		tmp, err := a.awsCli.FindOneInstance(ctx, "", instance)
		if err != nil {
			if errors.Is(err, garmErrors.ErrNotFound) {
				return nil
			}
			return fmt.Errorf("failed to determine instance: %w", err)
		}

		if tmp.InstanceId == nil {
			return fmt.Errorf("failed to determine instance: %w", err)
		}
		inst = *tmp.InstanceId
	}

	if inst == "" {
		return nil
	}

	if err := a.awsCli.TerminateInstance(ctx, inst); err != nil {
		return fmt.Errorf("failed to terminate instance: %w", err)
	}

	return nil
}

func (a *AwsProvider) GetInstance(ctx context.Context, instance string) (params.ProviderInstance, error) {
	awsInstance, err := a.awsCli.FindOneInstance(ctx, "", instance)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("failed to get VM details: %w", err)
	}
	if awsInstance.InstanceId == nil {
		return params.ProviderInstance{}, nil
	}

	providerInstance, err := util.AwsInstanceToParamsInstance(awsInstance)
	if err != nil {
		return params.ProviderInstance{}, fmt.Errorf("failed to convert instance: %w", err)
	}
	return providerInstance, nil
}

func (a *AwsProvider) ListInstances(ctx context.Context, poolID string) ([]params.ProviderInstance, error) {
	awsInstances, err := a.awsCli.ListDescribedInstances(ctx, poolID)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	var providerInstances []params.ProviderInstance
	for _, val := range awsInstances {
		inst, err := util.AwsInstanceToParamsInstance(val)
		if err != nil {
			return []params.ProviderInstance{}, fmt.Errorf("failed to convert instance: %w", err)
		}
		providerInstances = append(providerInstances, inst)
	}

	return providerInstances, nil
}

func (a *AwsProvider) RemoveAllInstances(ctx context.Context) error {
	return nil
}

func (a *AwsProvider) Stop(ctx context.Context, instance string, force bool) error {
	return a.awsCli.StopInstance(ctx, instance)
}

func (a *AwsProvider) Start(ctx context.Context, instance string) error {
	awsInstance, err := a.awsCli.FindOneInstance(ctx, "", instance)
	if err != nil {
		return fmt.Errorf("failed to determine instance: %w", err)
	}
	if awsInstance.State.Name == types.InstanceStateNameStopping {
		return fmt.Errorf("instance %s cannot be started in %s state", instance, awsInstance.State.Name)
	}
	return a.awsCli.StartInstance(ctx, instance)
}

func (a *AwsProvider) GetVersion(ctx context.Context) string {
	return Version
}
