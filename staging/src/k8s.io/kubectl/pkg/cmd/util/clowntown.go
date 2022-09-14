package util

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	utilexec "k8s.io/utils/exec"
	"os"
	"strings"
	"time"
)

func AddClowntownFlags(cmd *cobra.Command) {
	cmd.Flags().String("reason", "", "optional reason for doing clowny stuff 🤡")
	cmd.Flags().Bool("clowntown", false, "required for hand-editing 🤡")
}

func Clowntown(cmd *cobra.Command, args []string) error {
	var err error

	// if rsctl invokes kubectl we want to bypass the clowntown check
	if _, found := os.LookupEnv("RSCTL_BYPASS"); found {
		return nil
	}

	cluster, _ := cmd.Flags().GetString("context")
	clowntown, _ := cmd.Flags().GetBool("clowntown")
	reason, _ := cmd.Flags().GetString("reason")

	if cluster == "" {
		if cluster, err = currentKubeConfigContext(); err != nil {
			return utilexec.CodeExitError{Err: err, Code: 1}
		}
	}

	if isProdContext(cluster) && !clowntown {
		return utilexec.CodeExitError{
			Err: fmt.Errorf("why are you editing resources by hand? " +
				"you must use --clowntown if you are REALLY sure you know what you're doing 🤡"),
			Code: 1,
		}
	}

	if clowntown && isProdContext(cluster) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🤡 --clowntown in use, I hope you know what you're doing 🤡\n")
		sendSlackMessage(reason, cluster, os.Args)
	}

	return nil
}

func currentKubeConfigContext() (string, error) {
	configAccess := clientcmd.NewDefaultPathOptions()
	cfg, err := configAccess.GetStartingConfig()
	if err != nil {
		return "", err
	}

	return cfg.CurrentContext, nil
}

func isProdContext(c string) bool {
	return !strings.HasPrefix(c, "dev-")
}

func sendSlackMessage(reason, cluster string, args []string) {
	// try sending a Slack message for 3 seconds, then give up
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-west-2"),
		config.WithSharedConfigProfile("prod"))
	if err != nil {
		return
	}

	who := os.Getenv("USER")
	if who == "" {
		who = fmt.Sprintf("unknown (%d)", os.Getuid())
	}

	foo := sts.NewFromConfig(cfg)
	res, err := foo.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err == nil {
		fields := strings.Split(*res.UserId, ":")
		if len(fields) > 1 {
			who = fields[len(fields)-1]
		}
	}

	svc := ssm.NewFromConfig(cfg)
	param, err := svc.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String("/r7/slack/token"),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return
	}

	api := slack.New(*param.Parameter.Value)
	msg := fmt.Sprintf("🤡 %s ran `%s` in `%s`",
		who, strings.Join(args, " "), cluster)
	if reason != "" {
		msg = fmt.Sprintf("%s\nreason: %s", msg, reason)
	}
	if _, _, _, err = api.SendMessageContext(ctx, "C754JHX2S", slack.MsgOptionText(
		msg, false)); err != nil {
	}
}
