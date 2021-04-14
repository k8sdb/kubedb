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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

const (
	ResourceCodeElasticsearchVersion     = "esversion"
	ResourceKindElasticsearchVersion     = "ElasticsearchVersion"
	ResourceSingularElasticsearchVersion = "elasticsearchversion"
	ResourcePluralElasticsearchVersion   = "elasticsearchversions"
)

// ElasticsearchVersion defines a Elasticsearch database version.

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=elasticsearchversions,singular=elasticsearchversion,scope=Cluster,shortName=esversion,categories={datastore,kubedb,appscode}
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Distribution",type="string",JSONPath=".spec.distribution"
// +kubebuilder:printcolumn:name="DB_IMAGE",type="string",JSONPath=".spec.db.image"
// +kubebuilder:printcolumn:name="Deprecated",type="boolean",JSONPath=".spec.deprecated"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ElasticsearchVersion struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              ElasticsearchVersionSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// ElasticsearchVersionSpec is the spec for elasticsearch version
type ElasticsearchVersionSpec struct {
	// Version
	Version string `json:"version" protobuf:"bytes,1,opt,name=version"`
	// Distribution
	Distribution ElasticsearchDistro `json:"distribution,omitempty" protobuf:"bytes,2,opt,name=distribution,casttype=ElasticsearchDistro"`
	// Authentication plugin used by Elasticsearch cluster
	// Deprecated
	AuthPlugin ElasticsearchAuthPlugin `json:"authPlugin" protobuf:"bytes,3,opt,name=authPlugin,casttype=ElasticsearchAuthPlugin"`
	// Database Image
	DB ElasticsearchVersionDatabase `json:"db" protobuf:"bytes,4,opt,name=db"`
	// Exporter Image
	Exporter ElasticsearchVersionExporter `json:"exporter" protobuf:"bytes,5,opt,name=exporter"`
	// Deprecated versions usable but regarded as obsolete and best avoided, typically due to having been superseded.
	// +optional
	Deprecated bool `json:"deprecated,omitempty" protobuf:"varint,6,opt,name=deprecated"`
	// Init container Image
	InitContainer ElasticsearchVersionInitContainer `json:"initContainer" protobuf:"bytes,7,opt,name=initContainer"`
	// PSP names
	PodSecurityPolicies ElasticsearchVersionPodSecurityPolicy `json:"podSecurityPolicies" protobuf:"bytes,8,opt,name=podSecurityPolicies"`
	// Stash defines backup and restore task definitions.
	// +optional
	Stash appcat.StashAddonSpec `json:"stash,omitempty" protobuf:"bytes,9,opt,name=stash"`
	// SecurityContext is for the additional security information for the Elasticsearch container
	// +optional
	SecurityContext ElasticsearchSecurityContext `json:"securityContext" protobuf:"bytes,10,opt,name=securityContext"`
}

// ElasticsearchVersionDatabase is the Elasticsearch Database image
type ElasticsearchVersionDatabase struct {
	Image string `json:"image" protobuf:"bytes,1,opt,name=image"`
}

// ElasticsearchVersionExporter is the image for the Elasticsearch exporter
type ElasticsearchVersionExporter struct {
	Image string `json:"image" protobuf:"bytes,1,opt,name=image"`
}

// ElasticsearchVersionInitContainer is the Elasticsearch Container initializer
type ElasticsearchVersionInitContainer struct {
	Image   string `json:"image" protobuf:"bytes,1,opt,name=image"`
	YQImage string `json:"yqImage" protobuf:"bytes,2,opt,name=yqImage"`
}

// ElasticsearchVersionPodSecurityPolicy is the Elasticsearch pod security policies
type ElasticsearchVersionPodSecurityPolicy struct {
	DatabasePolicyName string `json:"databasePolicyName" protobuf:"bytes,1,opt,name=databasePolicyName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ElasticsearchVersionList is a list of ElasticsearchVersions
type ElasticsearchVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is a list of ElasticsearchVersion CRD objects
	Items []ElasticsearchVersion `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}

// ElasticsearchSecurityContext provides additional securityContext settings for the Elasticsearch Image
type ElasticsearchSecurityContext struct {
	// RunAsUser is default UID for the DB container. It defaults to 1000.
	RunAsUser *int64 `json:"runAsUser,omitempty" protobuf:"varint,1,opt,name=runAsUser"`

	// RunAsAnyNonRoot will be true if user can change the default UID to other than 1000.
	RunAsAnyNonRoot bool `json:"runAsAnyNonRoot,omitempty" protobuf:"varint,2,opt,name=runAsAnyNonRoot"`
}

// +kubebuilder:validation:Enum=OpenDistro;SearchGuard;X-Pack
type ElasticsearchAuthPlugin string

const (
	ElasticsearchAuthPluginOpenDistro  ElasticsearchAuthPlugin = "OpenDistro"
	ElasticsearchAuthPluginSearchGuard ElasticsearchAuthPlugin = "SearchGuard"
	ElasticsearchAuthPluginXpack       ElasticsearchAuthPlugin = "X-Pack"
)

// +kubebuilder:validation:Enum=ElasticStack;OpenDistro;SearchGuard
type ElasticsearchDistro string

const (
	ElasticsearchDistroElasticStack ElasticsearchDistro = "ElasticStack"
	ElasticsearchDistroOpenDistro   ElasticsearchDistro = "OpenDistro"
	ElasticsearchDistroSearchGuard  ElasticsearchDistro = "SearchGuard"
)
