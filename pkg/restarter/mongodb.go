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

package restarter

import (
	"context"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/apis/ops/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2"
	ops "kubedb.dev/apimachinery/client/clientset/versioned/typed/ops/v1alpha1"

	"github.com/appscode/go/crypto/rand"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type MongoDBRestarter struct {
	dbClient  cs.KubedbV1alpha2Interface
	opsClient ops.OpsV1alpha1Interface
}

func NewMongoDBRestarter(clientConfig *rest.Config) (*MongoDBRestarter, error) {
	dbClient, err := cs.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	opsClient, err := ops.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return &MongoDBRestarter{
		dbClient:  dbClient,
		opsClient: opsClient,
	}, nil
}

func (e *MongoDBRestarter) Restart(name, namespace string) (string, error) {
	db, err := e.dbClient.MongoDBs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if db.Status.Phase != api.DatabasePhaseReady {
		return "", fmt.Errorf("can't restart a database which is not in Ready state")
	}

	restartOpsRequest := &v1alpha1.MongoDBOpsRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix(db.Name + "-restart-cli"),
			Namespace: namespace,
		},
		Spec: v1alpha1.MongoDBOpsRequestSpec{
			Type: v1alpha1.OpsRequestTypeRestart,
			DatabaseRef: v1.LocalObjectReference{
				Name: name,
			},
			Restart: &v1alpha1.RestartSpec{},
		},
	}
	_, err = e.opsClient.MongoDBOpsRequests(namespace).Create(context.TODO(), restartOpsRequest, metav1.CreateOptions{})

	return restartOpsRequest.Name, err
}
