/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package databases

import (
	apiv1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
)

const (
	alpineCurlImg    = "rezoan/alpine-curl:latest"
	alpineTelnetImg  = "rezoan/telnet-curl:latest"
	esAdminUsername  = "ADMIN_USERNAME"
	esAdminPassword  = "ADMIN_PASSWORD"
	esNodeRoleClient = "node.role.client"
	esPort           = apiv1alpha2.ElasticsearchRestPort
	mcPort           = 11211
	mgPassword       = "password"
	mgPort           = apiv1alpha2.MongoDBDatabasePort
	mysqlPort        = apiv1alpha2.MySQLDatabasePort
	pgPort           = 5432
	primaryRoleLabel = "primary"
)
