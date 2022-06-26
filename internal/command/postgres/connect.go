package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/superfly/flyctl/internal/app"
	"github.com/superfly/flyctl/internal/client"
	"github.com/superfly/flyctl/internal/command"
	"github.com/superfly/flyctl/internal/command/ssh"
	"github.com/superfly/flyctl/internal/flag"
	"github.com/superfly/flyctl/pkg/agent"
)

func newConnect() *cobra.Command {
	const (
		short = "Connect to the Postgres console"
		long  = short + "\n"

		usage = "connect"
	)

	cmd := command.New(usage, short, long, runConnect,
		command.RequireSession,
		command.RequireAppName,
	)

	flag.Add(cmd,
		flag.App(),
		flag.AppConfig(),
		flag.String{
			Name:        "database",
			Shorthand:   "d",
			Description: "The name of the database you would like to connect to",
			Default:     "postgres",
		},
		flag.String{
			Name:        "user",
			Shorthand:   "u",
			Description: "The postgres user to connect with",
			Default:     "postgres",
		},
		flag.String{
			Name:        "password",
			Shorthand:   "p",
			Description: "The postgres user password",
		},
	)

	return cmd
}

func runConnect(ctx context.Context) error {
	var (
		appName = app.NameFromContext(ctx)
		client  = client.FromContext(ctx).API()
	)

	app, err := client.GetAppCompact(ctx, appName)
	if err != nil {
		return fmt.Errorf("failed retrieving app %s: %w", appName, err)
	}

	if !app.IsPostgresApp() {
		return fmt.Errorf("app %s is not a postgres app", appName)
	}

	agentclient, err := agent.Establish(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to establish agent: %w", err)
	}

	dialer, err := agentclient.Dialer(ctx, app.Organization.Slug)
	if err != nil {
		return fmt.Errorf("failed to build tunnel for %s: %v", app.Organization.Slug, err)
	}

	database := flag.GetString(ctx, "database")
	user := flag.GetString(ctx, "user")
	password := flag.GetString(ctx, "password")

	addr := fmt.Sprintf("%s.internal", appName)
	cmdStr := fmt.Sprintf("connect %s %s %s", database, user, password)

	return ssh.SSHConnect(&ssh.SSHParams{
		Ctx:    ctx,
		Org:    app.Organization,
		Dialer: dialer,
		App:    appName,
		Cmd:    cmdStr,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}, addr)
}
