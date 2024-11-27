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
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cloudbase/garm-provider-aws/config"
	"github.com/cloudbase/garm-provider-common/cloudconfig"
	"github.com/cloudbase/garm-provider-common/params"
	"github.com/cloudbase/garm-provider-common/util"
	"github.com/invopop/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

type ToolFetchFunc func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error)

var DefaultToolFetch ToolFetchFunc = util.GetTools

func generateJSONSchema() *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
	}
	// Reflect the extraSpecs struct
	schema := reflector.Reflect(extraSpecs{})

	return schema
}

func jsonSchemaValidation(schema json.RawMessage) error {
	jsonSchema := generateJSONSchema()
	schemaLoader := gojsonschema.NewGoLoader(jsonSchema)
	extraSpecsLoader := gojsonschema.NewBytesLoader(schema)
	result, err := gojsonschema.Validate(schemaLoader, extraSpecsLoader)
	if err != nil {
		return fmt.Errorf("failed to validate schema: %w", err)
	}
	if !result.Valid() {
		return fmt.Errorf("schema validation failed: %s", result.Errors())
	}
	return nil
}

func newExtraSpecsFromBootstrapData(data params.BootstrapInstance) (*extraSpecs, error) {
	spec := &extraSpecs{}

	if err := jsonSchemaValidation(data.ExtraSpecs); err != nil {
		return nil, fmt.Errorf("failed to validate extra specs: %w", err)
	}

	if len(data.ExtraSpecs) > 0 {
		if err := json.Unmarshal(data.ExtraSpecs, spec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal extra specs: %w", err)
		}
	}

	return spec, nil
}

type extraSpecs struct {
	SubnetID         *string          `json:"subnet_id,omitempty" jsonschema:"pattern=^subnet-[0-9a-fA-F]{17}$,description=The ID of the subnet formatted as subnet-xxxxxxxxxxxxxxxxx."`
	SSHKeyName       *string          `json:"ssh_key_name,omitempty" jsonschema:"description=The name of the Key Pair to use for the instance."`
	Iops             *int32           `json:"iops,omitempty" jsonschema:"description=Specifies the number of IOPS (Input/Output Operations Per Second) provisioned for the volume. Required for io1 and io2 volumes. Optional for gp3 volumes."`
	Throughput       *int32           `json:"throughput,omitempty" jsonschema:"description=Specifies the throughput (MiB/s) provisioned for the volume. Valid only for gp3 volumes.,minimum=125,maximum=1000"`
	VolumeSize       *int32           `json:"volume_size,omitempty" jsonschema:"description=Specifies the size of the volume in GiB. Required unless a snapshot ID is provided."`
	VolumeType       types.VolumeType `json:"volume_type,omitempty" jsonschema:"enum=gp2,enum=gp3,enum=io1,enum=io2,enum=st1,enum=sc1,enum=standard,description=Specifies the EBS volume type."`
	SecurityGroupIds []string         `json:"security_group_ids,omitempty" jsonschema:"description=The security group IDs to associate with the instance. Default: Amazon EC2 uses the default security group."`
	DisableUpdates   *bool            `json:"disable_updates,omitempty" jsonschema:"description=Disable automatic updates on the VM."`
	EnableBootDebug  *bool            `json:"enable_boot_debug,omitempty" jsonschema:"description=Enable boot debug on the VM."`
	ExtraPackages    []string         `json:"extra_packages,omitempty" jsonschema:"description=Extra packages to install on the VM."`
	cloudconfig.CloudConfigSpec
}

func GetRunnerSpecFromBootstrapParams(cfg *config.Config, data params.BootstrapInstance, controllerID string) (*RunnerSpec, error) {
	tools, err := DefaultToolFetch(data.OSType, data.OSArch, data.Tools)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools: %s", err)
	}

	extraSpecs, err := newExtraSpecsFromBootstrapData(data)
	if err != nil {
		return nil, fmt.Errorf("error loading extra specs: %w", err)
	}

	spec := &RunnerSpec{
		Region:          cfg.Region,
		ExtraPackages:   extraSpecs.ExtraPackages,
		Tools:           tools,
		BootstrapParams: data,
		SubnetID:        cfg.SubnetID,
		ControllerID:    controllerID,
	}

	spec.MergeExtraSpecs(extraSpecs)

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("error validating spec: %w", err)
	}

	return spec, nil
}

type RunnerSpec struct {
	Region           string
	DisableUpdates   bool
	ExtraPackages    []string
	EnableBootDebug  bool
	Tools            params.RunnerApplicationDownload
	BootstrapParams  params.BootstrapInstance
	SecurityGroupIds []string
	SubnetID         string
	SSHKeyName       *string
	Iops             *int32
	Throughput       *int32
	VolumeSize       *int32
	VolumeType       types.VolumeType
	ControllerID     string
}

