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
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		c         *Config
		errString string
	}{
		{
			name: "valid config",
			c: &Config{
				Credentials: Credentials{
					CredentialType: AWSCredentialTypeStaticCredentials,
					StaticCredentials: StaticCredentials{
						AccessKeyID:     "access_key_id",
						SecretAccessKey: "secret_access_key",
						SessionToken:    "session_token",
					},
				},
				SubnetID: "subnet_id",
				Region:   "region",
			},
			errString: "",
		},
		{
			name: "missing subnet_id",
			c: &Config{
				Credentials: Credentials{
					CredentialType: AWSCredentialTypeStaticCredentials,
					StaticCredentials: StaticCredentials{
						AccessKeyID:     "access_key_id",
						SecretAccessKey: "secret_access_key",
						SessionToken:    "session_token",
					},
				},
				Region: "region",
			},
			errString: "missing subnet_id",
		},
		{
			name: "missing region",
			c: &Config{
				Credentials: Credentials{
					CredentialType: AWSCredentialTypeStaticCredentials,
					StaticCredentials: StaticCredentials{
						AccessKeyID:     "access_key_id",
						SecretAccessKey: "secret_access_key",
						SessionToken:    "session_token",
					},
				},
				SubnetID: "subnet_id",
			},
			errString: "missing region",
		},
		{
			name: "missing credential type",
			c: &Config{
				SubnetID: "subnet_id",
				Region:   "region",
			},
			errString: "failed to validate credentials: missing credential_type",
		},
		{
			name: "invalid credential type",
			c: &Config{
				SubnetID: "subnet_id",
				Region:   "region",
				Credentials: Credentials{
					CredentialType: AWSCredentialType("bogus"),
				},
			},
			errString: "failed to validate credentials: unknown credential type: bogus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.Validate()
			if tt.errString == "" {
				require.Nil(t, err)
			} else {
				require.EqualError(t, err, tt.errString)
			}
		})
	}
}

func TestCredentialsValidate(t *testing.T) {
	tests := []struct {
		name      string
		c         Credentials
		errString string
	}{
		{
			name: "valid credentials",
			c: Credentials{
				CredentialType: AWSCredentialTypeStaticCredentials,
				StaticCredentials: StaticCredentials{
					AccessKeyID:     "access_key_id",
					SecretAccessKey: "secret_access_key",
					SessionToken:    "session_token",
				},
			},
			errString: "",
		},
		{
			name: "missing access_key_id",
			c: Credentials{
				CredentialType: AWSCredentialTypeStaticCredentials,
				StaticCredentials: StaticCredentials{
					AccessKeyID:     "",
					SecretAccessKey: "secret_access_key",
					SessionToken:    "session_token",
				},
			},
			errString: "missing access_key_id",
		},
		{
			name: "missing secret_access_key",
			c: Credentials{
				CredentialType: AWSCredentialTypeStaticCredentials,
				StaticCredentials: StaticCredentials{
					AccessKeyID:     "access_key_id",
					SecretAccessKey: "",
					SessionToken:    "session_token",
				},
			},
			errString: "missing secret_access_key",
		},
		{
			name: "missing session_token",
			c: Credentials{
				CredentialType: AWSCredentialTypeStaticCredentials,
				StaticCredentials: StaticCredentials{
					AccessKeyID:     "access_key_id",
					SecretAccessKey: "secret_access_key",
					SessionToken:    "",
				},
			},
			errString: "missing session_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.Validate()
			if tt.errString == "" {
				require.Nil(t, err)
			} else {
				require.EqualError(t, err, tt.errString)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test.toml")
	require.NoError(t, err, "Failed to create temp file")
	defer os.Remove(tempFile.Name())

	// Write some dummy TOML data to the temp file
	dummyTOML := `
		region = "region"
		subnet_id = "subnet_id"
		[credentials]
			credential_type = "static"
			[credentials.static]
			access_key_id = "access_key_id"
			secret_access_key = "secret"
			session_token = "token"
	`

	_, err = tempFile.Write([]byte(dummyTOML))
	require.NoError(t, err, "Failed to write to temp file")

	err = tempFile.Close()
	require.NoError(t, err, "Failed to close temp file")

	// Test case for successful read
	t.Run("success", func(t *testing.T) {
		got, err := NewConfig(tempFile.Name())
		require.NoError(t, err, "NewConfig() should not have returned an error")
		require.Equal(t, &Config{
			Credentials: Credentials{
				CredentialType: AWSCredentialTypeStaticCredentials,
				StaticCredentials: StaticCredentials{
					AccessKeyID:     "access_key_id",
					SecretAccessKey: "secret",
					SessionToken:    "token",
				},
			},
			SubnetID: "subnet_id",
			Region:   "region",
		}, got, "NewConfig() returned unexpected content")
	})

	// Test case for failed read (file does not exist)
	t.Run("fail", func(t *testing.T) {
		_, err := NewConfig("nonexistent.toml")
		require.Error(t, err, "NewConfig() expected an error, got none")
	})

	// Test case for failed read (invalid TOML)
	t.Run("fail", func(t *testing.T) {
		// Create a temporary file
		tempFile, err := os.CreateTemp("", "test.toml")
		require.NoError(t, err, "Failed to create temp file")
		defer os.Remove(tempFile.Name())

		// Write some invalid TOML data to the temp file
		invalidTOML := "invalid TOML"
		_, err = tempFile.Write([]byte(invalidTOML))
		require.NoError(t, err, "Failed to write to temp file")

		err = tempFile.Close()
		require.NoError(t, err, "Failed to close temp file")

		_, err = NewConfig(tempFile.Name())
		require.Error(t, err, "NewConfig() expected an error, got none")
	})
}
