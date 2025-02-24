package cli

import (
	"context"

	"github.com/epinio/epinio/internal/cli/usercmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CmdAppEnv implements the command: epinio app env
var CmdAppEnv = &cobra.Command{
	Use:   "env",
	Short: "Epinio application configuration",
	Long:  `Manage epinio application environment variables`,
}

func init() {
	CmdAppEnv.AddCommand(CmdEnvList)
	CmdAppEnv.AddCommand(CmdEnvSet)
	CmdAppEnv.AddCommand(CmdEnvShow)
	CmdAppEnv.AddCommand(CmdEnvUnset)
}

// CmdEnvList implements the command: epinio app env list
var CmdEnvList = &cobra.Command{
	Use:               "list APPNAME",
	Short:             "Lists application environment",
	Long:              "Lists environment variables of named application",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: matchingAppsFinder,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New(cmd.Context())
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		err = client.EnvList(cmd.Context(), args[0])
		if err != nil {
			return errors.Wrap(err, "error listing app environment")
		}

		return nil
	},
}

// CmdEnvSet implements the command: epinio app env set
var CmdEnvSet = &cobra.Command{
	Use:   "set APPNAME NAME VALUE",
	Short: "Extend application environment",
	Long:  "Add or change environment variable of named application",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New(cmd.Context())

		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		err = client.EnvSet(cmd.Context(), args[0], args[1], args[2])
		if err != nil {
			return errors.Wrap(err, "error setting into app environment")
		}

		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Ignore name and value of environment variable.
		// EV may exist or not, the command will set or modify.
		if len(args) > 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		app, err := usercmd.New(cmd.Context())
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		app.API.DisableVersionWarning()

		// #args == 0: application name.
		matches := app.AppsMatching(toComplete)

		return matches, cobra.ShellCompDirectiveNoFileComp
	},
}

// CmdEnvShow implements the command: epinio app env show
var CmdEnvShow = &cobra.Command{
	Use:   "show APPNAME NAME",
	Short: "Describe application's environment variable",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New(cmd.Context())

		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		err = client.EnvShow(cmd.Context(), args[0], args[1])
		if err != nil {
			return errors.Wrap(err, "error accessing app environment")
		}

		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 2 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		app, err := usercmd.New(cmd.Context())
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		app.API.DisableVersionWarning()

		if len(args) == 1 {
			// #args == 1: environment variable name (in application)
			matches := app.EnvMatching(context.Background(), args[0], toComplete)
			return matches, cobra.ShellCompDirectiveNoFileComp
		}

		// #args == 0: application name.
		matches := app.AppsMatching(toComplete)

		return matches, cobra.ShellCompDirectiveNoFileComp
	},
}

// CmdEnvUnset implements the command: epinio app env unset
var CmdEnvUnset = &cobra.Command{
	Use:               "unset APPNAME NAME",
	Short:             "Shrink application environment",
	Long:              "Remove environment variable from named application",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: matchingAppsFinder,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New(cmd.Context())
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		err = client.EnvUnset(cmd.Context(), args[0], args[1])
		if err != nil {
			return errors.Wrap(err, "error removing from app environment")
		}

		return nil
	},
}
