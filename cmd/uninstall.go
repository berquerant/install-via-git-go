package cmd

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/git"
	"berquerant/install-via-git-go/gitlock"
	"berquerant/install-via-git-go/inspect"
	"berquerant/install-via-git-go/lock"
	"berquerant/install-via-git-go/logx"
	"berquerant/install-via-git-go/runner"
	"berquerant/install-via-git-go/strategy"
	"context"

	"github.com/spf13/cobra"
)

func init() {
	setConfigFlag(uninstallCmd)
	setShellFlag(uninstallCmd)
	uninstallCmd.Flags().String("git", "git", "Git command")
	uninstallCmd.Flags().StringP("workDir", "w", ".", "Working directory")
	fail(uninstallCmd.MarkFlagDirname("workDir"))
	uninstallCmd.Flags().Bool("dry", false, "Execute up to strategy determination, no side effects")
	uninstallCmd.Flags().Bool("remove", false, "Remove repo")
	uninstallCmd.Flags().Bool("purge", false, "Remove repo and clear lock")
	uninstallCmd.MarkFlagsMutuallyExclusive("remove", "purge")
	rootCmd.AddCommand(uninstallCmd)
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Run uninstallation",
	Long:  `Uninstall tools according to the configurtion file`,
	RunE:  uninstall,
}

func uninstall(cmd *cobra.Command, _ []string) error {
	// load config
	cfg, err := parseConfigFromFlag(cmd)
	if err != nil {
		return err
	}
	// prepare builtin envs
	env, err := newEnv(cfg, cmd)
	if err != nil {
		return err
	}
	// prepare git cli
	gitCommandName, _ := cmd.Flags().GetString("git")
	workDir, err := getPath(cmd, "workDir")
	if err != nil {
		return errorx.Errorf(err, "invalid workDir")
	}
	gitWorkDir := workDir.Join(cfg.LocalDir).DirPath()
	gitCommand := git.NewCommand(git.NewCLI(gitWorkDir, env, gitCommandName))
	logx.Info("git", logx.S("git", gitCommandName), logx.S("workDir", gitWorkDir.String()))

	lockFile := workDir.Join(cfg.LockFile).FilePath()

	remove, _ := cmd.Flags().GetBool("remove")
	purge, _ := cmd.Flags().GetBool("purge")
	dry, _ := cmd.Flags().GetBool("dry")

	ius := &inspect.UpdateSpec{
		Uninstall: true,
		Remove:    remove || purge,
	}
	fact := strategy.NewFact(
		inspect.RepoExistence(cmd.Context(), gitCommand),
		inspect.LockExistence(lockFile),
		inspect.RepoStatus(cmd.Context(), gitCommand, lockFile),
		ius.Get(),
	)
	logx.Info(
		"strategy",
		logx.B("remove", remove),
		logx.B("purge", purge),
		logx.B("dry", dry),
		logx.S("update_spec", fact.USpec.String()),
		logx.S("type", fact.SelectStrategy().String()),
	)

	if dry {
		return nil
	}

	shell := getShell(cmd, cfg)
	logx.Info("start uninstallation!", logx.SS("shell", shell))
	argument := &runner.Argument{
		Config:       cfg,
		Env:          env,
		Shell:        shell,
		LocalRepoDir: gitWorkDir,
	}
	return (&uninstallRunner{
		Argument:   argument,
		workDir:    workDir.DirPath(),
		lockFile:   lockFile,
		gitCommand: gitCommand,
		fact:       fact,
		purge:      purge,
	}).run(cmd.Context())
}

type uninstallRunner struct {
	*runner.Argument
	workDir    filepathx.DirPath
	lockFile   filepathx.FilePath
	gitCommand git.Command
	fact       strategy.Fact
	purge      bool
}

func (r *uninstallRunner) run(ctx context.Context) error {
	keeper := gitlock.NewGitKeeper(lock.NewFileKeeper(r.lockFile), r.gitCommand)

	logx.Info("run strategy", logx.S("type", r.fact.SelectStrategy().String()))
	if err := runner.NewUninstall(
		r.Argument,
		r.fact.SelectStrategy().Runner(strategy.NewRunnerConfig(
			r.Config.URI,
			r.Config.Branch,
			keeper.Locker().Pair(),
			r.gitCommand,
		)),
	).Run(ctx); err != nil {
		return err
	}

	if r.purge {
		logx.Info("clear lock")
		if err := keeper.Locker().Clear(); err != nil {
			return errorx.Errorf(err, "clear")
		}
	}

	return nil
}
