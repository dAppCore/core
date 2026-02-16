package plugin

import (
	"fmt"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/i18n"
	"forge.lthn.ai/core/cli/pkg/io"
	"forge.lthn.ai/core/cli/pkg/plugin"
)

func addListCommand(parent *cli.Command) {
	listCmd := cli.NewCommand(
		"list",
		i18n.T("List installed plugins"),
		"",
		func(cmd *cli.Command, args []string) error {
			return runList()
		},
	)

	parent.AddCommand(listCmd)
}

func runList() error {
	basePath, err := pluginBasePath()
	if err != nil {
		return err
	}

	registry := plugin.NewRegistry(io.Local, basePath)
	if err := registry.Load(); err != nil {
		return err
	}

	plugins := registry.List()
	if len(plugins) == 0 {
		cli.Dim("No plugins installed")
		return nil
	}

	table := cli.NewTable("Name", "Version", "Source", "Status")
	for _, p := range plugins {
		status := "disabled"
		if p.Enabled {
			status = "enabled"
		}
		table.AddRow(p.Name, p.Version, p.Source, status)
	}

	fmt.Println()
	table.Render()
	fmt.Println()
	cli.Dim(fmt.Sprintf("%d plugin(s) installed", len(plugins)))

	return nil
}
