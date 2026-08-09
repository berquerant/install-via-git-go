package runner

import (
	"berquerant/install-via-git-go/backup"
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/logx"
)

type Backuper interface {
	Create() error
	Restore() error
	// Close removes the backup. Should be called after Create even on success.
	Close() error
}

type BackupList []Backuper

func NewBackupList(backuper ...Backuper) BackupList {
	return BackupList(backuper)
}

func (b BackupList) Create() error {
	return errorx.Serial(b, func(x Backuper) error {
		return x.Create()
	})
}

func (b BackupList) Restore() error {
	return errorx.Serial(b, func(x Backuper) error {
		return x.Restore()
	})
}

func (b BackupList) Close() error {
	return errorx.Serial(b, func(x Backuper) error {
		return x.Close()
	})
}

type noopBackup struct{}

func (*noopBackup) Create() error  { return nil }
func (*noopBackup) Restore() error { return nil }
func (*noopBackup) Close() error   { return nil }

type LockFileBackup struct {
	backupFile *backup.Backup
	origin     filepathx.FilePath
	commit     string
}

func NewLockFileBackup(origin filepathx.FilePath, commit string, clean bool) Backuper {
	if !(clean || commit != "") {
		return &noopBackup{}
	}
	return &LockFileBackup{
		origin: origin,
		commit: commit,
	}
}

func (b *LockFileBackup) Create() error {
	logx.Info("backup lockfile",
		logx.S("path", b.origin.String()),
		logx.S("explicitCommit", b.commit),
	)
	// override current commit by explicit commit
	backupFile, err := backup.IntoTempDir(b.origin.Path)
	if err != nil {
		return errorx.Errorf(err, "create backup")
	}
	if err := backupFile.Copy(); err != nil {
		return errorx.Errorf(err, "move backup")
	}

	if err := b.origin.Write(b.commit); err != nil {
		return errorx.Errorf(err, "override commit")
	}
	b.backupFile = backupFile
	return nil
}

func (b *LockFileBackup) Restore() error {
	if b.backupFile == nil {
		return nil
	}
	defer b.backupFile.Close()
	return b.backupFile.Restore()
}

func (b *LockFileBackup) Close() error {
	if b.backupFile == nil {
		return nil
	}
	return b.backupFile.Close()
}

type RepoBackup struct {
	backupDir  *backup.Backup
	gitWorkDir filepathx.DirPath
}

func NewRepoBackup(gitWorkDir filepathx.DirPath, clean bool) Backuper {
	if !(clean || gitWorkDir.Exist()) {
		return &noopBackup{}
	}
	return &RepoBackup{
		gitWorkDir: gitWorkDir,
	}
}

func (b *RepoBackup) Create() error {
	logx.Info("backup repo", logx.S("path", b.gitWorkDir.String()))
	repoBackup, err := backup.IntoTempDir(b.gitWorkDir.Path)
	if err != nil {
		return errorx.Errorf(err, "create backup")
	}
	if err := repoBackup.Copy(); err != nil {
		return errorx.Errorf(err, "move backup")
	}
	b.backupDir = repoBackup
	return nil
}

func (b *RepoBackup) Restore() error {
	if b.backupDir == nil {
		return nil
	}
	defer b.backupDir.Close()
	return b.backupDir.Restore()
}

func (b *RepoBackup) Close() error {
	if b.backupDir == nil {
		return nil
	}
	return b.backupDir.Close()
}
