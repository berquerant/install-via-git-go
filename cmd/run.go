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
	// determine strategy
	noupdate, _ := cmd.Flags().GetBool("noupdate")
	clean, _ := cmd.Flags().GetBool("clean")
	lockFile := workDir.Join(cfg.LockFile).FilePath()
	dry, _ := cmd.Flags().GetBool("dry")
	update, _ := cmd.Flags().GetBool("update")
	retry, _ := cmd.Flags().GetBool("retry")
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

	explicitCommit, _ := cmd.Flags().GetString("commit")
	backupList := runner.NewBackupList(
		runner.NewLockFileBackup(lockFile, explicitCommit, clean),
		runner.NewRepoBackup(gitWorkDir, clean),
	)
	if err := backupList.Create(); err != nil {
		return errorx.Errorf(err, "create backup")
	}

	shell := getShell(cmd, cfg)
	logx.Info("start installation!", logx.SS("shell", shell))
	installErr := (&installRunner{
		cfg:        cfg,
		workDir:    workDir.DirPath(),
		gitWorkDir: gitWorkDir,
		env:        env,
		lockFile:   lockFile,
		gitCommand: gitCommand,
		fact:       fact,
		noupdate:   noupdate,
		shell:      shell,
	}).run(cmd.Context())
	if installErr != nil {
		if err := backupList.Restore(); err != nil {
			logx.Error("restore backup", logx.Err(err))
		}
	}
	return installErr
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

	logx.Info("run strategy", logx.S("type", r.fact.SelectStrategy().String()))
	err := runner.NewStrategy(
		r.fact.SelectStrategy().Runner(strategy.NewRunnerConfig(
			r.cfg.URI,
			r.cfg.Branch,
			keeper.Locker().Pair(),
			r.gitCommand,
		)),
		r.cfg,
		r.gitWorkDir,
		r.env,
		r.shell,
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
	runner.NewRollback(
		keeper,
		r.noupdate,
		r.cfg,
		r.gitWorkDir,
		r.env,
		r.shell,
	).Run(ctx)
	return err
}
