# install-via-git-go

```
❯ install-via-git help
install-via-git installs tools via git.

Usage:
  install-via-git [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  parse       Parse config file
  run         Run installation
  skeleton    Generate config skeleton
  uninstall   Run uninstallation

Flags:
      --debug   Enable debug logs
  -h, --help    help for install-via-git

Use "install-via-git [command] --help" for more information about a command.
```

## Configuration

```
❯ install-via-git skeleton
# install-via-git configuration.
#
# install-via-git executes the process in the following order.
#
# 1. check the lockfile, the local repository
# 2. determine the update strategy
# 3. execute check
# 4. execute setup
# 5. manage the local repository
# 6. execute skip and exit if no update required
# 7. execute install
#
# If an error occurs after 5, rollback the lockfile, the local repository, execute rollback and exit.
#
# uninstall subcommand executes the following instead.
#
# 1. execute uninstall
# 2. remove the local repository if --remove or --purge flag is specified
# 3. clear the lockfile if --purge flag is specified
#
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
# - Tremove: remove repo
#
# The strategy depends on the following factors:
# - local repo existence
# - lock existence
# - lock and repo status
# - "update" cli option
# - "retry" cli option
# - "uninstall" cli subcommand
# - "remove" cli option
# - "purge" cli option
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
  - echo "Start skip"
# uninstall will run by uninstall subcommand (optional)
uninstall:
  - echo "Start uninstall"
```
