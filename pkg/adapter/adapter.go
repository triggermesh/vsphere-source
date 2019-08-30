// Copyright 2019 TriggerMesh, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package adapter

import (
	"context"
	"net/url"
	"strconv"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventtypes "github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/knative/pkg/logging"
	"github.com/triggermesh/vsphere-source/pkg/apis/sources/v1alpha1"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/event"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"go.uber.org/zap"
	"knative.dev/eventing-contrib/pkg/kncloudevents"
)

type Adapter struct {
	// SinkURI is the URI messages will be forwarded on to.
	SinkURI string

	VSphereURL      string
	VSphereUser     string
	VSpherePassword string

	// OnFailedPollWaitSecs determines the interval to wait after a
	// failed poll before making another one
	OnFailedPollWaitSecs time.Duration

	// Client sends cloudevents to the target.
	client client.Client
}

// Initialize cloudevent client
func (a *Adapter) initClient() error {
	if a.client == nil {
		var err error
		if a.client, err = kncloudevents.NewDefaultClient(a.SinkURI); err != nil {
			return err
		}
	}
	return nil
}

func (a *Adapter) Start(ctx context.Context, stopCh <-chan struct{}) error {
	logger := logging.FromContext(ctx)

	logger.Info("Starting with config: ", zap.Any("adapter", a))

	if err := a.initClient(); err != nil {
		logger.Error("Failed to create cloudevent client", zap.Error(err))
		return err
	}

	client, err := vSphereClient(ctx, a.VSphereURL, a.VSphereUser, a.VSpherePassword)
	if err != nil {
		logger.Error("Failed to create vSphere client", zap.Error(err))
		return err
	}

	return a.pollLoop(ctx, client, stopCh)
}

func (a *Adapter) pollLoop(ctx context.Context, client *govmomi.Client, stopCh <-chan struct{}) error {
	logger := logging.FromContext(ctx)
	for {
		select {
		case <-stopCh:
			logger.Info("Exiting")
			return nil
		default:
		}
		events, err := poll(ctx, client, a.SinkURI, 10)
		if err != nil {
			logger.Warn("Failed to poll events from vSphere", zap.Error(err))
			time.Sleep(a.OnFailedPollWaitSecs * time.Second)
			continue
		}
		for _, baseEvent := range events {
			a.receiveMessage(ctx, baseEvent.GetEvent())
		}
	}
}

func (a *Adapter) receiveMessage(ctx context.Context, event *types.Event) {
	logger := logging.FromContext(ctx).With(zap.Any("eventID", event.Key)).With(zap.Any("sink", a.SinkURI))
	logger.Debugw("Received message from vSphere:", zap.Any("message", event))

	err := a.postMessage(ctx, logger, event)
	if err != nil {
		logger.Infof("Event delivery failed: %s", err)
	} else {
		logger.Debug("Message successfully posted to Sink")
	}
}

// postMessage sends an VMware vSphere event to the SinkURI
func (a *Adapter) postMessage(ctx context.Context, logger *zap.SugaredLogger, event *types.Event) error {
	e := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			ID:     strconv.Itoa(int(event.Key)),
			Type:   v1alpha1.VSphereSourceEventType,
			Source: *cloudeventtypes.ParseURLRef(a.VSphereURL),
			Time:   &cloudeventtypes.Timestamp{Time: event.CreatedTime},
		}.AsV02(),
		Data: event.FullFormattedMessage,
	}

	_, err := a.client.Send(context.TODO(), e)
	return err
}

func vSphereClient(ctx context.Context, vSphereURL, user, password string) (*govmomi.Client, error) {
	// Parse URL from string
	u, err := soap.ParseURL(vSphereURL)
	if err != nil {
		return nil, err
	}

	u.User = url.UserPassword(user, password)

	// Connect and log in to ESX or vCenter
	return govmomi.NewClient(ctx, u, true)
}

// poll reads messages from the queue in batches of a given maximum size.
func poll(ctx context.Context, c *govmomi.Client, url string, maxBatchSize int32) ([]types.BaseEvent, error) {
	collector := event.NewHistoryCollector(c.Client, types.ManagedObjectReference{})
	return collector.ReadNextEvents(ctx, maxBatchSize)
}
