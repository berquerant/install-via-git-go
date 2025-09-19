package cmd

import (
	"berquerant/install-via-git-go/config"
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/execx"
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
	setConfigFlag(runCmd)
	setShellFlag(runCmd)
	runCmd.Flags().String("git", "git", "Git command")
	runCmd.Flags().StringP("workDir", "w", ".", "Working directory")
	fail(runCmd.MarkFlagDirname("workDir"))
	runCmd.Flags().BoolP("update", "u", false, "Force update")
	runCmd.Flags().BoolP("retry", "r", false, "Continue even if no update")
	runCmd.Flags().Bool("dry", false, "Execute up to strategy determination, no side effects")
	runCmd.Flags().String("commit", "", "Fix commit hash")
	runCmd.Flags().Bool("clean", false, "Remove lockfile and repo before installation")
	runCmd.Flags().Bool("noupdate", false, "Ignore lock and no update repo, just run scripts")
	runCmd.Flags().Bool("backupRepo", false, "Backup repo dir")
	runCmd.MarkFlagsMutuallyExclusive("update", "retry", "clean", "noupdate")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run installation",
	Long:  `Install tools according to the configuration file`,
	RunE:  run,
}

func newEnv(cfg *config.Config, cmd *cobra.Command) (execx.Env, error) {
	env := execx.EnvFromMap(cfg.Env)
	env.Set("IVG_URI", cfg.URI)
	env.Set("IVG_BRANCH", cfg.Branch)
	env.Set("IVG_LOCALD", cfg.LocalDir)
	env.Set("IVG_LOCK", cfg.LockFile)
	workDir, err := getPath(cmd, "workDir")
	if err != nil {
		return nil, errorx.Errorf(err, "invalid workDir")
	}
	env.Set("IVG_WORKD", workDir.String())
	return env, nil
}

func run(cmd *cobra.Command, _ []string) error {
	common, err := prepareCommonResource(cmd)
	if err != nil {
		return err
	}
	// determine strategy
	noupdate, _ := cmd.Flags().GetBool("noupdate")
	clean, _ := cmd.Flags().GetBool("clean")
	lockFile := common.lockFile()
	dry, _ := cmd.Flags().GetBool("dry")
	update, _ := cmd.Flags().GetBool("update")
	retry, _ := cmd.Flags().GetBool("retry")
	ius := &inspect.UpdateSpec{
		Update:   update,
		Retry:    retry,
		NoUpdate: noupdate,
	}
	fact := strategy.NewFact(
		inspect.RepoExistence(cmd.Context(), common.gitCommand),
		inspect.LockExistence(lockFile),
		inspect.RepoStatus(cmd.Context(), common.gitCommand, lockFile),
		ius.Get(),
	)
	// check hashes
	{
		commit, err := common.gitCommand.GetCommitHash(cmd.Context())
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
		logx.B("clean", clean),
		logx.B("dry", dry),
		logx.S("repo_exist", fact.RExist.String()),
		logx.S("lock_exist", fact.LExist.String()),
		logx.S("repo_status", fact.RStatus.String()),
		logx.S("update_spec", fact.USpec.String()),
		logx.S("type", fact.SelectStrategy().String()),
	)

	if dry {
		return nil
	}

	explicitCommit, _ := cmd.Flags().GetString("commit")
	backuperList := []runner.Backuper{
		runner.NewLockFileBackup(lockFile, explicitCommit, clean),
	}
	if backupRepo, _ := cmd.Flags().GetBool("backupRepo"); backupRepo {
		backuperList = append(backuperList, runner.NewRepoBackup(common.gitCommand.CLI().Dir(), clean))
	}
	backupList := runner.NewBackupList(backuperList...)
	if err := backupList.Create(); err != nil {
		return errorx.Errorf(err, "create backup")
	}

	shell := getShell(cmd, common.cfg)
	logx.Info("start installation!", logx.SS("shell", shell))
	argument := &runner.Argument{
		Config:       common.cfg,
		Env:          common.env,
		Shell:        shell,
		LocalRepoDir: common.gitCommand.CLI().Dir(),
	}
	installErr := (&installRunner{
		Argument:   argument,
		workDir:    common.workDir.DirPath(),
		lockFile:   lockFile,
		gitCommand: common.gitCommand,
		fact:       fact,
		noupdate:   noupdate,
	}).run(cmd.Context())
	if installErr != nil {
		if err := backupList.Restore(); err != nil {
			logx.Error("restore backup", logx.Err(err))
		}
	}
	return installErr
}

type installRunner struct {
	*runner.Argument
	workDir    filepathx.DirPath
	lockFile   filepathx.FilePath
	gitCommand git.Command
	fact       strategy.Fact
	noupdate   bool
}

func (r *installRunner) run(ctx context.Context) error {
	if err := r.workDir.Ensure(); err != nil {
		return errorx.Errorf(err, "ensure workDir")
	}
	if err := r.LocalRepoDir.Parent().DirPath().Ensure(); err != nil {
		return errorx.Errorf(err, "ensure git workDir")
	}

	logx.Info("check")
	if _, err := execx.NewExecutorFromStrings(r.Config.Steps.Check, r.Shell...).
		Execute(ctx, execx.WithEnv(r.Env), execx.WithDir(r.workDir)); err != nil {
		logx.Info("cancel installation because check failed", logx.Err(err))
		return nil
	}

	logx.Info("setup")
	if _, err := execx.NewExecutorFromStrings(r.Config.Steps.Setup, r.Shell...).
		Execute(ctx, execx.WithEnv(r.Env), execx.WithDir(r.workDir)); err != nil {
		return errorx.Errorf(err, "setup")
	}

	if err := r.lockFile.Ensure(); err != nil {
		return errorx.Errorf(err, "ensure lockfile")
	}

	keeper := gitlock.NewGitKeeper(lock.NewFileKeeper(r.lockFile), r.gitCommand)

	logx.Info("run strategy", logx.S("type", r.fact.SelectStrategy().String()))
	err := runner.NewStrategy(
		r.Argument,
		r.fact.SelectStrategy().Runner(strategy.NewRunnerConfig(
			r.Config.URI,
			r.Config.Branch,
			keeper.Locker().Pair(),
			r.gitCommand,
		)),
	).Run(ctx)

	if err == nil {
		if r.noupdate {
			return nil
		}
		if err := keeper.Commit(); err != nil {
			return errorx.Errorf(err, "commit")
		}
		return nil
	}

	// failed to run strategy
	logx.Error("run strategy", logx.Err(err))
	_ = runner.NewRollback(
		r.Argument,
		keeper,
		r.noupdate,
	).Run(ctx)
	return err
}
