/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package v1alpha1

import (
	v1alpha1 "github.com/trisberg/service/pkg/apis/service/v1alpha1"
	scheme "github.com/trisberg/service/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// HandlersGetter has a method to return a HandlerInterface.
// A group's client should implement this interface.
type BindingsGetter interface {
	Bindings(namespace string) BindingInterface
}

// HandlerInterface has methods to work with Handler resources.
type BindingInterface interface {
	Create(*v1alpha1.Binding) (*v1alpha1.Binding, error)
	Update(*v1alpha1.Binding) (*v1alpha1.Binding, error)
	UpdateStatus(*v1alpha1.Binding) (*v1alpha1.Binding, error)
	Delete(name string, options *v1.DeleteOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Binding, error)
	List(opts v1.ListOptions) (*v1alpha1.BindingList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Binding, err error)
	BindingExpansion
}

// bindings implements HandlerInterface
type bindings struct {
	client rest.Interface
	ns     string
}

// newHandlers returns a Handlers
func newHandlers(c *ServiceV1alpha1Client, namespace string) *bindings {
	return &bindings{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the handler, and returns the corresponding handler object, and an error if there is any.
func (c *bindings) Get(name string, options v1.GetOptions) (result *v1alpha1.Binding, err error) {
	result = &v1alpha1.Binding{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bindings").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Handlers that match those selectors.
func (c *bindings) List(opts v1.ListOptions) (result *v1alpha1.BindingList, err error) {
	result = &v1alpha1.BindingList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("bindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested bindings.
func (c *bindings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("bindings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a handler and creates it.  Returns the server's representation of the handler, and an error, if there is any.
func (c *bindings) Create(handler *v1alpha1.Binding) (result *v1alpha1.Binding, err error) {
	result = &v1alpha1.Binding{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("bindings").
		Body(handler).
		Do().
		Into(result)
	return
}

// Update takes the representation of a handler and updates it. Returns the server's representation of the handler, and an error, if there is any.
func (c *bindings) Update(handler *v1alpha1.Binding) (result *v1alpha1.Binding, err error) {
	result = &v1alpha1.Binding{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("bindings").
		Name(handler.Name).
		Body(handler).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *bindings) UpdateStatus(handler *v1alpha1.Binding) (result *v1alpha1.Binding, err error) {
	result = &v1alpha1.Binding{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("bindings").
		Name(handler.Name).
		SubResource("status").
		Body(handler).
		Do().
		Into(result)
	return
}

// Delete takes name of the handler and deletes it. Returns an error if one occurs.
func (c *bindings) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bindings").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *bindings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("bindings").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched handler.
func (c *bindings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Binding, err error) {
	result = &v1alpha1.Binding{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("bindings").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
