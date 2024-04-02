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

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/stretchr/testify/mock"
)

type MockComputeClient struct {
	mock.Mock
}

func (m *MockComputeClient) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ec2.StartInstancesOutput), args.Error(1)
}

func (m *MockComputeClient) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ec2.StopInstancesOutput), args.Error(1)
}

func (m *MockComputeClient) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ec2.DescribeInstancesOutput), args.Error(1)
}

func (m *MockComputeClient) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput, optFns ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ec2.TerminateInstancesOutput), args.Error(1)
}

func (m *MockComputeClient) RunInstances(ctx context.Context, params *ec2.RunInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ec2.RunInstancesOutput), args.Error(1)
}
