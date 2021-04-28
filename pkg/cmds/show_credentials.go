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

	"kubedb.dev/cli/pkg/credentials"

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
	showCredLong = templates.LongDesc(`
		Restart the pods of the database smartly.
		This command creates an Ops Request that will
		restart the database pods.
    `)

	showCredExample = templates.Examples(`
		# Restart a mongodb database
		dba restart mongodb mg-demo

		# Restart all mongodb databases
		dba restart mongodb

		# Restart a postgres database
		dba restart pg/postgres-demo

 		Valid resource types include:
    		* elasticsearch
			* mongodb
			* mariadb
			* mysql
			* postgres
			* redis
`)
)

type ShowCredentialsOptions struct {
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
}

func NewCmdShowCredentials(parent string, f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := &ShowCredentialsOptions{
		FilenameOptions: &resource.FilenameOptions{},
		Factory:         f,
		CmdParent:       parent,

		IOStreams: streams,
	}

	cmd := &cobra.Command{
		Use:     "show-credentials (-f FILENAME | TYPE [NAME_PREFIX | -l label] | TYPE/NAME)",
		Short:   i18n.T("Smartly restart the pods of the database."),
		Long:    showCredLong + "\n\n" + cmdutil.SuggestAPIResources("kubectl"),
		Example: showCredExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate(args))
			cmdutil.CheckErr(o.Run())
		},
		DisableFlagsInUseLine: true,
		DisableAutoGenTag:     true,
	}
	usage := "containing the databases to show credentials"
	cmdutil.AddFilenameOptionFlags(cmd, o.FilenameOptions, usage)
	cmd.Flags().StringVarP(&o.Selector, "selector", "l", o.Selector, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().BoolVar(&o.AllNamespaces, "all-namespaces", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")

	return cmd
}

func (o *ShowCredentialsOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
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

func (o *ShowCredentialsOptions) Validate(args []string) error {
	return nil
}

func (o *ShowCredentialsOptions) Run() error {
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
		showCred, err := credentials.NewShowCredentials(o.Factory, info.Mapping)
		if err != nil {
			if errs.Has(err.Error()) {
				continue
			}
			allErrs = append(allErrs, err)
			errs.Insert(err.Error())
			continue
		}
		data, err := showCred.GetCred(info.Name, info.Namespace)
		if err != nil {
			if errs.Has(err.Error()) {
				continue
			}
			allErrs = append(allErrs, err)
			errs.Insert(err.Error())
		} else {
			fmt.Fprintf(o.Out, "Credentials for database %s/%s:\n", info.Namespace, info.Name)
			for key, val := range data {
				fmt.Fprintf(o.Out, "%s = %s\n", key, val)
			}
			fmt.Fprintf(o.Out, "\n")
		}
	}

	return utilerrors.NewAggregate(allErrs)
}
