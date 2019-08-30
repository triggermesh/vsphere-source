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

package controller

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/triggermesh/vsphere-source/pkg/apis/sources/v1alpha1"
	"github.com/triggermesh/vsphere-source/pkg/controller/resources"
	"go.uber.org/zap"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"knative.dev/eventing-contrib/pkg/controller/sdk"
	"knative.dev/eventing-contrib/pkg/controller/sinks"
	"knative.dev/eventing-contrib/pkg/reconciler/eventtype"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// controllerAgentName is the string used by this controller to
	// identify itself when creating events.
	controllerAgentName = "vsphere-source-controller"

	// raImageEnvVar is the name of the environment variable that
	// contains the receive adapter's image. It must be defined.
	raImageEnvVar = "VSPHERE_RA_IMAGE"

	finalizerName = controllerAgentName
)

type reconciler struct {
	client              client.Client
	dynamicClient       dynamic.Interface
	scheme              *runtime.Scheme
	eventTypeReconciler eventtype.Reconciler

	receiveAdapterImage string
}

// Add creates a new VsphereSource Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, logger *zap.SugaredLogger) error {
	raImage, defined := os.LookupEnv(raImageEnvVar)
	if !defined {
		return fmt.Errorf("required environment variable '%s' not defined", raImageEnvVar)
	}

	log.Println("Adding the VMware vSphere Source controller.")
	p := &sdk.Provider{
		AgentName: controllerAgentName,
		Parent:    &v1alpha1.VSphereSource{},
		Owns:      []runtime.Object{&v1.Deployment{}, &eventingv1alpha1.EventType{}},
		Reconciler: &reconciler{
			scheme:              mgr.GetScheme(),
			receiveAdapterImage: raImage,
			eventTypeReconciler: eventtype.Reconciler{
				Scheme: mgr.GetScheme(),
			},
		},
	}

	return p.Add(mgr, logger)
}

func (r *reconciler) InjectClient(c client.Client) error {
	r.client = c
	r.eventTypeReconciler.Client = c
	return nil
}

func (r *reconciler) InjectConfig(c *rest.Config) error {
	var err error
	r.dynamicClient, err = dynamic.NewForConfig(c)
	return err
}

// Reconcile reads that state of the cluster for a VsphereSource object and makes changes based on the state read
// and what is in the VsphereSource.Spec
func (r *reconciler) Reconcile(ctx context.Context, object runtime.Object) error {
	logger := logging.FromContext(ctx).Desugar()
	var err error

	src, ok := object.(*v1alpha1.VSphereSource)
	if !ok {
		logger.Error("could not find vSphere source", zap.Any("object", object))
		return nil
	}

	// See if the source has been deleted.
	deletionTimestamp := src.DeletionTimestamp
	if deletionTimestamp != nil {
		r.removeFinalizer(src)
		return nil
	}

	r.addFinalizer(src)

	src.Status.InitializeConditions()

	sinkURI, err := sinks.GetSinkURI(ctx, r.client, src.Spec.Sink, src.Namespace)
	if err != nil {
		src.Status.MarkNoSink("NotFound", "")
		return err
	}
	src.Status.MarkSink(sinkURI)

	_, err = r.createReceiveAdapter(ctx, src, sinkURI)
	if err != nil {
		logger.Error("Unable to create the receive adapter", zap.Error(err))
		return err
	}
	src.Status.MarkDeployed()

	err = r.reconcileEventTypes(ctx, src)
	if err != nil {
		logger.Error("Unable to reconcile the event types", zap.Error(err))
		return err
	}
	src.Status.MarkEventTypes()

	return nil
}

func (r *reconciler) createReceiveAdapter(ctx context.Context, src *v1alpha1.VSphereSource, sinkURI string) (*v1.Deployment, error) {
	ra, err := r.getReceiveAdapter(ctx, src)
	if err != nil && !apierrors.IsNotFound(err) {
		logging.FromContext(ctx).Error("Unable to get an existing receive adapter", zap.Error(err))
		return nil, err
	}

	if ra != nil {
		logging.FromContext(ctx).Desugar().Info("Reusing existing receive adapter", zap.Any("receiveAdapter", ra))
		return ra, nil
	}

	svc := resources.MakeReceiveAdapter(&resources.ReceiveAdapterArgs{
		Image:   r.receiveAdapterImage,
		Source:  src,
		Labels:  getLabels(src),
		SinkURI: sinkURI,
	})

	if err := controllerutil.SetControllerReference(src, svc, r.scheme); err != nil {
		return nil, err
	}

	err = r.client.Create(ctx, svc)
	logging.FromContext(ctx).Desugar().Info("Receive Adapter created.", zap.Error(err), zap.Any("receiveAdapter", svc))
	return svc, err
}

func (r *reconciler) getReceiveAdapter(ctx context.Context, src *v1alpha1.VSphereSource) (*v1.Deployment, error) {
	dl := &v1.DeploymentList{}
	err := r.client.List(ctx, &client.ListOptions{
		Namespace:     src.Namespace,
		LabelSelector: r.getLabelSelector(src),
		// TODO this is only needed by the fake client. Real K8s does not need it. Remove it once
		// the fake is fixed.
		Raw: &metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1.SchemeGroupVersion.String(),
				Kind:       "Deployment",
			},
		},
	}, dl)

	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Unable to list deployments: %v", zap.Error(err))
		return nil, err
	}
	for _, dep := range dl.Items {
		if metav1.IsControlledBy(&dep, src) {
			return &dep, nil
		}
	}
	return nil, apierrors.NewNotFound(schema.GroupResource{}, "")
}

func (r *reconciler) reconcileEventTypes(ctx context.Context, src *v1alpha1.VSphereSource) error {
	args := r.newEventTypeReconcilerArgs(src)
	return r.eventTypeReconciler.Reconcile(ctx, src, args)
}

func (r *reconciler) newEventTypeReconcilerArgs(src *v1alpha1.VSphereSource) *eventtype.ReconcilerArgs {
	spec := eventingv1alpha1.EventTypeSpec{
		Type:   v1alpha1.VSphereSourceEventType,
		Source: src.Spec.VSphereURL,
		Broker: src.Spec.Sink.Name,
	}
	specs := make([]eventingv1alpha1.EventTypeSpec, 0, 1)
	specs = append(specs, spec)
	return &eventtype.ReconcilerArgs{
		Specs:     specs,
		Namespace: src.Namespace,
		Labels:    getLabels(src),
		Kind:      src.Spec.Sink.Kind,
	}
}

func (r *reconciler) addFinalizer(s *v1alpha1.VSphereSource) {
	finalizers := sets.NewString(s.Finalizers...)
	finalizers.Insert(finalizerName)
	s.Finalizers = finalizers.List()
}

func (r *reconciler) removeFinalizer(s *v1alpha1.VSphereSource) {
	finalizers := sets.NewString(s.Finalizers...)
	finalizers.Delete(finalizerName)
	s.Finalizers = finalizers.List()
}

func (r *reconciler) getLabelSelector(src *v1alpha1.VSphereSource) labels.Selector {
	return labels.SelectorFromSet(getLabels(src))
}

func getLabels(src *v1alpha1.VSphereSource) map[string]string {
	return map[string]string{
		"knative-eventing-source":      controllerAgentName,
		"knative-eventing-source-name": src.Name,
	}
}
