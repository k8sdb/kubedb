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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const (
	ResourceCodeMySQL     = "my"
	ResourceKindMySQL     = "MySQL"
	ResourceSingularMySQL = "mysql"
	ResourcePluralMySQL   = "mysqls"
)

// +kubebuilder:validation:Enum=GroupReplication
type MySQLClusterMode string

const (
	MySQLClusterModeGroup MySQLClusterMode = "GroupReplication"
)

// +kubebuilder:validation:Enum=Single-Primary;Multi-Primary
type MySQLGroupMode string

const (
	MySQLGroupModeSinglePrimary MySQLGroupMode = "Single-Primary"
	MySQLGroupModeMultiPrimary  MySQLGroupMode = "Multi-Primary"
)

// Mysql defines a Mysql database.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mysqls,singular=mysql,shortName=my,categories={datastore,kubedb,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MySQL struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              MySQLSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            MySQLStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type MySQLSpec struct {
	// Version of MySQL to be deployed.
	Version string `json:"version" protobuf:"bytes,1,opt,name=version"`

	// Number of instances to deploy for a MySQL database. In case of MySQL group
	// replication, max allowed value is 9 (default 3).
	// (see ref: https://dev.mysql.com/doc/refman/5.7/en/group-replication-frequently-asked-questions.html)
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,2,opt,name=replicas"`

	// MySQL cluster topology
	Topology *MySQLClusterTopology `json:"topology,omitempty" protobuf:"bytes,3,opt,name=topology"`

	// StorageType can be durable (default) or ephemeral
	StorageType StorageType `json:"storageType,omitempty" protobuf:"bytes,4,opt,name=storageType,casttype=StorageType"`

	// Storage spec to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty" protobuf:"bytes,5,opt,name=storage"`

	// Database authentication secret
	AuthSecret *core.LocalObjectReference `json:"authSecret,omitempty" protobuf:"bytes,6,opt,name=authSecret"`

	// Init is used to initialize database
	// +optional
	Init *InitSpec `json:"init,omitempty" protobuf:"bytes,7,opt,name=init"`

	// Monitor is used monitor database instance
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty" protobuf:"bytes,9,opt,name=monitor"`

	// ConfigSecret is an optional field to provide custom configuration file for database (i.e custom-mysql.cnf).
	// If specified, this file will be used as configuration file otherwise default configuration file will be used.
	ConfigSecret *core.LocalObjectReference `json:"configSecret,omitempty" protobuf:"bytes,10,opt,name=configSecret"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate,omitempty" protobuf:"bytes,11,opt,name=podTemplate"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty" protobuf:"bytes,12,rep,name=serviceTemplates"`

	// Indicates that the database server need to be encrypted connections(ssl)
	// +optional
	RequireSSL bool `json:"requireSSL,omitempty" protobuf:"varint,13,opt,name=requireSSL"`

	// TLS contains tls configurations for client and server.
	// +optional
	TLS *kmapi.TLSConfig `json:"tls,omitempty" protobuf:"bytes,14,opt,name=tls"`

	// Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.
	// +optional
	Halted bool `json:"halted,omitempty" protobuf:"varint,15,opt,name=halted"`

	// TerminationPolicy controls the delete operation for database
	// +optional
	TerminationPolicy TerminationPolicy `json:"terminationPolicy,omitempty" protobuf:"bytes,16,opt,name=terminationPolicy,casttype=TerminationPolicy"`

	// Indicated whether to use DNS or IP address to address pods in a db cluster.
	// If IP address is used, HostNetwork will be used. Defaults to DNS.
	// +kubebuilder:default:=DNS
	// +optional
	// +default="DNS"
	UseAddressType AddressType `json:"useAddressType,omitempty" protobuf:"bytes,17,opt,name=useAddressType,casttype=AddressType"`
}

// +kubebuilder:validation:Enum=server;client;metrics-exporter
type MySQLCertificateAlias string

const (
	MySQLServerCert          MySQLCertificateAlias = "server"
	MySQLClientCert          MySQLCertificateAlias = "client"
	MySQLMetricsExporterCert MySQLCertificateAlias = "metrics-exporter"
)

type MySQLClusterTopology struct {
	// If set to -
	// "GroupReplication", GroupSpec is required and MySQL servers will start  a replication group
	Mode *MySQLClusterMode `json:"mode,omitempty" protobuf:"bytes,1,opt,name=mode,casttype=MySQLClusterMode"`

	// Group replication info for MySQL
	Group *MySQLGroupSpec `json:"group,omitempty" protobuf:"bytes,2,opt,name=group"`
}

type MySQLGroupSpec struct {
	// TODO: "Multi-Primary" needs to be implemented
	// Group Replication can be deployed in either "Single-Primary" or "Multi-Primary" mode
	Mode *MySQLGroupMode `json:"mode,omitempty" protobuf:"bytes,1,opt,name=mode,casttype=MySQLGroupMode"`

	// Group name is a version 4 UUID
	// ref: https://dev.mysql.com/doc/refman/5.7/en/group-replication-options.html#sysvar_group_replication_group_name
	Name string `json:"name,omitempty" protobuf:"bytes,2,opt,name=name"`
}

type MySQLStatus struct {
	// Specifies the current phase of the database
	// +optional
	Phase DatabasePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=DatabasePhase"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,2,opt,name=observedGeneration"`
	// Conditions applied to the database, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MySQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is a list of MySQL TPR objects
	Items []MySQL `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}
