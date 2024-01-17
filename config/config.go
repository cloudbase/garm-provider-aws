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

package config

import (
	"context"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// NewConfig returns a new Config
func NewConfig(cfgFile string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(cfgFile, &config); err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}
	return &config, nil
}

type Config struct {
	Credentials Credentials `toml:"credentials"`
	SubnetID    string      `toml:"subnet_id"`
	Region      string      `toml:"region"`
}

func (c *Config) Validate() error {
	if err := c.Credentials.Validate(); err != nil {
		return fmt.Errorf("failed to validate credentials: %w", err)
	}

	if c.SubnetID == "" {
		return fmt.Errorf("missing subnet_id")
	}

	if c.Region == "" {
		return fmt.Errorf("missing region")
	}
	return nil
}

type Credentials struct {
	// AWS Access key ID
	AccessKeyID string `toml:"access_key_id"`

	// AWS Secret Access Key
	SecretAccessKey string `toml:"secret_access_key"`

	// AWS Session Token
	SessionToken string `toml:"session_token"`
}

func (c Credentials) Validate() error {
	if c.AccessKeyID == "" {
		return fmt.Errorf("missing access_key_id")
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("missing secret_access_key")
	}

	if c.SessionToken == "" {
		return fmt.Errorf("missing session_token")
	}

	return nil
}

func (c Config) GetAWSConfig(ctx context.Context) (aws.Config, error) {
	if err := c.Credentials.Validate(); err != nil {
		return aws.Config{}, fmt.Errorf("failed to validate credentials: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				c.Credentials.AccessKeyID,
				c.Credentials.SecretAccessKey,
				c.Credentials.SessionToken)),
		config.WithRegion(c.Region),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to get aws config: %w", err)
	}
	return cfg, nil
}
