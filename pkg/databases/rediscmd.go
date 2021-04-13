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
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	apiv1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	shell "github.com/codeskyblue/go-sh"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func addRedisCMD(cmds *cobra.Command) {
	var redisName string
	var dbname string
	var namespace string
	var fileName string
	var command string
	var redisCmd = &cobra.Command{
		Use:   "redis",
		Short: "Use to operate redis pods",
		Long:  `Use this cmd to operate redis pods.`,
		Run: func(cmd *cobra.Command, args []string) {
			println("Use subcommands such as connect or apply to operate PSQL pods")
		},
	}
	var redisConnectCmd = &cobra.Command{
		Use:   "connect",
		Short: "Connect to a redis object's pod",
		Long:  `Use this cmd to exec into a redis object's primary pod.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatal("Enter redis object's name as an argument")
			}
			redisName = args[0]

			podName, err := getRedisInfo(namespace, redisName)
			if err != nil {
				log.Fatal(err)
			}
			redisConnect(namespace, podName)
		},
	}

	var redisApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Apply commands to a redis object's pod",
		Long: `Use this cmd to apply commands from a file to a redis object's primary pod.
				Syntax: $ kubedb redis apply <redis-object-name> -n <namespace> -f <fileName>
				`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatal("Enter redis object's name as an argument. Your commands will be applied on a database inside it's primary pod")
			}
			redisName = args[0]

			if fileName == "" && command == "" {
				log.Fatal(" Use --file or --command to apply supported commands to a redis object's pods")
			}

			podName, err := getRedisInfo(namespace, redisName)
			if err != nil {
				log.Fatal(err)
			}

			if command != "" {
				redisApplyCommand(namespace, podName, command)
			}

			if fileName != "" {
				redisApplyFile(namespace, podName, fileName)
			}
		},
	}

	cmds.AddCommand(redisCmd)
	redisCmd.AddCommand(redisConnectCmd)
	redisCmd.AddCommand(redisApplyCmd)
	redisCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace of the redis object to connect to.")

	redisApplyCmd.Flags().StringVarP(&fileName, "file", "f", "", "path to sql file")
	redisApplyCmd.Flags().StringVarP(&command, "command", "c", "", "command to execute")
	redisApplyCmd.Flags().StringVarP(&dbname, "dbname", "d", "redis", "name of database inside redis object's pod to execute command")
}

func redisConnect(namespace string, podName string) {
	sh := shell.NewSession()
	sh.ShowCMD = false

	err := sh.Command("kubectl", "exec",
		"-it", "-n", namespace, podName, "--",
		"redis-cli", "-n", "0", "-c",
		"-h", podName, "-p", strconv.Itoa(apiv1alpha2.RedisDatabasePort),
	).SetStdin(os.Stdin).Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func redisApplyFile(namespace, podName, fileName string) {
	fileName, err := filepath.Abs(fileName)
	if err != nil {
		log.Fatalln(err)
	}

	var reader io.Reader
	reader, err = os.Open(fileName)
	if err != nil {
		log.Fatalln(err)
	}

	println("Applying commands from file: ", fileName)
	sh := shell.NewSession()
	sh.ShowCMD = false
	err = sh.Command("kubectl", "exec",
		"-i", "-n", namespace, podName, "--",
		"redis-cli", "-n", "0", "-c",
		"-h", podName, "-p", strconv.Itoa(apiv1alpha2.RedisDatabasePort),
	).SetStdin(reader).Run()
	if err != nil {
		log.Fatalln(err)
	}
	println("Command(s) applied")
}

func redisApplyCommand(namespace, podName, command string) {
	println("Applying commands from console: ", command)
	command = strings.ReplaceAll(command, ";", "\n")
	reader := strings.NewReader(command)

	sh := shell.NewSession()
	sh.ShowCMD = false
	err := sh.Command("kubectl", "exec",
		"-i", "-n", namespace, podName, "--",
		"redis-cli", "-n", "0", "-c",
		"-h", podName, "-p", strconv.Itoa(apiv1alpha2.RedisDatabasePort),
	).SetStdin(reader).Run()
	if err != nil {
		log.Fatalln(err)
	}
	println("Command(s) applied")
}

func getRedisInfo(namespace, dbObjectName string) (podName string, err error) {
	config, err := getKubeConfig()
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}
	dbClient := cs.NewForConfigOrDie(config)
	redis, err := dbClient.KubedbV1alpha2().Redises(namespace).Get(context.TODO(), dbObjectName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if redis.Status.Phase != apiv1alpha2.DatabasePhaseReady {
		return "", errors.New("redis is not ready")
	}
	//if cluster is enabled
	client := kubernetes.NewForConfigOrDie(config)
	label := labels.Set(redis.OffshootLabels()).String()
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return "", err
	}
	for _, pod := range pods.Items {
		podName = pod.Name
		break
	}
	return podName, nil
}

//TODO: redis apply lua script
