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
	"errors"
	"fmt"
	"log"
	"os"

	apiv1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	"kubedb.dev/cli/pkg/lib"

	shell "github.com/codeskyblue/go-sh"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func AddElasticsearchCMD(cmds *cobra.Command) {
	var esName string
	var namespace string
	var esCmd = &cobra.Command{
		Use:   "elasticsearch",
		Short: "Use to operate elasticsearch pods",
		Long:  `Use this cmd to operate elasticsearch pods.`,
		Run: func(cmd *cobra.Command, args []string) {
			println("Use subcommands such as connect or apply to operate PSQL pods")
		},
	}
	var esConnectCmd = &cobra.Command{
		Use:   "connect",
		Short: "Connect to a elasticsearch object's pod",
		Long:  `Use this cmd to exec into a elasticsearch object's primary pod.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatal("Enter elasticsearch object's name as an argument")
			}
			esName = args[0]

			podName, secretName, err := getElasticsearchInfo(namespace, esName)
			if err != nil {
				log.Fatal(err)
			}

			auth, tunnel, err := lib.TunnelToDBPod(esPort, namespace, podName, secretName)
			if err != nil {
				log.Fatal("Couldn't tunnel through. Error = ", err)
			}
			esConnect(auth, tunnel.Local)
			tunnel.Close()
		},
	}

	cmds.AddCommand(esCmd)
	esCmd.AddCommand(esConnectCmd)
	esCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace of the elasticsearch object to connect to.")
}

func esConnect(auth *corev1.Secret, localPort int) {
	sh := shell.NewSession()
	err := sh.Command("docker", "run", "--network=host", "-it",
		"-e", fmt.Sprintf("USERNAME=%s", auth.Data[esAdminUsername]), "-e", fmt.Sprintf("PASSWORD=%s", auth.Data[esAdminPassword]),
		"-e", fmt.Sprintf("ADDRESS=localhost:%d", localPort),
		alpineCurlImg,
	).SetStdin(os.Stdin).Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func getElasticsearchInfo(namespace, dbObjectName string) (podName string, secretName string, err error) {
	config, err := lib.GetKubeConfig()
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}
	dbClient := cs.NewForConfigOrDie(config)
	elasticsearch, err := dbClient.KubedbV1alpha2().Elasticsearches(namespace).Get(context.TODO(), dbObjectName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	if elasticsearch.Status.Phase != apiv1alpha2.DatabasePhaseReady {
		return "", "", errors.New("elasticsearch is not ready")
	}
	client := kubernetes.NewForConfigOrDie(config)
	secretName = dbObjectName + "-auth"
	_, err = client.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	label := labels.Set(elasticsearch.OffshootLabels()).String()
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return "", "", err
	}
	for _, pod := range pods.Items {
		if elasticsearch.Spec.Topology == nil || pod.Labels[esNodeRoleClient] == "set" {
			podName = pod.Name
			break
		}
	}
	return podName, secretName, nil
}
