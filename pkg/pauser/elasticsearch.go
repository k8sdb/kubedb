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

package pauser

import (
	"context"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	cs "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2"
	dbutil "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	kmapi "kmodules.xyz/client-go/api/v1"
	scs "stash.appscode.dev/apimachinery/client/clientset/versioned/typed/stash/v1beta1"
)

type ElasticsearchPauser struct {
	dbClient    cs.KubedbV1alpha2Interface
	stashClient scs.StashV1beta1Interface
	onlyDb      bool
	onlyBackup  bool
}

func NewElasticsearchPauser(clientConfig *rest.Config, onlyDb, onlyBackup bool) (*ElasticsearchPauser, error) {
	// Create DB client
	dbClient, err := cs.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	stashClient, err := scs.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return &ElasticsearchPauser{
		dbClient:    dbClient,
		stashClient: stashClient,
		onlyDb:      onlyDb,
		onlyBackup:  onlyBackup,
	}, nil
}

func (e *ElasticsearchPauser) Pause(name, namespace string) error {
	db, err := e.dbClient.Elasticsearches(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	pauseAll := !(e.onlyBackup || e.onlyDb)

	if e.onlyDb || pauseAll {
		_, err = dbutil.UpdateElasticsearchStatus(context.TODO(), e.dbClient, db.ObjectMeta, func(status *api.ElasticsearchStatus) (types.UID, *api.ElasticsearchStatus) {
			status.Conditions = kmapi.SetCondition(status.Conditions, kmapi.NewCondition(
				api.DatabasePaused,
				"Paused by KubeDB CLI tool",
				db.Generation,
			))
			return db.UID, status
		}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	if e.onlyBackup || pauseAll {
		err = PauseBackupConfiguration(e.stashClient, db.ObjectMeta)
		if err != nil {
			return err
		}
	}

	return nil
}
