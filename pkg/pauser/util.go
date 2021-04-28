package pauser

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"stash.appscode.dev/apimachinery/apis"
	stash "stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	scs "stash.appscode.dev/apimachinery/client/clientset/versioned/typed/stash/v1beta1"
	scsutil "stash.appscode.dev/apimachinery/client/clientset/versioned/typed/stash/v1beta1/util"
)

func PauseBackupConfiguration(stashClient scs.StashV1beta1Interface, dbMeta metav1.ObjectMeta) error {
	configs, err := stashClient.BackupConfigurations(dbMeta.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var dbBackupConfig *stash.BackupConfiguration
	for _, config := range configs.Items {
		if config.Spec.Target.Ref.Name == dbMeta.Name && config.Spec.Target.Ref.Kind == apis.KindAppBinding {
			dbBackupConfig = &config
			break
		}
	}

	if dbBackupConfig != nil && !dbBackupConfig.Spec.Paused {
		_, err := scsutil.TryUpdateBackupConfiguration(context.TODO(), stashClient, dbBackupConfig.ObjectMeta, func(configuration *stash.BackupConfiguration) *stash.BackupConfiguration {
			configuration.Spec.Paused = true
			return configuration
		}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
