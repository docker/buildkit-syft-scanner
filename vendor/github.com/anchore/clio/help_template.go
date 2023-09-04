package clio

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/anchore/fangs"
)

var _ postConstruct = updateHelpUsageTemplate

func updateHelpUsageTemplate(a *application) {
	cmd := a.root

	var helpUsageTemplate = fmt.Sprintf(`{{if (or .Long .Short)}}{{.Long}}{{if not .Long}}{{.Short}}{{end}}

{{end}}Usage:{{if (and .Runnable (ne .CommandPath "%s"))}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if .HasExample}}

{{.Example}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{if not .CommandPath}}Global {{end}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if (and .HasAvailableInheritedFlags (ne .CommandPath "%s"))}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{if .CommandPath}}{{.CommandPath}} {{end}}[command] --help" for more information about a command.{{end}}
`, a.setupConfig.ID.Name, a.setupConfig.ID.Name)

	cmd.SetUsageTemplate(helpUsageTemplate)
	cmd.SetHelpTemplate(helpUsageTemplate)
}

var _ postConstruct = showConfigInRootHelp

func showConfigInRootHelp(a *application) {
	cmd := a.root

	helpFn := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// root.Example is set _after all added commands_ because it collects all the
		// options structs in order to output an accurate "config file" summary
		// note: since all commands tend to share help functions it's important to only patch the example
		// when there is no parent command (i.e. the root command).
		if cmd == a.root {
			cfgs := append([]any{&a.state.Config, a}, a.state.Config.FromCommands...)
			for _, cfg := range cfgs {
				// load each config individually, as there may be conflicting names / types that will cause
				// viper to fail to read them all and panic
				if err := fangs.Load(a.setupConfig.FangsConfig, cmd, cfg); err != nil {
					t := reflect.TypeOf(cfg)
					panic(fmt.Sprintf("error loading config object: `%s:%s`: %s", t.PkgPath(), t.Name(), err.Error()))
				}
			}
			summary := a.summarizeConfig(cmd)
			if a.state.RedactStore != nil {
				summary = a.state.RedactStore.RedactString(summary)
			}
			cmd.Example = summary
		}
		helpFn(cmd, args)
	})
}
