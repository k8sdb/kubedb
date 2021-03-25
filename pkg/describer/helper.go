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

package describer

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"github.com/fatih/camelcase"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/describe"
	"k8s.io/kubectl/pkg/util/slice"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func describeStorage(st api.StorageType, pvcSpec *core.PersistentVolumeClaimSpec, w describe.PrefixWriter) {
	if st == api.StorageTypeEphemeral {
		w.Write(LEVEL_0, "StorageType:\t%s\n", api.StorageTypeEphemeral)
	} else {
		w.Write(LEVEL_0, "StorageType:\t%s\n", api.StorageTypeDurable)
	}
	if pvcSpec == nil {
		w.Write(LEVEL_0, "No volumes.\n")
		return
	}

	accessModes := getAccessModesAsString(pvcSpec.AccessModes)
	val := pvcSpec.Resources.Requests[core.ResourceStorage]
	capacity := val.String()
	w.Write(LEVEL_0, "Volume:\n")
	if pvcSpec.StorageClassName != nil {
		w.Write(LEVEL_1, "StorageClass:\t%s\n", *pvcSpec.StorageClassName)
	}
	w.Write(LEVEL_1, "Capacity:\t%s\n", capacity)
	if accessModes != "" {
		w.Write(LEVEL_1, "Access Modes:\t%s\n", accessModes)
	}
}

func describeInitialization(init *api.InitSpec, w describe.PrefixWriter) {
	if init == nil {
		return
	}

	w.WriteLine("\nInit:")
	if init.Script != nil {
		w.Write(LEVEL_1, "Script Source:\n")
		describeVolume(LEVEL_2, init.Script.VolumeSource, w)
	}
	if init.WaitForInitialRestore {
		w.Write(LEVEL_1, "WaitForInitialRestore:\t%s\n", init.WaitForInitialRestore)
	}
}

func describeMonitor(monitor *mona.AgentSpec, w describe.PrefixWriter) {
	if monitor == nil {
		return
	}

	w.Write(LEVEL_0, "\n")
	w.Write(LEVEL_0, "Monitoring System:\n")
	w.Write(LEVEL_0, "  Agent:\t%s\n", monitor.Agent)
	if monitor.Prometheus != nil {
		prom := monitor.Prometheus
		w.Write(LEVEL_0, "  Prometheus:\n")
		if prom.Exporter.Port != 0 {
			w.Write(LEVEL_0, "    Exporter:\n")
			w.Write(LEVEL_0, "      Port:\t%v\n", prom.Exporter.Port)
		}
		if prom.ServiceMonitor != nil {
			w.Write(LEVEL_0, "    ServiceMonitor:\n")
			if prom.ServiceMonitor.Labels != nil {
				printLabelsMultiline(LEVEL_0, w, "      Labels", prom.ServiceMonitor.Labels)
			}
			if prom.ServiceMonitor.Interval != "" {
				w.Write(LEVEL_0, "      Interval:\t%s\n", prom.ServiceMonitor.Interval)
			}
		}
	}
}

func showAppBinding(ab *appcat.AppBinding, w describe.PrefixWriter) error {
	w.Write(LEVEL_0, "\n")
	w.Write(LEVEL_0, "AppBinding:\n")
	if ab == nil || ab.Name == "" {
		w.Write(LEVEL_1, "AppBinding has not been created yet.\n")
		return nil
	}
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ab)
	if err != nil {
		return err
	}
	printUnstructuredContent(w, LEVEL_1, obj, "",
		".metadata.managedFields",
		".metadata.finalizers",
		".metadata.generation",
		".metadata.resourceVersion",
		".metadata.selfLink",
		".metadata.uid",
		".metadata.ownerReferences")
	return nil
}

func printUnstructuredContent(w describe.PrefixWriter, level int, content map[string]interface{}, skipPrefix string, skip ...string) {
	fields := []string{}
	for field := range content {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	for _, field := range fields {
		value := content[field]
		switch typedValue := value.(type) {
		case map[string]interface{}:
			skipExpr := fmt.Sprintf("%s.%s", skipPrefix, field)
			if slice.ContainsString(skip, skipExpr, nil) {
				continue
			}
			w.Write(level, "%s:\n", smartLabelFor(field))
			printUnstructuredContent(w, level+1, typedValue, skipExpr, skip...)

		case []interface{}:
			skipExpr := fmt.Sprintf("%s.%s", skipPrefix, field)
			if slice.ContainsString(skip, skipExpr, nil) {
				continue
			}
			w.Write(level, "%s:\n", smartLabelFor(field))
			for _, child := range typedValue {
				switch typedChild := child.(type) {
				case map[string]interface{}:
					printUnstructuredContent(w, level+1, typedChild, skipExpr, skip...)
				default:
					w.Write(level+1, "%v\n", typedChild)
				}
			}

		default:
			skipExpr := fmt.Sprintf("%s.%s", skipPrefix, field)
			if slice.ContainsString(skip, skipExpr, nil) {
				continue
			}
			w.Write(level, "%s:\t%v\n", smartLabelFor(field), typedValue)
		}
	}
}

func smartLabelFor(field string) string {
	// skip creating smart label if field name contains
	// special characters other than '-'
	if strings.IndexFunc(field, func(r rune) bool {
		return !unicode.IsLetter(r) && r != '-'
	}) != -1 {
		return field
	}

	commonAcronyms := []string{"API", "URL", "UID", "OSB", "GUID"}
	parts := camelcase.Split(field)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "_" {
			continue
		}

		if slice.ContainsString(commonAcronyms, strings.ToUpper(part), nil) {
			part = strings.ToUpper(part)
		} else {
			part = strings.Title(part)
		}
		result = append(result, part)
	}

	return strings.Join(result, " ")
}

