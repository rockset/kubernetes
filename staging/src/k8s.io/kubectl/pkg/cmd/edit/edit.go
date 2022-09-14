/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package edit

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/cmd/util/editor"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	utilexec "k8s.io/utils/exec"
)

var (
	editLong = templates.LongDesc(i18n.T(`
		Edit a resource from the default editor.

		The edit command allows you to directly edit any API resource you can retrieve via the
		command line tools. It will open the editor defined by your KUBE_EDITOR, or EDITOR
		environment variables, or fall back to 'vi' for Linux or 'notepad' for Windows.
		You can edit multiple objects, although changes are applied one at a time. The command
		accepts filenames as well as command line arguments, although the files you point to must
		be previously saved versions of resources.

		Editing is done with the API version used to fetch the resource.
		To edit using a specific API version, fully-qualify the resource, version, and group.

		The default format is YAML. To edit in JSON, specify "-o json".

		The flag --windows-line-endings can be used to force Windows line endings,
		otherwise the default for your operating system will be used.

		In the event an error occurs while updating, a temporary file will be created on disk
		that contains your unapplied changes. The most common error when updating a resource
		is another editor changing the resource on the server. When this occurs, you will have
		to apply your changes to the newer version of the resource, or update your temporary
		saved copy to include the latest resource version.`))

	editExample = templates.Examples(i18n.T(`
		# Edit the service named 'docker-registry':
		kubectl edit svc/docker-registry

		# Use an alternative editor
		KUBE_EDITOR="nano" kubectl edit svc/docker-registry

		# Edit the job 'myjob' in JSON using the v1 API format:
		kubectl edit job.v1.batch/myjob -o json

		# Edit the deployment 'mydeployment' in YAML and save the modified config in its annotation:
		kubectl edit deployment/mydeployment -o yaml --save-config`))
)

// NewCmdEdit creates the `edit` command
func NewCmdEdit(f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := editor.NewEditOptions(editor.NormalEditMode, ioStreams)
	o.ValidateOptions = cmdutil.ValidateOptions{EnableValidation: true}

	cmd := &cobra.Command{
		Use:                   "edit (RESOURCE/NAME | -f FILENAME)",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Edit a resource on the server"),
		Long:                  editLong,
		Example:               editExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, args, cmd))
			cmdutil.CheckErr(o.Run())
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			context, _ := cmd.Flags().GetString("context")
			clowntown, _ := cmd.Flags().GetBool("clowntown")
			officeHours, _ := cmd.Flags().GetBool("office-hours")

			if isProdContext(context) && !clowntown {
				return utilexec.CodeExitError{
					Err: fmt.Errorf("why are you editing resources by hand? " +
						"you must use --clowntown if you are REALLY sure you know what you're doing 🤡"),
					Code: 1,
				}
			}

			if clowntown && isProdContext(context) {
				fmt.Fprintf(cmd.OutOrStdout(), "🤡 --clowntown in use, I hope you know what you're doing 🤡\n")
			}

			now := time.Now()
			if pst, err := time.LoadLocation("America/Los_Angeles"); err != nil {
				now = now.In(pst)
			}

			if !isOfficeHours(now) && !officeHours && isProdContext(context) {
				return utilexec.CodeExitError{
					Err: fmt.Errorf("why are you editing resources by hand outside office hours? " +
						"you must use --office-hours if you are REALLY sure you know what you're doing ⏰"),
					Code: 1,
				}

			}

			if officeHours && isProdContext(context) {
				fmt.Fprintf(cmd.OutOrStdout(), "⏰ --office-hours in use, I hope you know what you're doing ⏰\n")
			}

			return nil
		},
	}

	cmd.Flags().Bool("clowntown", false, "required for hand-editing")
	cmd.Flags().Bool("office-hours", false, "required for hand-editing outside office hours")

	// bind flag structs
	o.RecordFlags.AddFlags(cmd)
	o.PrintFlags.AddFlags(cmd)

	usage := "to use to edit the resource"
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
	cmdutil.AddValidateOptionFlags(cmd, &o.ValidateOptions)
	cmd.Flags().BoolVarP(&o.OutputPatch, "output-patch", "", o.OutputPatch, "Output the patch if the resource is edited.")
	cmd.Flags().BoolVar(&o.WindowsLineEndings, "windows-line-endings", o.WindowsLineEndings,
		"Defaults to the line ending native to your platform.")
	cmdutil.AddFieldManagerFlagVar(cmd, &o.FieldManager, "kubectl-edit")
	cmdutil.AddApplyAnnotationVarFlags(cmd, &o.ApplyAnnotation)
	return cmd
}

func isOfficeHours(now time.Time) bool {
	return now.Hour() > 9 && now.Hour() < 18 && now.Weekday() > 0 && now.Weekday() < 6
}

func isProdContext(c string) bool {
	return !strings.HasPrefix(c, "dev-")
}
