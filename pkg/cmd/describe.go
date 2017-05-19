package cmd

import (
	"fmt"
	"io"

	"github.com/k8sdb/kubedb/pkg/cmd/describer"
	"github.com/k8sdb/kubedb/pkg/cmd/util"
	"github.com/k8sdb/kubedb/pkg/kube"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/runtime"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"
	"strings"
)

var (
	describe_long = templates.LongDesc(`
		Show details of a specific resource or group of resources.
		This command joins many API calls together to form a detailed description of a
		given resource or group of resources.` + valid_resources)

	describe_example = templates.Examples(`
		# Describe a elastic
		kubedb describe elastics elasticsearch-demo

		# Describe a postgres
		kubedb describe pg/postgres-demo

		# Describe all dormantDatabases
		kubedb describe drmn`)
)

func NewCmdDescribe(out, cmdErr io.Writer) *cobra.Command {
	describerSettings := &kubectl.DescriberSettings{}

	cmd := &cobra.Command{
		Use:     "describe (TYPE [NAME_PREFIX] | TYPE/NAME)",
		Short:   "Show details of a specific resource or group of resources",
		Long:    describe_long,
		Example: describe_example,
		Run: func(cmd *cobra.Command, args []string) {
			f := kube.NewKubeFactory(cmd)
			cmdutil.CheckErr(RunDescribe(f, out, cmdErr, cmd, args, describerSettings))
		},
	}

	util.AddDescribeFlags(cmd)
	cmd.Flags().BoolVar(&describerSettings.ShowEvents, "show-events", true, "If true, display events related to the described object.")
	return cmd
}

func RunDescribe(f cmdutil.Factory, out, cmdErr io.Writer, cmd *cobra.Command, args []string, describerSettings *kubectl.DescriberSettings) error {
	allNamespaces := cmdutil.GetFlagBool(cmd, "all-namespaces")
	cmdNamespace, enforceNamespace, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	if allNamespaces {
		enforceNamespace = false
	}
	if len(args) == 0 {
		fmt.Fprint(cmdErr, "You must specify the type of resource to describe. ", valid_resources)
		return cmdutil.UsageError(cmd, "Required resource not specified.")
	}

	mapper, typer, err := f.UnstructuredObject()
	if err != nil {
		return err
	}

	var printAll bool = false
	resources := strings.Split(args[0], ",")
	for i, r := range resources {
		if r == "all" {
			printAll = true
		} else {
			items := strings.Split(r, "/")
			kind, err := util.GetSupportedResourceKind(items[0])
			if err != nil {
				return err
			}
			items[0] = kind
			resources[i] = strings.Join(items, "/")
		}
	}
	if printAll {
		if resources, err = util.GetAllSupportedResources(f); err != nil {
			return err
		}
	}
	args[0] = strings.Join(resources, ",")

	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.UnstructuredClientForMapping), runtime.UnstructuredJSONScheme).
		ContinueOnError().
		NamespaceParam(cmdNamespace).DefaultNamespace().AllNamespaces(allNamespaces).
		FilenameParam(enforceNamespace, &resource.FilenameOptions{}).
		ResourceTypeOrNameArgs(true, args...).
		Flatten().
		Do()
	err = r.Err()
	if err != nil {
		return err
	}

	allErrs := []error{}
	infos, err := r.Infos()
	if err != nil {
		allErrs = append(allErrs, err)
	}

	rDescriber := describer.NewDescriber(f)
	first := true
	for _, info := range infos {
		s, err := rDescriber.Describe(info.Object, describerSettings)
		if err != nil {
			allErrs = append(allErrs, err)
			continue
		}
		if first {
			first = false
			fmt.Fprint(out, s)
		} else {
			fmt.Fprintf(out, "\n\n%s", s)
		}
	}

	return utilerrors.NewAggregate(allErrs)

}
