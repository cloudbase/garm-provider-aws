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
	"github.com/xeipuuv/gojsonschema"
)

const (
	jsonSchema string = `
		{
			"$schema": "http://cloudbase.it/garm-provider-aws/schemas/extra_specs#",
			"type": "object",
			"description": "Schema defining supported extra specs for the Garm AWS Provider",
			"properties": {
				"subnet_id": {
					"type": "string",
					"pattern": "^subnet-[0-9a-fA-F]{17}$"
				},
				"disable_updates": {
					"type": "boolean",
					"description": "Disable automatic updates on the VM."
				},
				"enable_boot_debug": {
					"type": "boolean",
					"description": "Enable boot debug on the VM."
				},
				"extra_packages": {
					"type": "array",
					"description": "Extra packages to install on the VM.",
					"items": {
						"type": "string"
					}
				},
				"runner_install_template": {
					"type": "string",
					"description": "This option can be used to override the default runner install template. If used, the caller is responsible for the correctness of the template as well as the suitability of the template for the target OS. Use the extra_context extra spec if your template has variables in it that need to be expanded."
				},
				"extra_context": {
					"type": "object",
					"description": "Extra context that will be passed to the runner_install_template.",
					"additionalProperties": {
						"type": "string"
					}
				},
				"pre_install_scripts": {
					"type": "object",
					"description": "A map of pre-install scripts that will be run before the runner install script. These will run as root and can be used to prep a generic image before we attempt to install the runner. The key of the map is the name of the script as it will be written to disk. The value is a byte array with the contents of the script."
				}
			},
			"additionalProperties": false
		}
	`
)

type ToolFetchFunc func(osType params.OSType, osArch params.OSArch, tools []params.RunnerApplicationDownload) (params.RunnerApplicationDownload, error)

var DefaultToolFetch ToolFetchFunc = util.GetTools

func jsonSchemaValidation(schema json.RawMessage) error {
	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)
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
	SubnetID        *string  `json:"subnet_id,omitempty"`
	DisableUpdates  *bool    `json:"disable_updates"`
	EnableBootDebug *bool    `json:"enable_boot_debug"`
	ExtraPackages   []string `json:"extra_packages"`
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

	return spec, nil
}

type RunnerSpec struct {
	Region          string
	DisableUpdates  bool
	ExtraPackages   []string
	EnableBootDebug bool
	Tools           params.RunnerApplicationDownload
	BootstrapParams params.BootstrapInstance
	SubnetID        string
	ControllerID    string
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
