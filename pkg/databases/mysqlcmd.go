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
	"strings"

	apiv1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	shell "github.com/codeskyblue/go-sh"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func addMysqlCMD(cmds *cobra.Command) {
	var mysqlName string
	var dbName string
	var namespace string
	var fileName string
	var command string
	var mysqlCmd = &cobra.Command{
		Use:   "mysql",
		Short: "Use to operate mysql pods",
		Long:  `Use this cmd to operate mysql pods.`,
		Run: func(cmd *cobra.Command, args []string) {
			println("Use subcommands such as connect or apply to operate mysql pods")
		},
	}
	var mysqlConnectCmd = &cobra.Command{
		Use:   "connect",
		Short: "Connect to a mysql object's pod",
		Long:  `Use this cmd to exec into a mysql object's primary pod.`,
		Run: func(cmd *cobra.Command, args []string) {
			println("Connect to a mysql pod")
			if len(args) == 0 {
				log.Fatal("Enter mysql object's name as an argument")
			}
			mysqlName = args[0]

			podName, secretName, err := getMysqlInfo(namespace, mysqlName)
			if err != nil {
				log.Fatal(err)
			}

			auth, tunnel, err := tunnelToDBPod(mysqlPort, namespace, podName, secretName)
			if err != nil {
				log.Fatal(err)
			}
			mysqlConnect(auth, tunnel.Local)
			tunnel.Close()
		},
	}

	var mysqlApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Apply commands to a mysql pod",
		Long: `Use this cmd to apply commands from a file to a mysql object's' primary pod.
				Syntax: $ kubedb mysql apply <mysql-name> -n <namespace> -f <fileName>
				`,
		Run: func(cmd *cobra.Command, args []string) {
			println("Applying commands...")
			if len(args) == 0 {
				log.Fatal("Enter mysql object's name as an argument. Your commands will be applied on a database inside it's primary pod")
			}
			mysqlName = args[0]

			podName, secretName, err := getMysqlInfo(namespace, mysqlName)
			if err != nil {
				log.Fatal(err)
			}

			auth, tunnel, err := tunnelToDBPod(mysqlPort, namespace, podName, secretName)
			if err != nil {
				log.Fatal(err)
			}

			if command != "" {
				mysqlApplyCommand(auth, tunnel.Local, dbName, command)
			}

			if fileName != "" {
				mysqlApplyFile(auth, tunnel.Local, dbName, fileName)
			}

			if fileName == "" && command == "" {
				log.Fatal(" Use --file or --command to apply commands to mysql pods")
			}

			tunnel.Close()
		},
	}

	cmds.AddCommand(mysqlCmd)
	mysqlCmd.AddCommand(mysqlConnectCmd)
	mysqlCmd.AddCommand(mysqlApplyCmd)
	mysqlCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Namespace of the mysql object to connect to.")

	mysqlApplyCmd.Flags().StringVarP(&fileName, "file", "f", "", "path to sql file")
	mysqlApplyCmd.Flags().StringVarP(&command, "command", "c", "", "command to execute")
	mysqlApplyCmd.Flags().StringVarP(&dbName, "dbName", "d", "mysql", "name of database inside mysql-db pod to execute command")

}

func mysqlConnect(auth *corev1.Secret, localPort int) {
	sh := shell.NewSession()
	sh.ShowCMD = false
	err := sh.Command("docker", "run",
		"-e", fmt.Sprintf("MYSQL_PWD=%s", auth.Data["password"]),
		"--network=host", "-it", "mysql",
		"mysql",
		"--host=127.0.0.1", fmt.Sprintf("--port=%d", localPort),
		fmt.Sprintf("--user=%s", auth.Data["username"])).SetStdin(os.Stdin).Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func mysqlApplyFile(auth *corev1.Secret, localPort int, dbname string, fileName string) {
	fileName, err := filepath.Abs(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	tempFileName := "/tmp/my.sql"

	println("Applying commands from file: ", fileName)
	sh := shell.NewSession()
	err = sh.Command("docker", "run",
		"--network=host",
		"-e", fmt.Sprintf("MYSQL_PWD=%s", auth.Data["password"]),
		"-v", fmt.Sprintf("%s:%s", fileName, tempFileName), "mysql",
		"mysql",
		"--host=127.0.0.1", fmt.Sprintf("--port=%d", localPort),
		fmt.Sprintf("--user=%s", auth.Data["username"]), dbname,
		"-e", fmt.Sprintf("source %s", tempFileName),
	).SetStdin(os.Stdin).Run()
	if err != nil {
		log.Fatalln(err)
	}
	println("Command(s) applied")
}

func mysqlApplyCommand(auth *corev1.Secret, localPort int, dbname string, command string) {
	println("Applying command(s): ", command)
	sh := shell.NewSession()
	err := sh.Command("docker", "run",
		"-e", fmt.Sprintf("MYSQL_PWD=%s", auth.Data["password"]),
		"--network=host", "mysql",
		"mysql",
		"--host=127.0.0.1", fmt.Sprintf("--port=%d", localPort),
		fmt.Sprintf("--user=%s", auth.Data["username"]),
		dbname, "-e", command,
	).SetStdin(os.Stdin).Run()
	if err != nil {
		log.Fatalln(err)
	}

	println("Command(s) applied")
}

func getMysqlInfo(namespace string, dbObjectName string) (podName string, secretName string, err error) {
	config, err := getKubeConfig()
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}
	dbClient := cs.NewForConfigOrDie(config)
	mysql, err := dbClient.KubedbV1alpha2().MySQLs(namespace).Get(context.TODO(), dbObjectName, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	if mysql.Status.Phase != apiv1alpha2.DatabasePhaseReady {
		return "", "", errors.New("MySQL is not ready")
	}
	secretName = mysql.Spec.AuthSecret.Name
	if mysql.Spec.Topology == nil {
		podName = dbObjectName + "-0" //standalone mode
	} else {
		tempPodName := dbObjectName + "-0"
		client := kubernetes.NewForConfigOrDie(config)
		_, err = client.CoreV1().Pods(namespace).Get(context.TODO(), tempPodName, metav1.GetOptions{})
		if err != nil {
			log.Fatal("Pods are not ready")
		}
		command := "select MEMBER_HOST from performance_schema.replication_group_members" +
			" INNER JOIN performance_schema.global_status ON " +
			"performance_schema.replication_group_members.MEMBER_ID=performance_schema.global_status.VARIABLE_VALUE;"
		auth, tunnel, err := tunnelToDBPod(mysqlPort, namespace, tempPodName, secretName)
		if err != nil {
			log.Fatalln(err)
		}
		sh := shell.NewSession()
		sh.ShowCMD = false
		out, err := sh.Command("docker", "run",
			"-e", fmt.Sprintf("MYSQL_PWD=%s", auth.Data["password"]),
			"--network=host", "-it", "mysql",
			"mysql",
			"--host=127.0.0.1", fmt.Sprintf("--port=%d", tunnel.Local),
			fmt.Sprintf("--user=%s", auth.Data["username"]), "mysql",
			"-NBse", command,
		).SetStdin(os.Stdin).Output()
		if err != nil {
			log.Fatalln(err)
		}
		primaryHostName := strings.TrimPrefix(string(out), " ")
		for i := 0; i < int(*(mysql.Spec.Replicas)); i++ {
			tempPodName = fmt.Sprintf(dbObjectName+"-%v", i)
			if strings.Contains(primaryHostName, tempPodName) {
				podName = tempPodName
				break
			}
		}
	}
	return podName, secretName, nil
}
