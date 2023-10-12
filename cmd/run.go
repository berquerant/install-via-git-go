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
	"errors"

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
	runCmd.MarkFlagsMutuallyExclusive("update", "retry", "clean", "noupdate")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run installation",
	Long:  `Install tools according to the configuration file`,
	RunE:  run,
}

func newEnv(cfg *config.Config) execx.Env {
	env := execx.EnvFromMap(cfg.Env)
	// set builtin envs
	env.Set("IVG_URI", cfg.URI)
	env.Set("IVG_BRANCH", cfg.Branch)
	env.Set("IVG_LOCALD", cfg.LocalDir)
	env.Set("IVG_LOCK", cfg.LockFile)
	return env
}
func run(cmd *cobra.Command, _ []string) error {
	// load config
	cfg, err := parseConfigFromFlag(cmd)
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
	env := newEnv(cfg)
	env.Set("IVG_WORKD", workDir.String())
	gitCommand := git.NewCommand(git.NewCLI(gitWorkDir, env, gitCommandName))
	logx.Info("git", logx.S("git", gitCommandName), logx.S("workDir", gitWorkDir.String()))

	// prepare restore functions
	noupdate, _ := cmd.Flags().GetBool("noupdate")
	clean, _ := cmd.Flags().GetBool("clean")
	lockFile := workDir.Join(cfg.LockFile).FilePath()
	explicitCommit, _ := cmd.Flags().GetString("commit")

	backupList := runner.NewBackupList(
		runner.NewLockFileBackup(lockFile, explicitCommit, clean),
		runner.NewRepoBackup(gitWorkDir, clean),
	)
	if err := backupList.Create(); err != nil {
		return err
	}
	dry, _ := cmd.Flags().GetBool("dry")
	if !dry {
		defer func() {
			if err := backupList.Restore(); err != nil {
				logx.Error("restore backup", logx.Err(err))
			}
		}()
	}

	update, _ := cmd.Flags().GetBool("update")
	retry, _ := cmd.Flags().GetBool("retry")

	// determine strategy
	fact := strategy.NewFact(
		inspect.RepoExistence(cmd.Context(), gitCommand),
		inspect.LockExistence(lockFile),
		inspect.RepoStatus(cmd.Context(), gitCommand, lockFile),
		inspect.UpdateSpec(update, retry, noupdate),
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

	shell := getShell(cmd, cfg)
	logx.Info("start installation!", logx.SS("shell", shell))
	runner := &installRunner{
		cfg:        cfg,
		workDir:    workDir.DirPath(),
		gitWorkDir: gitWorkDir,
		env:        env,
		lockFile:   lockFile,
		gitCommand: gitCommand,
		fact:       fact,
		noupdate:   noupdate,
		shell:      shell,
	}
	return runner.run(cmd.Context())
}

type installRunner struct {
	cfg        *config.Config
	workDir    filepathx.DirPath
	gitWorkDir filepathx.DirPath
	env        execx.Env
	lockFile   filepathx.FilePath
	gitCommand git.Command
	fact       strategy.Fact
	noupdate   bool
	shell      []string
}

func (r *installRunner) run(ctx context.Context) error {
	if err := r.workDir.Ensure(); err != nil {
		return errorx.Errorf(err, "ensure workDir")
	}
	if err := r.gitWorkDir.Parent().DirPath().Ensure(); err != nil {
		return errorx.Errorf(err, "ensure git workDir")
	}

	logx.Info("setup")
	if _, err := execx.NewExecutorFromStrings(r.cfg.Steps.Setup, r.shell...).
		Execute(ctx, execx.WithEnv(r.env), execx.WithDir(r.workDir)); err != nil {
		return errorx.Errorf(err, "setup")
	}

	if err := r.lockFile.Ensure(); err != nil {
		return errorx.Errorf(err, "ensure lockfile")
	}

	keeper := gitlock.NewGitKeeper(lock.NewFileKeeper(r.lockFile), r.gitCommand)
	runner := &strategyRunner{
		cfg:     r.cfg,
		workDir: r.gitWorkDir,
		env:     r.env,
		runner: r.fact.SelectStrategy().Runner(strategy.NewRunnerConfig(
			r.cfg.URI,
			r.cfg.Branch,
			keeper.Locker().Pair(),
			r.gitCommand,
		)),
		shell: r.shell,
	}
	logx.Info("run strategy", logx.S("type", r.fact.SelectStrategy().String()))
	err := runner.run(ctx)
	if err == nil {
		if r.noupdate {
			return nil
		}
		if err := keeper.Commit(); err != nil {
			return errorx.Errorf(err, "commit")
		}
		return nil
	}

	logx.Error("run strategy", logx.Err(err))
	rRunner := &rollbackRunner{
		cfg:      r.cfg,
		keeper:   keeper,
		workDir:  r.gitWorkDir,
		env:      r.env,
		noupdate: r.noupdate,
		shell:    r.shell,
	}
	rRunner.run(ctx)
	return err
}

type rollbackRunner struct {
	cfg      *config.Config
	keeper   *gitlock.GitKeeper
	workDir  filepathx.DirPath // local repo dir
	env      execx.Env
	noupdate bool
	shell    []string
}

func (r *rollbackRunner) run(ctx context.Context) {
	if r.noupdate {
		logx.Info("skip rollback repo and lockfile")
		if _, err := execx.NewExecutorFromStrings(r.cfg.Steps.Rollback, r.shell...).
			Execute(ctx, execx.WithDir(r.workDir), execx.WithEnv(r.env)); err != nil {
			logx.Error("run rollback", logx.Err(err))
		}
		return
	}

	logx.Error("rollback")
	if err := r.keeper.Rollback(ctx); err != nil {
		logx.Error("rollback error", logx.Err(err))
	}
	if _, err := execx.NewExecutorFromStrings(r.cfg.Steps.Rollback, r.shell...).
		Execute(ctx, execx.WithDir(r.workDir), execx.WithEnv(r.env)); err != nil {
		logx.Error("run rollback", logx.Err(err))
	}
}

type strategyRunner struct {
	cfg     *config.Config
	workDir filepathx.DirPath // local repo dir
	env     execx.Env
	runner  strategy.Runner
	shell   []string
}

func (s *strategyRunner) run(ctx context.Context) error {
	if err := s.runner.Run(ctx); err != nil {
		if !errors.Is(err, strategy.ErrNoopStrategy) {
			return errorx.Errorf(err, "run strategy")
		}

		logx.Info("skip")
		if _, err := execx.NewExecutorFromStrings(s.cfg.Steps.Skip, s.shell...).
			Execute(ctx, execx.WithDir(s.workDir), execx.WithEnv(s.env)); err != nil {
			return errorx.Errorf(err, "run skip")
		}
		return nil
	}

	logx.Info("install")
	if _, err := execx.NewExecutorFromStrings(s.cfg.Steps.Install, s.shell...).
		Execute(ctx, execx.WithDir(s.workDir), execx.WithEnv(s.env)); err != nil {
		return errorx.Errorf(err, "run install")
	}
	return nil
}