func (r *RunnerSpec) Validate() error {
	if r.Region == "" {
		return fmt.Errorf("missing region")
	}
	if r.BootstrapParams.Name == "" {
		return fmt.Errorf("missing bootstrap params")
	}
	if r.Iops != nil {
		switch r.VolumeType {
		case types.VolumeTypeIo1:
			if *r.Iops < 100 || *r.Iops > 64000 {
				return fmt.Errorf("EBS iops for volume type %s must be between 100 and 64000", r.VolumeType)
			}
		case types.VolumeTypeIo2:
			if *r.Iops < 100 || *r.Iops > 256000 {
				return fmt.Errorf("EBS iops for volume type %s must be between 100 and 256000", r.VolumeType)
			}
		case types.VolumeTypeGp3:
			if *r.Iops < 3000 || *r.Iops > 16000 {
				return fmt.Errorf("EBS iops for volume type %s must be between 3000 and 16000", r.VolumeType)
			}
		default:
			return fmt.Errorf("EBS iops is only valid for volume types io1, io2 and gp3")
		}
	}
	if r.Throughput != nil && r.VolumeType != types.VolumeTypeGp3 {
		return fmt.Errorf("EBS throughput is only valid for volume type gp3")
	}
	if r.VolumeSize != nil {
		switch r.VolumeType {
		case types.VolumeTypeIo1:
			if *r.VolumeSize < 4 || *r.VolumeSize > 16384 {
				return fmt.Errorf("EBS volume size for volume type %s must be between 4 and 16384", r.VolumeType)
			}
		case types.VolumeTypeIo2:
			if *r.VolumeSize < 4 || *r.VolumeSize > 16384 {
				return fmt.Errorf("EBS volume size for volume type %s must be between 4 and 16384", r.VolumeType)
			}
		case types.VolumeTypeGp2, types.VolumeTypeGp3:
			if *r.VolumeSize < 1 || *r.VolumeSize > 16384 {
				return fmt.Errorf("EBS volume size for volume type %s must be between 1 and 16384", r.VolumeType)
			}
		case types.VolumeTypeSt1, types.VolumeTypeSc1:
			if *r.VolumeSize < 125 || *r.VolumeSize > 16384 {
				return fmt.Errorf("EBS volume size for volume type %s must be between 125 and 16384", r.VolumeType)
			}
		case types.VolumeTypeStandard, "":
			if *r.VolumeSize < 1 || *r.VolumeSize > 1024 {
				return fmt.Errorf("EBS volume size for volume type standard must be between 1 and 1024")
			}
		default:
			return fmt.Errorf("EBS volume size is only valid for volume types io1, io2, gp2, gp3, st1, sc1 and standard")
		}
	}

	if r.VolumeType != "" {
		switch r.VolumeType {
		case types.VolumeTypeIo1, types.VolumeTypeIo2:
			if r.Iops == nil {
				return fmt.Errorf("the parameter iops must be specified for %s volumes", r.VolumeType)
			}
		}
	}
	return nil
}

func (r *RunnerSpec) MergeExtraSpecs(extraSpecs *extraSpecs) {
	if extraSpecs.SubnetID != nil && *extraSpecs.SubnetID != "" {
		r.SubnetID = *extraSpecs.SubnetID
	}

	if extraSpecs.Iops != nil {
		r.Iops = extraSpecs.Iops
	}

	if extraSpecs.Throughput != nil {
		r.Throughput = extraSpecs.Throughput
	}

	if extraSpecs.VolumeSize != nil {
		r.VolumeSize = extraSpecs.VolumeSize
	}

	if extraSpecs.VolumeType != "" {
		r.VolumeType = extraSpecs.VolumeType
	}

	if extraSpecs.SSHKeyName != nil {
		r.SSHKeyName = extraSpecs.SSHKeyName
	}

	if len(extraSpecs.SecurityGroupIds) > 0 {
		r.SecurityGroupIds = extraSpecs.SecurityGroupIds
	}

	if extraSpecs.DisableUpdates != nil {
		r.DisableUpdates = *extraSpecs.DisableUpdates
	}

	if extraSpecs.EnableBootDebug != nil {
		r.EnableBootDebug = *extraSpecs.EnableBootDebug
	}
}

func (r *RunnerSpec) ComposeUserData() (string, error) {
	bootstrapParams := r.BootstrapParams
	bootstrapParams.UserDataOptions.DisableUpdatesOnBoot = r.DisableUpdates
	bootstrapParams.UserDataOptions.ExtraPackages = r.ExtraPackages
	bootstrapParams.UserDataOptions.EnableBootDebug = r.EnableBootDebug
	switch bootstrapParams.OSType {
	case params.Linux:
		udata, err := cloudconfig.GetCloudConfig(bootstrapParams, r.Tools, bootstrapParams.Name)
		if err != nil {
			return "", fmt.Errorf("failed to generate userdata: %w", err)
		}
		asBase64 := base64.StdEncoding.EncodeToString([]byte(udata))
		return asBase64, nil
	case params.Windows:
		udata, err := cloudconfig.GetCloudConfig(bootstrapParams, r.Tools, bootstrapParams.Name)
		if err != nil {
			return "", fmt.Errorf("failed to generate userdata: %w", err)
		}
		wrapped := fmt.Sprintf("<powershell>%s</powershell>", udata)
		asBase64 := base64.StdEncoding.EncodeToString([]byte(wrapped))
		return asBase64, nil
	}
	return "", fmt.Errorf("unsupported OS type for cloud config: %s", bootstrapParams.OSType)
}
