package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(skeletonCmd)
}

var skeletonCmd = &cobra.Command{
	Use:   "skeleton",
	Short: "Generate config skeleton",
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Println(skeleton)
	},
}

const skeleton = `# install-via-git configuration.
#
# install-via-git executes the process in the following order.
#
# 1. check the lockfile, the local repo
# 2. determine the update strategy
# 3. execute check
# 4. execute setup
# 5. manage the local repository
# 6. execute skip and exit if no update required
# 7. execute install
#
# If an error occurs after 5, rollback the lockfile, the local repo, execute rollback and exit.
# The update strategies are below:
#
# - Tunknown: no proper strategy.
# - TinitFromEmpty: clone repo, create lock
# - TinitFromEmptyToLock: clone repo, checkout to lock
# - TinitFromEmptyToLatest: clone repo, update lock
# - TcreateLock: create lock
# - TcreateLatestLock: pull latest, create lock
# - TupdateToLock: checkout to lock
# - TupdateToLatestWithLock: pull latest, update lock
# - Tnoop: no operation for repo, no update required
# - Tretry: no operation for repo, but continue installation
# - Tnoupdate: no operation for repo and lock, but continue installation
#
# The strategy depends on the following factors:
# - local repo existence
# - lock existence
# - lock and repo status
# - "update" cli option
# - "retry" cli option
#
# repository uri
uri: https://github.com/some/toolname.git
# target branch name (optional, default is main)
branch: master
# git clone destination (optional, default is repo).
# clone to workDir/locald.
locald: localrepo
# file to store commit hash (optional, default is lock).
# empty file is assumed to not exist
lock: lockfile
# shell to execute scripts (setup, install, ...) (optional).
# command line "--shell" overrides this.
shell:
  - /bin/bash
# environment variables (optional).
# check, setup, install, rollback, skip can refer the following variables:
# - IVG_URI=value of repository
# - IVG_BRANCH=value of branch
# - IVG_LOCALD=value of locald
# - IVG_LOCK=value of lock
# install can refer the following variables:
# - IVG_WORKD=absolute path of workDir
env:
  MY_NAME: myname
# check will always run in workDir (optional)
# cancel installation when returning a failure exit status
check:
  - echo "Start check"
# setup will always run in workDir (optional)
setup:
  - echo "Start setup"
# install will run when installation is required in workDir/locald (optional)
install:
  - echo "Start install"
# rollback will run when an error occurs in workDir/locald (optional)
rollback:
  - echo "Start rollback"
# skip will run when no update is required in workDir/locald (optional)
skip:
  - echo "Start skip"`
