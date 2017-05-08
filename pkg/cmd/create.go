package cmd

import (
	"errors"
	"io"

	"github.com/k8sdb/kubedb/pkg/cmd/util"
	"github.com/k8sdb/kubedb/pkg/kube"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/runtime"
)

// ref: k8s.io/kubernetes/pkg/kubectl/cmd/create.go

var (
	create_long = templates.LongDesc(`
		Create a resource by filename or stdin.

		JSON and YAML formats are accepted.`)

	create_example = templates.Examples(`
		# Create a elastic using the data in elastic.json.
		kubedb create -f ./elastic.json

		# Create a elastic based on the JSON passed into stdin.
		cat elastic.json | kubedb create -f -`)
)

func NewCmdCreate(out io.Writer, errOut io.Writer) *cobra.Command {
	options := &resource.FilenameOptions{}

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a resource by filename or stdin",
		Long:    create_long,
		Example: create_example,
		Run: func(cmd *cobra.Command, args []string) {
			if cmdutil.IsFilenameEmpty(options.Filenames) {
				defaultRunFunc := cmdutil.DefaultSubCommandRun(errOut)
				defaultRunFunc(cmd, args)
				return
			}
			f := kube.NewKubeFactory(cmd)
			cmdutil.CheckErr(RunCreate(f, out, options))
		},
	}

	util.AddCreateFlags(cmd, options)
	return cmd
}

func RunCreate(f cmdutil.Factory, out io.Writer, options *resource.FilenameOptions) error {
	cmdNamespace, enforceNamespace, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	mapper, typer, err := f.UnstructuredObject()
	if err != nil {
		return err
	}

	r := resource.NewBuilder(
		mapper,
		typer,
		resource.ClientMapperFunc(f.UnstructuredClientForMapping),
		runtime.UnstructuredJSONScheme).
		Schema(util.Validator()).
		ContinueOnError().
		NamespaceParam(cmdNamespace).DefaultNamespace().
		FilenameParam(enforceNamespace, options).
		Flatten().
		Do()

	err = r.Err()
	if err != nil {
		return err
	}

	infoList := make([]*resource.Info, 0)
	err = r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}

		kind := info.GetObjectKind().GroupVersionKind().Kind
		if err := util.CheckSupportedResource(kind); err != nil {
			return err
		}

		infoList = append(infoList, info)
		return nil
	})
	if err != nil {
		return err
	}

	showAlias := false
	if len(infoList) > 1 {
		showAlias = true
	}

	count := 0
	for _, info := range infoList {
		if err := createAndRefresh(info); err != nil {
			return cmdutil.AddSourceToErr("creating", info.Source, err)
		}
		count++
		resourceName := info.Mapping.Resource
		if showAlias {
			if alias, ok := util.ResourceShortFormFor(info.Mapping.Resource); ok {
				resourceName = alias
			}
		}
		cmdutil.PrintSuccess(mapper, false, out, resourceName, info.Name, false, "created")
	}

	if count == 0 {
		return errors.New("no objects passed to create")
	}
	return nil
}

func createAndRefresh(info *resource.Info) error {
	obj, err := resource.NewHelper(info.Client, info.Mapping).Create(info.Namespace, true, info.Object)
	if err != nil {
		return err
	}
	info.Refresh(obj, true)
	return nil
}
