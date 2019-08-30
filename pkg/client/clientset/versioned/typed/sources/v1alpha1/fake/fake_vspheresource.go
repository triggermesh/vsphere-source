/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/triggermesh/vsphere-source/pkg/apis/sources/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVSphereSources implements VSphereSourceInterface
type FakeVSphereSources struct {
	Fake *FakeSourcesV1alpha1
	ns   string
}

var vspheresourcesResource = schema.GroupVersionResource{Group: "sources.triggermesh.dev", Version: "v1alpha1", Resource: "vspheresources"}

var vspheresourcesKind = schema.GroupVersionKind{Group: "sources.triggermesh.dev", Version: "v1alpha1", Kind: "VSphereSource"}

// Get takes name of the vSphereSource, and returns the corresponding vSphereSource object, and an error if there is any.
func (c *FakeVSphereSources) Get(name string, options v1.GetOptions) (result *v1alpha1.VSphereSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(vspheresourcesResource, c.ns, name), &v1alpha1.VSphereSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VSphereSource), err
}

// List takes label and field selectors, and returns the list of VSphereSources that match those selectors.
func (c *FakeVSphereSources) List(opts v1.ListOptions) (result *v1alpha1.VSphereSourceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(vspheresourcesResource, vspheresourcesKind, c.ns, opts), &v1alpha1.VSphereSourceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VSphereSourceList{ListMeta: obj.(*v1alpha1.VSphereSourceList).ListMeta}
	for _, item := range obj.(*v1alpha1.VSphereSourceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested vSphereSources.
func (c *FakeVSphereSources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(vspheresourcesResource, c.ns, opts))

}

// Create takes the representation of a vSphereSource and creates it.  Returns the server's representation of the vSphereSource, and an error, if there is any.
func (c *FakeVSphereSources) Create(vSphereSource *v1alpha1.VSphereSource) (result *v1alpha1.VSphereSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(vspheresourcesResource, c.ns, vSphereSource), &v1alpha1.VSphereSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VSphereSource), err
}

// Update takes the representation of a vSphereSource and updates it. Returns the server's representation of the vSphereSource, and an error, if there is any.
func (c *FakeVSphereSources) Update(vSphereSource *v1alpha1.VSphereSource) (result *v1alpha1.VSphereSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(vspheresourcesResource, c.ns, vSphereSource), &v1alpha1.VSphereSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VSphereSource), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeVSphereSources) UpdateStatus(vSphereSource *v1alpha1.VSphereSource) (*v1alpha1.VSphereSource, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(vspheresourcesResource, "status", c.ns, vSphereSource), &v1alpha1.VSphereSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VSphereSource), err
}

// Delete takes name of the vSphereSource and deletes it. Returns an error if one occurs.
func (c *FakeVSphereSources) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(vspheresourcesResource, c.ns, name), &v1alpha1.VSphereSource{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVSphereSources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(vspheresourcesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.VSphereSourceList{})
	return err
}

// Patch applies the patch and returns the patched vSphereSource.
func (c *FakeVSphereSources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VSphereSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(vspheresourcesResource, c.ns, name, pt, data, subresources...), &v1alpha1.VSphereSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VSphereSource), err
}
