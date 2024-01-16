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

package util

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/cloudbase/garm-provider-common/params"
)

func AwsInstanceToParamsInstance(ec2Instance types.Instance) (params.ProviderInstance, error) {
	if ec2Instance.InstanceId == nil {
		return params.ProviderInstance{}, fmt.Errorf("instance ID is nil")
	}
	details := params.ProviderInstance{
		ProviderID: *ec2Instance.InstanceId,
	}

	for _, tag := range ec2Instance.Tags {
		if tag.Key == nil || tag.Value == nil {
			continue
		}
		switch *tag.Key {
		case "Name":
			details.Name = *tag.Value
		case "OSType":
			details.OSType = params.OSType(*tag.Value)
		case "OSArch":
			details.OSArch = params.OSArch(*tag.Value)
		}
	}

	switch ec2Instance.State.Name {
	case types.InstanceStateNameRunning,
		types.InstanceStateNameShuttingDown,
		types.InstanceStateNameStopping:

		details.Status = params.InstanceRunning
	case types.InstanceStateNameStopped,
		types.InstanceStateNameTerminated:

		details.Status = params.InstanceStopped
	default:
		details.Status = params.InstanceStatusUnknown
	}
	return details, nil
}

func IsEC2NotFoundErr(err error) bool {
	var apiErr smithy.APIError
	ok := errors.As(err, &apiErr)

	if ok && apiErr.ErrorCode() == "InvalidInstanceID.NotFound" {
		return true
	}
	return false
}
