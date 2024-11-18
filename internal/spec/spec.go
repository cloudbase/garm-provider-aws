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
	SubnetID         *string  `json:"subnet_id,omitempty" jsonschema:"pattern=^subnet-[0-9a-fA-F]{17}$"`
	SSHKeyName       *string  `json:"ssh_key_name,omitempty" jsonschema:"description=The name of the Key Pair to use for the instance."`
	SecurityGroupIds []string `json:"security_group_ids,omitempty" jsonschema:"description=The security groups IDs to associate with the instance. Default: Amazon EC2 uses the default security group."`
	DisableUpdates   *bool    `json:"disable_updates,omitempty" jsonschema:"description=Disable automatic updates on the VM."`
	EnableBootDebug  *bool    `json:"enable_boot_debug,omitempty" jsonschema:"description=Enable boot debug on the VM"`
	ExtraPackages    []string `json:"extra_packages,omitempty" jsonschema:"description=Extra packages to install on the VM"`
	// The Cloudconfig struct from common package
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
	ControllerID     string
}

func (r *RunnerSpec) Validate() error {
	if r.Region == "" {
		return fmt.Errorf("missing region")
	}
	if r.BootstrapParams.Name == "" {
		return fmt.Errorf("missing bootstrap params")
	}
	return nil
}

func (r *RunnerSpec) MergeExtraSpecs(extraSpecs *extraSpecs) {
	if extraSpecs.SubnetID != nil && *extraSpecs.SubnetID != "" {
		r.SubnetID = *extraSpecs.SubnetID
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
