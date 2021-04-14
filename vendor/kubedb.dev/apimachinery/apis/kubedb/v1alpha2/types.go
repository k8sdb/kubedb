/*
Copyright AppsCode Inc. and Contributors

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

package v1alpha2

import (
	core "k8s.io/api/core/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

type InitSpec struct {
	// Initialized indicates that this database has been initialized.
	// This will be set by the operator when status.conditions["Provisioned"] is set to ensure
	// that database is not mistakenly reset when recovered using disaster recovery tools.
	Initialized bool `json:"initialized,omitempty" protobuf:"varint,1,opt,name=initialized"`
	// Wait for initial DataRestore condition
	WaitForInitialRestore bool              `json:"waitForInitialRestore,omitempty" protobuf:"varint,2,opt,name=waitForInitialRestore"`
	Script                *ScriptSourceSpec `json:"script,omitempty" protobuf:"bytes,3,opt,name=script"`
}

type ScriptSourceSpec struct {
	ScriptPath        string `json:"scriptPath,omitempty" protobuf:"bytes,1,opt,name=scriptPath"`
	core.VolumeSource `json:",inline,omitempty" protobuf:"bytes,2,opt,name=volumeSource"`
}

// +kubebuilder:validation:Enum=Provisioning;DataRestoring;Ready;Critical;NotReady;Halted
type DatabasePhase string

const (
	// used for Databases that are currently provisioning
	DatabasePhaseProvisioning DatabasePhase = "Provisioning"
	// used for Databases for which data is currently restoring
	DatabasePhaseDataRestoring DatabasePhase = "DataRestoring"
	// used for Databases that are currently ReplicaReady, AcceptingConnection and Ready
	DatabasePhaseReady DatabasePhase = "Ready"
	// used for Databases that can connect, ReplicaReady == false || Ready == false (eg, ES yellow)
	DatabasePhaseCritical DatabasePhase = "Critical"
	// used for Databases that can't connect
	DatabasePhaseNotReady DatabasePhase = "NotReady"
	// used for Databases that are halted
	DatabasePhaseHalted DatabasePhase = "Halted"
)

// +kubebuilder:validation:Enum=Durable;Ephemeral
type StorageType string

const (
	// default storage type and requires spec.storage to be configured
	StorageTypeDurable StorageType = "Durable"
	// Uses emptyDir as storage
	StorageTypeEphemeral StorageType = "Ephemeral"
)

// +kubebuilder:validation:Enum=Halt;Delete;WipeOut;DoNotTerminate
type TerminationPolicy string

const (
	// Deletes database pods, service but leave the PVCs and stash backup data intact.
	TerminationPolicyHalt TerminationPolicy = "Halt"
	// Deletes database pods, service, pvcs but leave the stash backup data intact.
	TerminationPolicyDelete TerminationPolicy = "Delete"
	// Deletes database pods, service, pvcs and stash backup data.
	TerminationPolicyWipeOut TerminationPolicy = "WipeOut"
	// Rejects attempt to delete database using ValidationWebhook.
	TerminationPolicyDoNotTerminate TerminationPolicy = "DoNotTerminate"
)

// +kubebuilder:validation:Enum=primary;standby;stats
type ServiceAlias string

const (
	PrimaryServiceAlias ServiceAlias = "primary"
	StandbyServiceAlias ServiceAlias = "standby"
	StatsServiceAlias   ServiceAlias = "stats"
)

// +kubebuilder:validation:Enum=DNS;IP;IPv4;IPv6
type AddressType string

const (
	AddressTypeDNS AddressType = "DNS"
	// Uses spec.podIP as address for db pods.
	AddressTypeIP AddressType = "IP"
	// Uses first IPv4 address from spec.podIP, spec.podIPs fields as address for db pods.
	AddressTypeIPv4 AddressType = "IPv4"
	// Uses first IPv6 address from spec.podIP, spec.podIPs fields as address for db pods.
	AddressTypeIPv6 AddressType = "IPv6"
)

func (a AddressType) IsIP() bool {
	return a == AddressTypeIP || a == AddressTypeIPv4 || a == AddressTypeIPv6
}

type NamedServiceTemplateSpec struct {
	// Alias represents the identifier of the service.
	Alias ServiceAlias `json:"alias" protobuf:"bytes,1,opt,name=alias"`

	// ServiceTemplate is an optional configuration for a service used to expose database
	// +optional
	ofst.ServiceTemplateSpec `json:",inline,omitempty" protobuf:"bytes,2,opt,name=serviceTemplateSpec"`
}

type KernelSettings struct {
	// Privileged specifies the status whether the init container
	// requires privileged access to perform the following commands.
	// +optional
	Privileged bool `json:"privileged,omitempty" protobuf:"varint,1,opt,name=privileged"`
	// Sysctls hold a list of sysctls commands needs to apply to kernel.
	// +optional
	Sysctls []core.Sysctl `json:"sysctls,omitempty" protobuf:"bytes,2,rep,name=sysctls"`
}
