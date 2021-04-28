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

package cmds

import (
	"fmt"

	"kubedb.dev/cli/pkg/pauser"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	pauseLong = templates.LongDesc(`
		Pause the community-operator's watch for the objects.
		The community-operator will stop to process the object. 
    `)

	pauseExample = templates.Examples(`
		# Pause an elasticsearch
		dba pause elasticsearch elasticsearch-demo

		# Pause a postgres
		dba pause pg/postgres-demo

		# Pause all postgres
		dba pause postgreses

 		Valid resource types include:
    		* elasticsearch
			* mongodb
			* mariadb
			* mysql
			* postgres
			* redis
`)
)

type PauseOptions struct {
	CmdParent string
	Selector  string
	Namespace string

	NewBuilder func() *resource.Builder

	BuilderArgs []string

	EnforceNamespace bool
	AllNamespaces    bool

	Factory         cmdutil.Factory
	FilenameOptions *resource.FilenameOptions

	genericclioptions.IOStreams

	onlyDb     bool
	onlyBackup bool
}

func NewCmdPause(parent string, f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := &PauseOptions{
		FilenameOptions: &resource.FilenameOptions{},
		Factory:         f,
		CmdParent:       parent,

		IOStreams: streams,
	}

	cmd := &cobra.Command{
		Use:     "pause (-f FILENAME | TYPE [NAME_PREFIX | -l label] | TYPE/NAME)",
		Short:   i18n.T("Pause the processing of an object."),
		Long:    pauseLong + "\n\n" + cmdutil.SuggestAPIResources("kubectl"),
		Example: pauseExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(args))
			cmdutil.CheckErr(o.Run())
		},
		DisableFlagsInUseLine: true,
		DisableAutoGenTag:     true,
	}
	usage := "containing the resource to pause"
	cmdutil.AddFilenameOptionFlags(cmd, o.FilenameOptions, usage)
	cmd.Flags().StringVarP(&o.Selector, "selector", "l", o.Selector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().BoolVar(&o.AllNamespaces, "all-namespaces", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().BoolVar(&o.onlyDb, "only-db", false, "If provided, only the database is paused.")
	cmd.Flags().BoolVar(&o.onlyBackup, "only-backupconfig", false, "If provided, only the backupconfiguration for the database is paused.")

	return cmd
}

func (o *PauseOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	var err error
	o.Namespace, o.EnforceNamespace, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	if o.AllNamespaces {
		o.EnforceNamespace = false
	}

	if len(args) == 0 && cmdutil.IsFilenameSliceEmpty(o.FilenameOptions.Filenames, o.FilenameOptions.Kustomize) {
		return fmt.Errorf("You must specify the type of resource to describe. %s\n", cmdutil.SuggestAPIResources(o.CmdParent))
	}

	o.BuilderArgs = args

	o.NewBuilder = f.NewBuilder

	return nil
}

func (o *PauseOptions) Validate(args []string) error {
	return nil
}

func (o *PauseOptions) Run() error {
	r := o.NewBuilder().
		Unstructured().
		ContinueOnError().
		NamespaceParam(o.Namespace).DefaultNamespace().AllNamespaces(o.AllNamespaces).
		FilenameParam(o.EnforceNamespace, o.FilenameOptions).
		LabelSelectorParam(o.Selector).
		ResourceTypeOrNameArgs(true, o.BuilderArgs...).
		Flatten().
		Do()
	err := r.Err()
	if err != nil {
		return err
	}

	var allErrs []error
	infos, err := r.Infos()
	if err != nil {
		allErrs = append(allErrs, err)
		return utilerrors.NewAggregate(allErrs)
	}

	if len(infos) == 0 {
		fmt.Fprintf(o.Out, "No resources found in %s namespace.\n", o.Namespace)
		return nil
	}

	errs := sets.NewString()
	for _, info := range infos {
		psr, err := pauser.NewPauser(o.Factory, info.Mapping, o.onlyDb, o.onlyBackup)
		if err != nil {
			if errs.Has(err.Error()) {
				continue
			}
			allErrs = append(allErrs, err)
			errs.Insert(err.Error())
			continue
		}
		err = psr.Pause(info.Name, info.Namespace)
		if err != nil {
			if errs.Has(err.Error()) {
				continue
			}
			allErrs = append(allErrs, err)
			errs.Insert(err.Error())
		}
		pauseAll := !(o.onlyBackup || o.onlyDb)

		if o.onlyDb || pauseAll {
			fmt.Fprintf(o.Out, "Successfully paused %s/%s.\n", info.Namespace, info.Name)
		}
		if o.onlyBackup || pauseAll {
			fmt.Fprintf(o.Out, "Successfully paused backupconfigurations of %s/%s.\n", info.Namespace, info.Name)
		}
	}

	return utilerrors.NewAggregate(allErrs)
}
