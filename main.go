// SPDX-License-Identifier: Apache-2.0
// Copyright 2024 Cloudbase Solutions SRL
//
//	Licensed under the Apache License, Version 2.0 (the "License"); you may
//	not use this file except in compliance with the License. You may obtain
//	a copy of the License at
//
//	     http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//	WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//	License for the specific language governing permissions and limitations
//	under the License.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudbase/garm-provider-aws/provider"
	"github.com/cloudbase/garm-provider-common/execution"
)

var signals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
}

var (
	// Version is the version of the application
	Version = "v0.0.0-unknown"
)

func main() {
	// This is an unofficial command. It will be added into future versions of the
	// external provider interface. For now we manually hardcode it here. This is not
	// used by GARM itself. It is informative for the user to be able to check the version
	// of the provider.
	garmCommand := os.Getenv("GARM_COMMAND")
	if garmCommand == "GetVersion" {
		fmt.Println(Version)
		os.Exit(0)
	}

	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	executionEnv, err := execution.GetEnvironment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting environment: %q", err)
		os.Exit(1)
	}

	prov, err := provider.NewAwsProvider(ctx, executionEnv.ProviderConfigFile, executionEnv.ControllerID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating provider: %q", err)
		os.Exit(1)
	}

	result, err := execution.Run(ctx, prov, executionEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run command: %+v\n", err)
		os.Exit(1)
	}
	if len(result) > 0 {
		fmt.Fprint(os.Stdout, result)
	}
}
