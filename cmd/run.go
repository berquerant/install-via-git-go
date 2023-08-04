package cmd

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/execx"
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/git"
	"berquerant/install-via-git-go/gitlock"
	"berquerant/install-via-git-go/inspect"
	"berquerant/install-via-git-go/lock"
	"berquerant/install-via-git-go/logx"
	"berquerant/install-via-git-go/strategy"
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	setConfigFlag(runCmd)
	runCmd.Flags().String("git", "git", "Git command")
	runCmd.Flags().StringP("workDir", "w", ".", "Working directory")
	fail(runCmd.MarkFlagDirname("workDir"))
	runCmd.Flags().BoolP("update", "u", false, "Force update")
	runCmd.Flags().BoolP("retry", "r", false, "Continue even if no update")
	runCmd.Flags().Bool("dry", false, "Execute up to strategy determination, no side effects")
	runCmd.Flags().String("commit", "", "Fix commit hash")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run installation",
	Long:  `Install tools according to the configuration file`,
	RunE:  run,
}

func run(cmd *cobra.Command, _ []string) error {
	// load config
	config, err := parseConfigFromFlag(cmd)
	if err != nil {
		return err
	}

	// prepare git cli
	env := newEnv(config)
	gitCommandName, _ := cmd.Flags().GetString("git")
	workDir, err := getPath(cmd, "workDir")
	if err != nil {
		return errorx.Errorf(err, "invalid workDir")
	}
	env.Set("IVG_WORKD", workDir.String())

	gitWorkDir := workDir.Join(config.LocalDir).DirPath()
	gitCommand := git.NewCommand(git.NewCLI(gitWorkDir, env, gitCommandName))
	logx.Info("git", logx.S("git", gitCommandName), logx.S("workDir", gitWorkDir.String()))

	// determine strategy
	explicitCommit, _ := cmd.Flags().GetString("commit")
	lockFile := workDir.Join(config.LockFile).FilePath()
	originCommit, _ := lockFile.Read()
	if explicitCommit != "" {
		// override current commit
		if err := lockFile.Write(explicitCommit); err != nil {
			return errorx.Errorf(err, "override commit")
		}
	}
	rollbackToOriginCommit := func() error {
		return lockFile.Write(originCommit)
	}

	update, _ := cmd.Flags().GetBool("update")
	retry, _ := cmd.Flags().GetBool("retry")

	fact := strategy.NewFact(
		inspect.RepoExistence(cmd.Context(), gitCommand),
		inspect.LockExistence(lockFile),
		inspect.RepoStatus(cmd.Context(), gitCommand, lockFile),
		inspect.UpdateSpec(update, retry),
	)
	// check hashes
	{
		commit, err := gitCommand.GetCommitHash(cmd.Context())
		logx.Info("current hash", logx.S("hash", commit), logx.Err(err))
	}
	{
		commit, err := lockFile.Read()
		logx.Info("lock hash", logx.S("hash", commit), logx.Err(err))
	}
	logx.Info(
		"strategy",
		logx.S("lock", lockFile.String()),
		logx.B("update", update),
		logx.B("retry", retry),
		logx.S("repo_exist", fact.RExist.String()),
		logx.S("lock_exist", fact.LExist.String()),
		logx.S("repo_status", fact.RStatus.String()),
		logx.S("update_spec", fact.USpec.String()),
		logx.S("type", fact.SelectStrategy().String()),
	)

	if dry, _ := cmd.Flags().GetBool("dry"); dry {
		return rollbackToOriginCommit()
	}

	logx.Info("start installation!")
	if err := workDir.DirPath().Ensure(); err != nil {
		return errorx.Errorf(err, "ensure workDir")
	}
	if err := gitWorkDir.Parent().DirPath().Ensure(); err != nil {
		return errorx.Errorf(err, "ensure git workDir")
	}

	logx.Info("setup")
	if _, err := stringsToExecutor(config.Steps.Setup).
		Execute(cmd.Context(), execx.WithEnv(env), execx.WithDir(workDir.DirPath())); err != nil {
		return errorx.Errorf(err, "setup")
	}

	if err := lockFile.Ensure(); err != nil {
		return errorx.Errorf(err, "ensure lockfile")
	}

	keeper := gitlock.NewGitKeeper(lock.NewFileKeeper(lockFile), gitCommand)
	runner := &strategyRunner{
		config:  config,
		workDir: gitWorkDir,
		env:     env,
		runner: fact.SelectStrategy().Runner(strategy.NewRunnerConfig(
			config.URI,
			config.Branch,
			keeper.Locker().Pair(),
			gitCommand,
		)),
	}
	logx.Info("run strategy", logx.S("type", fact.SelectStrategy().String()))
	if err := runner.run(cmd.Context()); err != nil {
		logx.Error("rollback", logx.Err(err))
		if err := keeper.Rollback(cmd.Context()); err != nil {
			logx.Error("rollback error", logx.Err(err))
		}
		return err
	}
	if err := keeper.Commit(); err != nil {
		return errorx.Errorf(err, "commit")
	}
	return nil
}

type strategyRunner struct {
	config  *Config
	workDir filepathx.DirPath
	env     execx.Env
	runner  strategy.Runner
}

func (s *strategyRunner) run(ctx context.Context) error {
	if err := s.runner.Run(ctx); err != nil {
		if !errors.Is(err, strategy.ErrNoopStrategy) {
			return errorx.Errorf(err, "run strategy")
		}

		logx.Info("skip")
		if _, err := stringsToExecutor(s.config.Steps.Skip).
			Execute(ctx, execx.WithDir(s.workDir), execx.WithEnv(s.env)); err != nil {
			return errorx.Errorf(err, "run skip")
		}
		return nil
	}

	logx.Info("install")
	if _, err := stringsToExecutor(s.config.Steps.Install).
		Execute(ctx, execx.WithDir(s.workDir), execx.WithEnv(s.env)); err != nil {
		return errorx.Errorf(err, "run install")
	}
	return nil
}

func stringsToExecutor(scripts []string) execx.Executor {
	if len(scripts) == 0 {
		return execx.NewNoopExecutor()
	}
	script := strings.Join(scripts, "\n")
	return execx.NewRawScript("set -ex\n" + script)
}
