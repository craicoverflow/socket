package pluginloader

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

func AddCommands(cmd *cobra.Command) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	b, err := ioutil.ReadFile(path.Join(cwd, "./plugins/git.yaml"))
	if err != nil {
		return err
	}
	var cliPlugin *PluginConfig
	err = yaml.Unmarshal(b, &cliPlugin)
	if err != nil {
		return err
	}

	if &cliPlugin.Commands != nil && len(cliPlugin.Commands) > 0 {
		for _, cfg := range cliPlugin.Commands {
			cmd.AddCommand(addCommand(&cfg))
		}
	}

	return nil
}

func addCommand(cmdCfg *CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:           cmdCfg.Name,
		Short:         cmdCfg.ShortDescription,
		SilenceErrors: true,
		Example: cmdCfg.Examples,
		Args:          cobra.ExactArgs(len(cmdCfg.MapsTo.Args)),
		RunE: func(cmd *cobra.Command, args []string) error {

			args = append([]string{cmdCfg.MapsTo.Subcommand}, args...)
			c := exec.Command(cmdCfg.MapsTo.Name, args...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			var buf bytes.Buffer
			c.Stderr = io.MultiWriter(os.Stderr, &buf)

			return c.Run()
		},
	}

	if cmdCfg.Flags != nil && len(cmdCfg.Flags) > 0 {
		for _, f := range cmdCfg.Flags {
			fs := cmd.Flags()
			addFlag(&f, fs)
		}
	}

	return cmd
}

func addFlag(flagCfg *FlagConfig, fs *pflag.FlagSet) {
	switch flagCfg.Type {
	case "string":
		fs.StringP(flagCfg.Name, flagCfg.Alias, flagCfg.DefaultValue, flagCfg.Description)
	case "bool":
		v, _ := strconv.ParseBool(flagCfg.DefaultValue)
		fs.BoolP(flagCfg.Name, flagCfg.Alias, v, flagCfg.Description)
	case "int":
		v, _ := strconv.Atoi(flagCfg.DefaultValue)
		fs.IntP(flagCfg.Name, flagCfg.Alias, v, flagCfg.Description)
	}
}
