package cli

import (
	"github.com/epinio/epinio/internal/cli/usercmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CmdAppChart implements the command: epinio app chart
var CmdAppChart = &cobra.Command{
	Use:   "chart",
	Short: "Epinio application chart management",
	Long:  `Manage epinio application charts`,
}

func init() {
	CmdAppChart.AddCommand(CmdAppChartList)
	CmdAppChart.AddCommand(CmdAppChartShow)
	CmdAppChart.AddCommand(CmdAppChartDefault)
}

// CmdAppChartDefault implements the command: epinio app chart default
var CmdAppChartDefault = &cobra.Command{
	Use:               "default [CHARTNAME]",
	Short:             "Set or show app chart default",
	Long:              "Set or show app chart default",
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: matchingChartFinder,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New(cmd.Context())
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		if len(args) == 1 {
			err = client.ChartDefaultSet(cmd.Context(), args[0])
			if err != nil {
				return errors.Wrap(err, "error setting app chart default")
			}
		} else {
			err = client.ChartDefaultShow(cmd.Context())
			if err != nil {
				return errors.Wrap(err, "error showing app chart default")
			}
		}

		return nil
	},
}

// CmdAppChartList implements the command: epinio app chart list
var CmdAppChartList = &cobra.Command{
	Use:   "list",
	Short: "List application charts",
	Long:  "List applications charts",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New(cmd.Context())
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		err = client.ChartList(cmd.Context())
		if err != nil {
			return errors.Wrap(err, "error listing app charts")
		}

		return nil
	},
}

// CmdAppChartShow implements the command: epinio app env show
var CmdAppChartShow = &cobra.Command{
	Use:               "show CHARTNAME",
	Short:             "Describe application chart",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: matchingChartFinder,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		client, err := usercmd.New(cmd.Context())
		if err != nil {
			return errors.Wrap(err, "error initializing cli")
		}

		err = client.ChartShow(cmd.Context(), args[0])
		if err != nil {
			return errors.Wrap(err, "error showing app chart")
		}

		return nil
	},
}
