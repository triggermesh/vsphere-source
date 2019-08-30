/*
Copyright 2019 Triggermesh, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"log"
	"os"

	"github.com/triggermesh/vsphere-source/pkg/adapter"
	"knative.dev/pkg/signals"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"golang.org/x/net/context"
)

const (
	envSinkURI         = "SINK_URI"
	envVSphereURL      = "VSPHERE_URL"
	envVSphereUser     = "VSPHERE_USER"
	envVSpherePassword = "VSPHERE_PASSWORD"
)

func getRequiredEnv(envKey string) string {
	val, defined := os.LookupEnv(envKey)
	if !defined {
		log.Fatalf("required environment variable not defined '%s'", envKey)
	}
	return val
}

func main() {
	flag.Parse()

	ctx := context.Background()
	logCfg := zap.NewProductionConfig()
	logCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := logCfg.Build()
	if err != nil {
		log.Fatalf("Unable to create logger: %v", err)
	}

	adapter := &adapter.Adapter{
		VSphereURL:           getRequiredEnv(envVSphereURL),
		SinkURI:              getRequiredEnv(envSinkURI),
		VSphereUser:          getRequiredEnv(envVSphereUser),
		VSpherePassword:      getRequiredEnv(envVSpherePassword),
		OnFailedPollWaitSecs: 2,
	}

	logger.Info("Starting VMware Receive Adapter.", zap.Any("adapter", adapter))
	stopCh := signals.SetupSignalHandler()
	if err := adapter.Start(ctx, stopCh); err != nil {
		logger.Fatal("failed to start adapter: ", zap.Error(err))
	}
}
