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
	"log"

	"github.com/triggermesh/vsphere-source/pkg/apis"
	"github.com/triggermesh/vsphere-source/pkg/controller"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"knative.dev/pkg/logging/logkey"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	logCfg := zap.NewProductionConfig()
	logCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := logCfg.Build()
	if err != nil {
		log.Fatal(err)
	}
	logger = logger.With(zap.String(logkey.ControllerType, "vsphere-controller"))

	// Get a config to talk to the apiserver
	log.Println("setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal("Unable to set up client config:", err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	log.Println("setting up manager")
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		log.Fatal("unable to set up overall controller manager:", err)
	}

	log.Println("Registering Components.")

	// Setup Scheme for all resources
	log.Println("setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal("unable add APIs to scheme", err)
	}

	// Setup all Controllers
	log.Println("Setting up controller")
	if err := controller.Add(mgr, logger.Sugar()); err != nil {
		log.Fatal("unable to register controllers to the manager:", err)
	}

	// Start the Cmd
	log.Println("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Fatal("unable to run the manager:", err)
	}
}
