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
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/stretchr/testify/require"
)

func TestAwsInstanceToParamsInstance(t *testing.T) {
	tests := []struct {
		name        string
		ec2Instance types.Instance
		want        params.ProviderInstance
		errString   string
	}{
		{
			name: "valid instance",
			ec2Instance: types.Instance{
				InstanceId: aws.String("instance_id"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("name"),
					},
					{
						Key:   aws.String("OSType"),
						Value: aws.String("os_type"),
					},
					{
						Key:   aws.String("OSArch"),
						Value: aws.String("os_arch"),
					},
				},
				State: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
			},
			want: params.ProviderInstance{
				ProviderID: "instance_id",
				Name:       "name",
				OSType:     params.OSType("os_type"),
				OSArch:     params.OSArch("os_arch"),
				Status:     params.InstanceRunning,
			},
			errString: "",
		},
		{
			name: "missing instance ID",
			ec2Instance: types.Instance{
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("name"),
					},
					{
						Key:   aws.String("OSType"),
						Value: aws.String("os_type"),
					},
					{
						Key:   aws.String("OSArch"),
						Value: aws.String("os_arch"),
					},
				},
				State: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
			},
			want:      params.ProviderInstance{},
			errString: "instance ID is nil",
		},
		{
			name: "missing tags",
			ec2Instance: types.Instance{
				InstanceId: aws.String("instance_id"),
				State: &types.InstanceState{
					Name: types.InstanceStateNameRunning,
				},
			},
			want: params.ProviderInstance{
				ProviderID: "instance_id",
				Status:     params.InstanceRunning,
			},
			errString: "",
		},
		{
			name: "terminated status",
			ec2Instance: types.Instance{
				InstanceId: aws.String("instance_id"),
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("name"),
					},
					{
						Key:   aws.String("OSType"),
						Value: aws.String("os_type"),
					},
					{
						Key:   aws.String("OSArch"),
						Value: aws.String("os_arch"),
					},
				},
				State: &types.InstanceState{
					Name: types.InstanceStateNameTerminated,
				},
			},
			want: params.ProviderInstance{
				ProviderID: "instance_id",
				Name:       "name",
				OSType:     params.OSType("os_type"),
				OSArch:     params.OSArch("os_arch"),
				Status:     params.InstanceStopped,
			},
			errString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AwsInstanceToParamsInstance(tt.ec2Instance)
			require.Equal(t, tt.want, result)
			if tt.errString == "" {
				require.Nil(t, err)
			} else {
				require.EqualError(t, err, tt.errString)
			}
		})
	}
}

func TestIsEC2NotFoundErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "not found error",
			err: &smithy.GenericAPIError{
				Code: "InvalidInstanceID.NotFound",
			},
			want: true,
		},
		{
			name: "not found error",
			err: &smithy.GenericAPIError{
				Code: "OtherError",
			},
			want: false,
		},
		{
			name: "not found error",
			err:  errors.New("other error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEC2NotFoundErr(tt.err)
			require.Equal(t, tt.want, result)
		})
	}
}
