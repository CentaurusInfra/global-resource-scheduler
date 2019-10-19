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
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRoles implements RoleInterface
type FakeRoles struct {
	Fake *FakeRbacV1
	ns   string
	te   string
}

var rolesResource = schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"}

var rolesKind = schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"}

// Get takes name of the role, and returns the corresponding role object, and an error if there is any.
func (c *FakeRoles) Get(name string, options v1.GetOptions) (result *rbacv1.Role, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithMultiTenancy(rolesResource, c.ns, name, c.te), &rbacv1.Role{})

	if obj == nil {
		return nil, err
	}

	return obj.(*rbacv1.Role), err
}

// List takes label and field selectors, and returns the list of Roles that match those selectors.
func (c *FakeRoles) List(opts v1.ListOptions) (result *rbacv1.RoleList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithMultiTenancy(rolesResource, rolesKind, c.ns, opts, c.te), &rbacv1.RoleList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &rbacv1.RoleList{ListMeta: obj.(*rbacv1.RoleList).ListMeta}
	for _, item := range obj.(*rbacv1.RoleList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested roles.
func (c *FakeRoles) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithMultiTenancy(rolesResource, c.ns, opts, c.te))

}

// Create takes the representation of a role and creates it.  Returns the server's representation of the role, and an error, if there is any.
func (c *FakeRoles) Create(role *rbacv1.Role) (result *rbacv1.Role, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithMultiTenancy(rolesResource, c.ns, role, c.te), &rbacv1.Role{})

	if obj == nil {
		return nil, err
	}

	return obj.(*rbacv1.Role), err
}

// Update takes the representation of a role and updates it. Returns the server's representation of the role, and an error, if there is any.
func (c *FakeRoles) Update(role *rbacv1.Role) (result *rbacv1.Role, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithMultiTenancy(rolesResource, c.ns, role, c.te), &rbacv1.Role{})

	if obj == nil {
		return nil, err
	}

	return obj.(*rbacv1.Role), err
}

// Delete takes name of the role and deletes it. Returns an error if one occurs.
func (c *FakeRoles) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithMultiTenancy(rolesResource, c.ns, name, c.te), &rbacv1.Role{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRoles) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithMultiTenancy(rolesResource, c.ns, listOptions, c.te)

	_, err := c.Fake.Invokes(action, &rbacv1.RoleList{})
	return err
}

// Patch applies the patch and returns the patched role.
func (c *FakeRoles) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *rbacv1.Role, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithMultiTenancy(rolesResource, c.te, c.ns, name, pt, data, subresources...), &rbacv1.Role{})

	if obj == nil {
		return nil, err
	}

	return obj.(*rbacv1.Role), err
}
