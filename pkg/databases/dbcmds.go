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
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	apiv1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"kmodules.xyz/client-go/tools/portforward"
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
	kubeConfigPath   = "KUBEDB_KUBE_CONFIG_PATH"
)

func AddDatabaseCMDs(cmds *cobra.Command) {
	addPostgresCMD(cmds)
	addMysqlCMD(cmds)
	addMongoCMD(cmds)
	addRedisCMD(cmds)
	addElasticsearchCMD(cmds)
	addMemcachedCMD(cmds)
}

func tunnelToDBPod(dbPort int, namespace string, podName string, secretName string) (*corev1.Secret, *portforward.Tunnel, error) {
	//TODO: Always close the tunnel after using thing function
	config, err := getKubeConfig()
	if err != nil {
		return nil, nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	if namespace == "" {
		println("Using default namespace. Enter your namespace using -n=<your-namespace>")
	}

	_, err = client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			fmt.Println("Pod doesn't exist")
		}
		return nil, nil, err
	}
	auth, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	tunnel := portforward.NewTunnel(portforward.TunnelOptions{
		Client:    client.CoreV1().RESTClient(),
		Config:    config,
		Resource:  "pods",
		Name:      podName,
		Namespace: namespace,
		Remote:    dbPort,
	})
	err = tunnel.ForwardPort()
	if err != nil {
		log.Fatalln(err)
	}

	return auth, tunnel, err
}

func getKubeConfig() (*restclient.Config, error) {
	kubeconfigPath := os.Getenv(kubeConfigPath)
	if kubeconfigPath == "" {
		kubeconfigPath = filepath.Join(homedir.HomeDir(), ".kube", "kind-config-kind")
	}
	masterURL := ""

	return clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
}
