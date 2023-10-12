package cmd

import (
	"berquerant/install-via-git-go/config"
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/exit"
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/logx"
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	rootCmd = &cobra.Command{
		Use:   "install-via-git",
		Short: "Install tools via git.",
		Long:  `install-via-git installs tools via git.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			logx.Setup(debug)
			cmd.SetOut(os.Stdout)
			displayFlags(cmd.Flags())
		},
	}
)

func Execute() error {
	defer func() {
		_ = logx.Sync()
	}()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	defer stop()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logx.Error("execute", logx.Err(err))
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logs")
}

func displayFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(f *pflag.Flag) {
		logx.Debug(
			"flags",
			logx.S("name", f.Name),
			logx.B("changed", f.Changed),
			logx.S("value", f.Value.String()),
		)
	})
}

func setConfigFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("config", "c", "ivg.yml", "Configuration file, - to read from stdin")
	fail(cmd.MarkFlagFilename("config", "yml", "yaml"))
}

func parseConfigFromFlag(cmd *cobra.Command) (*config.Config, error) {
	cfg, _ := cmd.Flags().GetString("config")
	return parseConfigFromOption(cfg)
}

func parseConfigFromOption(opt string) (*config.Config, error) {
	logx.Info("config", logx.S("value", opt))
	if opt == "-" {
		return parseConfigFromStdin()
	}
	return parseConfigFile(opt)
}

func parseConfigFromStdin() (*config.Config, error) {
	cfg, err := config.Parse(os.Stdin)
	if err != nil {
		return nil, errorx.Errorf(err, "load config from stdin")
	}
	return cfg, nil
}

func parseConfigFile(cfgFile string) (*config.Config, error) {
	logx.Debug("parse config", logx.S("path", cfgFile))

	cfg, err := func() (*config.Config, error) {
		path, err := filepathx.NewPath(cfgFile)
		if err != nil {
			return nil, err
		}
		f, err := path.FilePath().Open()
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return config.Parse(f)
	}()
	if err != nil {
		return nil, errorx.Errorf(err, "load config file %s", cfgFile)
	}
	return cfg, nil
}

func getPath(cmd *cobra.Command, name string) (filepathx.Path, error) {
	v, _ := cmd.Flags().GetString(name)
	return filepathx.NewPath(v)
}

func fail(err error) {
	if err == nil {
		return
	}
	logx.Error("fail", logx.Err(err))
	exit.Fail()
}

func setShellFlag(cmd *cobra.Command) {
	cmd.Flags().StringSlice(
		"shell", []string{}, "Shell used to run scripts, separated by comma, e.g. arch,--arm64e,/bin/bash")
}

func getShell(cmd *cobra.Command, cfg *config.Config) []string {
	shell, _ := cmd.Flags().GetStringSlice("shell")
	if len(shell) > 0 {
		return shell
	}
	if len(cfg.Shell) > 0 {
		return cfg.Shell
	}
	return []string{"bash"}
}
