package cmd

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/execx"
	"berquerant/install-via-git-go/exit"
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/logx"
	"context"
	"errors"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		URI      string            `yaml:"uri" json:"uri"`
		Branch   string            `yaml:"branch,omitempty" json:"branch,omitempty"`
		LocalDir string            `yaml:"locald,omitempty" json:"locald,omitempty"`
		LockFile string            `yaml:"lock,omitempty" json:"lock,omitempty"`
		Steps    Steps             `yaml:"steps,inline" json:"steps"`
		Env      map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	}

	Steps struct {
		Setup    []string `yaml:"setup,omitempty" json:"setup,omitempty"`
		Install  []string `yaml:"install,omitempty" json:"install,omitempty"`
		Rollback []string `yaml:"rollback,omitempty" json:"rollback,omitempty"`
		Skip     []string `yaml:"skip,omitempty" json:"skip,omitempty"`
	}
)

func defaultConfig() Config {
	return Config{
		Branch:   "main",
		LockFile: "lock",
		LocalDir: "repo",
	}
}

var (
	rootCmd = &cobra.Command{
		Use:   "install-via-git",
		Short: "Install tools via git.",
		Long:  `install-via-git installs tools via git.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			logger, _ := cmd.Flags().GetString("logger")
			logx.Setup(debug, intoLoggerType(logger))
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
	rootCmd.PersistentFlags().String("logger", "block", "logger type [zap, block]")
}

func intoLoggerType(value string) logx.Type {
	switch value {
	case "zap":
		return logx.Tzap
	default:
		return logx.Tblock
	}
}

func setConfigFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("config", "c", "ivg.yml", "Configuration file")
	fail(cmd.MarkFlagFilename("config", "yml", "yaml"))
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

func parseConfigFile(cfgFile string) (*Config, error) {
	logx.Debug("parse config", logx.S("path", cfgFile))
	path, err := filepathx.NewPath(cfgFile)
	if err != nil {
		return nil, errorx.Errorf(err, "load config file %s", cfgFile)
	}
	content, err := path.FilePath().Read()
	if err != nil {
		return nil, errorx.Errorf(err, "load config file %s", cfgFile)
	}

	config := defaultConfig()
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, errorx.Errorf(err, "parse config file %s", cfgFile)
	}

	if config.URI == "" {
		return nil, errors.New("empty uri")
	}
	return &config, nil
}

func getPath(cmd *cobra.Command, name string) (filepathx.Path, error) {
	v, _ := cmd.Flags().GetString(name)
	return filepathx.NewPath(v)
}

func newEnv(config *Config) execx.Env {
	env := execx.EnvFromMap(config.Env)
	// set builtin envs
	env.Set("IVG_URI", config.URI)
	env.Set("IVG_BRANCH", config.Branch)
	env.Set("IVG_LOCALD", config.LocalDir)
	env.Set("IVG_LOCK", config.LockFile)
	return env
}

func fail(err error) {
	if err == nil {
		return
	}
	logx.Error("fail", logx.Err(err))
	exit.Fail()
}
