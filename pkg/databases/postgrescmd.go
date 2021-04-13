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
	"path/filepath"

	apiv1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	"github.com/appscode/go/types"
	shell "github.com/codeskyblue/go-sh"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func addPostgresCMD(cmds *cobra.Command) {
	var pgName string
	var dbname string
	var namespace string
	var fileName string
	var command string
	var pgCmd = &cobra.Command{
		Use:   "postgres",
		Short: "Use to operate postgres pods",
		Long:  `Use this cmd to operate postgres pods.`,
		Run: func(cmd *cobra.Command, args []string) {
			println("Use subcommands such as connect or apply to operate PSQL pods")
		},
	}
	var pgConnectCmd = &cobra.Command{
		Use:   "connect",
		Short: "Connect to a postgres object's pod",
		Long:  `Use this cmd to exec into a postgres object's primary pod.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatal("Enter postgres object's name as an argument")
			}
			pgName = args[0]

			podName, secretName, err := getPostgresInfo(namespace, pgName)
			if err != nil {
				log.Fatal(err)
			}

			auth, tunnel, err := tunnelToDBPod(pgPort, namespace, podName, secretName)
			if err != nil {
				log.Fatal("Couldn't tunnel through. Error = ", err)
			}
			pgConnect(auth, tunnel.Local)
			tunnel.Close()
		},
	}

	var pgApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Apply commands to a postgres object's pod",
		Long: `Use this cmd to apply commands from a file to a postgres object's primary pod.
				Syntax: $ kubedb postgres apply <pg-object-name> -n <namespace> -f <fileName>
				`,
		Run: func(cmd *cobra.Command, args []string) {
			println("Applying commands")
			if len(args) == 0 {
				log.Fatal("Enter postgres object's name as an argument. Your commands will be applied on a database inside it's primary pod")
			}
			pgName = args[0]

			if fileName == "" && command == "" {
				log.Fatal(" Use --file or --command to apply supported commands to a postgres object's pods")
			}

			podName, secretName, err := getPostgresInfo(namespace, pgName)
			if err != nil {
				log.Fatal(err)
			}

			auth, tunnel, err := tunnelToDBPod(pgPort, namespace, podName, secretName)
			if err != nil {
				log.Fatal(err)
			}

			if command != "" {
				pgApplyCommand(auth, tunnel.Local, dbname, command)
			}

			if fileName != "" {
				pgApplySql(auth, tunnel.Local, fileName)
			}

			tunnel.Close()
		},
	}

	cmds.AddCommand(pgCmd)
	pgCmd.AddCommand(pgConnectCmd)
	pgCmd.AddCommand(pgApplyCmd)
	pgCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace of the postgres object to connect to.")

	pgApplyCmd.Flags().StringVarP(&fileName, "file", "f", "", "path to sql file")
	pgApplyCmd.Flags().StringVarP(&command, "command", "c", "", "command to execute")
	pgApplyCmd.Flags().StringVarP(&dbname, "dbname", "d", "postgres", "name of database inside postgres object's pod to execute command")
}

func pgConnect(auth *corev1.Secret, localPort int) {
	sh := shell.NewSession()
	sh.SetEnv("PGPASSWORD", string(auth.Data["POSTGRES_PASSWORD"]))
	sh.ShowCMD = false

	err := sh.Command("docker", "run", "--network=host", "-it",
		"postgres:11.1-alpine",
		"psql",
		"--host=127.0.0.1", fmt.Sprintf("--port=%d", localPort),
		fmt.Sprintf("--username=%s", auth.Data["POSTGRES_USER"])).SetStdin(os.Stdin).Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func pgApplySql(auth *corev1.Secret, localPort int, fileName string) {
	sh := shell.NewSession()
	sh.SetEnv("PGPASSWORD", string(auth.Data["POSTGRES_PASSWORD"]))
	sh.ShowCMD = false

	fileName, err := filepath.Abs(fileName)
	if err != nil {
		log.Fatalln(err)
	}

	err = sh.Command("docker", "run", "--network=host", "-it", "-v",
		fmt.Sprintf("%s:/tmp/pgsql.sql", fileName),
		"postgres:11.1-alpine",
		"psql",
		"--host=127.0.0.1", fmt.Sprintf("--port=%d", localPort),
		fmt.Sprintf("--username=%s", auth.Data["POSTGRES_USER"]),
		"--file=/tmp/pgsql.sql").SetStdin(os.Stdin).Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func pgApplyCommand(auth *corev1.Secret, localPort int, dbname string, command string) {
	sh := shell.NewSession()
	sh.SetEnv("PGPASSWORD", string(auth.Data["POSTGRES_PASSWORD"]))

	sh.ShowCMD = false

	err := sh.Command("docker", "run", "--network=host", "-it",
		"postgres:11.1-alpine",
		"psql",
		"--host=127.0.0.1", fmt.Sprintf("--port=%d", localPort),
		fmt.Sprintf("--dbname=%s", dbname),
		fmt.Sprintf("--username=%s", auth.Data["POSTGRES_USER"]),
		fmt.Sprintf("--command=%s", command)).SetStdin(os.Stdin).Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func getPostgresInfo(namespace string, dbObjectName string) (podName string, secretName string, err error) {
	config, err := getKubeConfig()
	if err != nil {
		return "", "", err
	}
	dbClient := cs.NewForConfigOrDie(config)
	postgres, err := dbClient.KubedbV1alpha2().Postgreses(namespace).Get(context.TODO(), dbObjectName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	if postgres.Status.Phase != apiv1alpha2.DatabasePhaseReady {
		return "", "", errors.New("postgres is not ready")
	}
	secretName = postgres.Spec.AuthSecret.Name
	if postgres.Spec.Replicas == types.Int32P(1) {
		podName = dbObjectName + "-0" //standalone mode
	} else {
		client := kubernetes.NewForConfigOrDie(config)

		for i := 0; i < int(*(postgres.Spec.Replicas)); i++ {
			tempPodName := fmt.Sprintf(dbObjectName+"-%v", i)
			pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), tempPodName, metav1.GetOptions{})
			if err != nil {
				return "", "", err
			}

			if pod.Labels[apiv1alpha2.LabelRole] == primaryRoleLabel {
				podName = tempPodName
				break
			}
		}
	}
	return podName, secretName, nil
}