func showWorkload(client kubernetes.Interface, namespace string, selector labels.Selector, w describe.PrefixWriter) {
	pc := client.CoreV1().Pods(namespace)
	opts := metav1.ListOptions{LabelSelector: selector.String()}

	if statefulSets, err := client.AppsV1().StatefulSets(namespace).List(context.TODO(), opts); err == nil {
		for _, s := range statefulSets.Items {
			selector, err := metav1.LabelSelectorAsSelector(s.Spec.Selector)
			if err != nil {
				continue
			}

			running, waiting, succeeded, failed, err := getPodStatusForController(pc, selector)
			if err != nil {
				continue
			}

			describeStatefulSet(&s, running, waiting, succeeded, failed, w)
		}
	}

	if deployments, err := client.AppsV1().Deployments(namespace).List(context.TODO(), opts); err == nil {
		for _, d := range deployments.Items {
			selector, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
			if err != nil {
				continue
			}

			running, waiting, succeeded, failed, err := getPodStatusForController(pc, selector)
			if err != nil {
				continue
			}

			describeDeployment(&d, running, waiting, succeeded, failed, w)
		}
	}

	if services, err := client.CoreV1().Services(namespace).List(context.TODO(), opts); err == nil {
		for _, s := range services.Items {
			endpoints, _ := client.CoreV1().Endpoints(namespace).Get(context.TODO(), s.Name, metav1.GetOptions{})
			describeService(&s, endpoints, w)
		}
	}
}

func showSecret(client kubernetes.Interface, namespace string, secrets map[string]*core.LocalObjectReference, w describe.PrefixWriter) {
	sc := client.CoreV1().Secrets(namespace)

	for key, sv := range secrets {
		secret, err := sc.Get(context.TODO(), sv.Name, metav1.GetOptions{})
		if err != nil {
			continue
		}
		describeSecret(secret, key, w)
	}
}

func showTopology(client kubernetes.Interface, namespace string, selector labels.Selector, specific map[string]labels.Selector, w describe.PrefixWriter) {
	w.Write(LEVEL_0, "\n")
	w.Write(LEVEL_0, "Topology:\n")
	w.Write(LEVEL_0, "  Type\tPod\tStartTime\tPhase\n")
	w.Write(LEVEL_0, "  ----\t---\t---------\t-----\n")

	pods, _ := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	for _, pod := range pods.Items {
		types := make([]string, 0)
		for key, val := range specific {
			if val.Matches(labels.Set(pod.Labels)) {
				types = append(types, key)
			}
		}
		w.Write(LEVEL_0, "  %s\t%s\t%s\t%s\n",
			strings.Join(types, "|"),
			pod.Name,
			pod.Status.StartTime,
			pod.Status.Phase,
		)
	}

	w.Flush()
}

func getAccessModesAsString(modes []core.PersistentVolumeAccessMode) string {
	modes = removeDuplicateAccessModes(modes)
	var modesStr []string
	if containsAccessMode(modes, core.ReadWriteOnce) {
		modesStr = append(modesStr, "RWO")
	}
	if containsAccessMode(modes, core.ReadOnlyMany) {
		modesStr = append(modesStr, "ROX")
	}
	if containsAccessMode(modes, core.ReadWriteMany) {
		modesStr = append(modesStr, "RWX")
	}
	return strings.Join(modesStr, ",")
}

func removeDuplicateAccessModes(modes []core.PersistentVolumeAccessMode) []core.PersistentVolumeAccessMode {
	var accessModes []core.PersistentVolumeAccessMode
	for _, m := range modes {
		if !containsAccessMode(accessModes, m) {
			accessModes = append(accessModes, m)
		}
	}
	return accessModes
}

func containsAccessMode(modes []core.PersistentVolumeAccessMode, mode core.PersistentVolumeAccessMode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}
