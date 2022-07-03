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
func (in *Agent) DeepCopyInto(out *Agent) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Agent.
func (in *Agent) DeepCopy() *Agent {
	if in == nil {
		return nil
	}
	out := new(Agent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApiService) DeepCopyInto(out *ApiService) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApiService.
func (in *ApiService) DeepCopy() *ApiService {
	if in == nil {
		return nil
	}
	out := new(ApiService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Console) DeepCopyInto(out *Console) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Console.
func (in *Console) DeepCopy() *Console {
	if in == nil {
		return nil
	}
	out := new(Console)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CoreEngine) DeepCopyInto(out *CoreEngine) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CoreEngine.
func (in *CoreEngine) DeepCopy() *CoreEngine {
	if in == nil {
		return nil
	}
	out := new(CoreEngine)
	in.DeepCopyInto(out)
	return out
}

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
func (in *EventBank) DeepCopyInto(out *EventBank) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventBank.
func (in *EventBank) DeepCopy() *EventBank {
	if in == nil {
		return nil
	}
	out := new(EventBank)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalAgent) DeepCopyInto(out *ExternalAgent) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalAgent.
func (in *ExternalAgent) DeepCopy() *ExternalAgent {
	if in == nil {
		return nil
	}
	out := new(ExternalAgent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExternalAgent) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalAgentList) DeepCopyInto(out *ExternalAgentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ExternalAgent, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalAgentList.
func (in *ExternalAgentList) DeepCopy() *ExternalAgentList {
	if in == nil {
		return nil
	}
	out := new(ExternalAgentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExternalAgentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalAgentSpec) DeepCopyInto(out *ExternalAgentSpec) {
	*out = *in
	in.Agent.DeepCopyInto(&out.Agent)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalAgentSpec.
func (in *ExternalAgentSpec) DeepCopy() *ExternalAgentSpec {
	if in == nil {
		return nil
	}
	out := new(ExternalAgentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalAgentStatus) DeepCopyInto(out *ExternalAgentStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalAgentStatus.
func (in *ExternalAgentStatus) DeepCopy() *ExternalAgentStatus {
	if in == nil {
		return nil
	}
	out := new(ExternalAgentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IntegrationManager) DeepCopyInto(out *IntegrationManager) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IntegrationManager.
func (in *IntegrationManager) DeepCopy() *IntegrationManager {
	if in == nil {
		return nil
	}
	out := new(IntegrationManager)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KlovercloudCD) DeepCopyInto(out *KlovercloudCD) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
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
	in.Security.DeepCopyInto(&out.Security)
	in.LightHouse.DeepCopyInto(&out.LightHouse)
	in.ApiService.DeepCopyInto(&out.ApiService)
	in.Agent.DeepCopyInto(&out.Agent)
	in.IntegrationManager.DeepCopyInto(&out.IntegrationManager)
	in.EventBank.DeepCopyInto(&out.EventBank)
	in.CoreEngine.DeepCopyInto(&out.CoreEngine)
	in.Console.DeepCopyInto(&out.Console)
	in.Terminal.DeepCopyInto(&out.Terminal)
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LightHouse) DeepCopyInto(out *LightHouse) {
	*out = *in
	in.Command.DeepCopyInto(&out.Command)
	in.Query.DeepCopyInto(&out.Query)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LightHouse.
func (in *LightHouse) DeepCopy() *LightHouse {
	if in == nil {
		return nil
	}
	out := new(LightHouse)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LightHouseCommand) DeepCopyInto(out *LightHouseCommand) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LightHouseCommand.
func (in *LightHouseCommand) DeepCopy() *LightHouseCommand {
	if in == nil {
		return nil
	}
	out := new(LightHouseCommand)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LightHouseQuery) DeepCopyInto(out *LightHouseQuery) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LightHouseQuery.
func (in *LightHouseQuery) DeepCopy() *LightHouseQuery {
	if in == nil {
		return nil
	}
	out := new(LightHouseQuery)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Security) DeepCopyInto(out *Security) {
	*out = *in
	out.User = in.User
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Security.
func (in *Security) DeepCopy() *Security {
	if in == nil {
		return nil
	}
	out := new(Security)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Terminal) DeepCopyInto(out *Terminal) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Terminal.
func (in *Terminal) DeepCopy() *Terminal {
	if in == nil {
		return nil
	}
	out := new(Terminal)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *User) DeepCopyInto(out *User) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new User.
func (in *User) DeepCopy() *User {
	if in == nil {
		return nil
	}
	out := new(User)
	in.DeepCopyInto(out)
	return out
}
