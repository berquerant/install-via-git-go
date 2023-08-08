package cmd

import (
	"berquerant/install-via-git-go/backup"
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
	runCmd.Flags().Bool("clean", false, "Remove lockfile and repo before installation")
	runCmd.Flags().Bool("noupdate", false, "Ignore lock and no update repo, just run scripts")
	runCmd.Flags().String("shell", "bash", "Shell used to run scripts")
	runCmd.MarkFlagsMutuallyExclusive("update", "retry", "clean", "noupdate")
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
	gitCommandName, _ := cmd.Flags().GetString("git")
	workDir, err := getPath(cmd, "workDir")
	if err != nil {
		return errorx.Errorf(err, "invalid workDir")
	}
	gitWorkDir := workDir.Join(config.LocalDir).DirPath()
	env := newEnv(config)
	env.Set("IVG_WORKD", workDir.String())
	gitCommand := git.NewCommand(git.NewCLI(gitWorkDir, env, gitCommandName))
	logx.Info("git", logx.S("git", gitCommandName), logx.S("workDir", gitWorkDir.String()))

	// prepare restore functions
	noupdate, _ := cmd.Flags().GetBool("noupdate")
	clean, _ := cmd.Flags().GetBool("clean")
	lockFile := workDir.Join(config.LockFile).FilePath()
	explicitCommit, _ := cmd.Flags().GetString("commit")
	// restore lockfile when dryrun and given explicit commit or clean install
	restoreLockFile, err := prepareBackupForLockfile(lockFile, explicitCommit, clean)
	if err != nil {
		return errorx.Errorf(err, "backup lockfile")
	}

	// restore repo when clean install and local repo exist
	restoreRepo, err := prepareBackupForRepo(gitWorkDir, clean)
	if err != nil {
		return errorx.Errorf(err, "backup repo")
	}

	dry, _ := cmd.Flags().GetBool("dry")
	restore := func() {
		if !dry {
			return
		}

		logx.Info("restore lockfile", logx.S("path", lockFile.String()))
		if err := restoreLockFile(); err != nil {
			logx.Error("restore lockfile", logx.Err(err))
		}
		logx.Info("restore repo", logx.S("path", gitWorkDir.String()))
		if err := restoreRepo(); err != nil {
			logx.Error("restore repo", logx.Err(err))
		}
	}
	defer restore()

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

	shell, _ := cmd.Flags().GetString("shell")
	logx.Info("start installation!")
	runner := &installRunner{
		config:     config,
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
	config     *Config
	workDir    filepathx.DirPath
	gitWorkDir filepathx.DirPath
	env        execx.Env
	lockFile   filepathx.FilePath
	gitCommand git.Command
	fact       strategy.Fact
	noupdate   bool
	shell      string
}

func (r *installRunner) run(ctx context.Context) error {
	if err := r.workDir.Ensure(); err != nil {
		return errorx.Errorf(err, "ensure workDir")
	}
	if err := r.gitWorkDir.Parent().DirPath().Ensure(); err != nil {
		return errorx.Errorf(err, "ensure git workDir")
	}

	logx.Info("setup")
	if _, err := stringsToExecutor(r.config.Steps.Setup, r.shell).
		Execute(ctx, execx.WithEnv(r.env), execx.WithDir(r.workDir)); err != nil {
		return errorx.Errorf(err, "setup")
	}

	if err := r.lockFile.Ensure(); err != nil {
		return errorx.Errorf(err, "ensure lockfile")
	}

	keeper := gitlock.NewGitKeeper(lock.NewFileKeeper(r.lockFile), r.gitCommand)
	runner := &strategyRunner{
		config:  r.config,
		workDir: r.gitWorkDir,
		env:     r.env,
		runner: r.fact.SelectStrategy().Runner(strategy.NewRunnerConfig(
			r.config.URI,
			r.config.Branch,
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
		config:   r.config,
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
	config   *Config
	keeper   *gitlock.GitKeeper
	workDir  filepathx.DirPath // local repo dir
	env      execx.Env
	noupdate bool
	shell    string
}

func (r *rollbackRunner) run(ctx context.Context) {
	if r.noupdate {
		logx.Info("skip rollback repo and lockfile")
		if _, err := stringsToExecutor(r.config.Steps.Rollback, r.shell).
			Execute(ctx, execx.WithDir(r.workDir), execx.WithEnv(r.env)); err != nil {
			logx.Error("run rollback", logx.Err(err))
		}
		return
	}

	logx.Error("rollback")
	if err := r.keeper.Rollback(ctx); err != nil {
		logx.Error("rollback error", logx.Err(err))
	}
	if _, err := stringsToExecutor(r.config.Steps.Rollback, r.shell).
		Execute(ctx, execx.WithDir(r.workDir), execx.WithEnv(r.env)); err != nil {
		logx.Error("run rollback", logx.Err(err))
	}
}

type strategyRunner struct {
	config  *Config
	workDir filepathx.DirPath // local repo dir
	env     execx.Env
	runner  strategy.Runner
	shell   string
}

func (s *strategyRunner) run(ctx context.Context) error {
	if err := s.runner.Run(ctx); err != nil {
		if !errors.Is(err, strategy.ErrNoopStrategy) {
			return errorx.Errorf(err, "run strategy")
		}

		logx.Info("skip")
		if _, err := stringsToExecutor(s.config.Steps.Skip, s.shell).
			Execute(ctx, execx.WithDir(s.workDir), execx.WithEnv(s.env)); err != nil {
			return errorx.Errorf(err, "run skip")
		}
		return nil
	}

	logx.Info("install")
	if _, err := stringsToExecutor(s.config.Steps.Install, s.shell).
		Execute(ctx, execx.WithDir(s.workDir), execx.WithEnv(s.env)); err != nil {
		return errorx.Errorf(err, "run install")
	}
	return nil
}

func stringsToExecutor(scripts []string, shell string) execx.Executor {
	if len(scripts) == 0 {
		return execx.NewNoopExecutor()
	}
	script := strings.Join(scripts, "\n")
	return execx.NewRawScript("set -ex\n"+script, shell)
}

func noopRestore() error {
	logx.Info("noop restore")
	return nil
}

func prepareBackupForLockfile(lockFile filepathx.FilePath, commit string, clean bool) (func() error, error) {
	if !(commit != "" || clean) {
		// no need to backup
		return noopRestore, nil
	}

	logx.Info("backup lockfile",
		logx.S("path", lockFile.String()),
		logx.S("explicitCommit", commit),
		logx.B("clean", clean),
	)
	// override current commit by explicit commit
	lockFileBackup, err := backup.IntoTempDir(lockFile.Path)
	if err != nil {
		return nil, err
	}
	if err := lockFileBackup.Move(); err != nil {
		return nil, err
	}

	if commit != "" {
		if err := lockFile.Write(commit); err != nil {
			return nil, errorx.Errorf(err, "override commit")
		}
	}
	return func() error {
		return lockFileBackup.Restore()
	}, nil
}

func prepareBackupForRepo(gitWorkDir filepathx.DirPath, clean bool) (func() error, error) {
	if !(clean && gitWorkDir.Exist()) {
		// no need to backup
		return noopRestore, nil
	}

	logx.Info("backup repo", logx.S("path", gitWorkDir.String()))
	repoBackup, err := backup.IntoTempDir(gitWorkDir.Path)
	if err != nil {
		return nil, err
	}
	if err := repoBackup.Rename(); err != nil {
		return nil, err
	}
	return func() error {
		return repoBackup.Restore(backup.WithRename(true))
	}, nil
}
