//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DB) DeepCopyInto(out *DB) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DB.
func (in *DB) DeepCopy() *DB {
	if in == nil {
		return nil
	}
	out := new(DB)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KlovercloudCD) DeepCopyInto(out *KlovercloudCD) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KlovercloudCD.
func (in *KlovercloudCD) DeepCopy() *KlovercloudCD {
	if in == nil {
		return nil
	}
	out := new(KlovercloudCD)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KlovercloudCD) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KlovercloudCDList) DeepCopyInto(out *KlovercloudCDList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KlovercloudCD, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KlovercloudCDList.
func (in *KlovercloudCDList) DeepCopy() *KlovercloudCDList {
	if in == nil {
		return nil
	}
	out := new(KlovercloudCDList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KlovercloudCDList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KlovercloudCDSpec) DeepCopyInto(out *KlovercloudCDSpec) {
	*out = *in
	out.Database = in.Database
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KlovercloudCDSpec.
func (in *KlovercloudCDSpec) DeepCopy() *KlovercloudCDSpec {
	if in == nil {
		return nil
	}
	out := new(KlovercloudCDSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KlovercloudCDStatus) DeepCopyInto(out *KlovercloudCDStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KlovercloudCDStatus.
func (in *KlovercloudCDStatus) DeepCopy() *KlovercloudCDStatus {
	if in == nil {
		return nil
	}
	out := new(KlovercloudCDStatus)
	in.DeepCopyInto(out)
	return out
}
