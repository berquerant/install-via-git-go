package backup_test

import (
	"berquerant/install-via-git-go/backup"
	"berquerant/install-via-git-go/filepathx"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackup(t *testing.T) {
	baseDir := t.TempDir()
	base, err := filepathx.NewPath(baseDir)
	assert.Nil(t, err)
	basePath := base.DirPath()

	// srcdir/
	//   file1
	//   dir1/
	//     file2
	//     dir2/
	t.Run("dir", func(t *testing.T) {
		var (
			srcDir = basePath.Join("srcdir").DirPath()
			file1  = srcDir.Join("file1").FilePath()
			dir1   = srcDir.Join("dir1").DirPath()
			file2  = dir1.Join("file2").FilePath()
			dir2   = dir1.Join("dir2").DirPath()
		)

		prepare := func(t *testing.T) {
			assert.Nil(t, srcDir.Remove())
			assert.Nil(t, srcDir.Ensure())
			assert.Nil(t, file1.Write("file1"))
			assert.Nil(t, dir1.Ensure())
			assert.Nil(t, file2.Write("file2"))
			assert.Nil(t, dir2.Ensure())
		}
		check := func(t *testing.T) {
			assert.True(t, dir2.Exist())
			assert.True(t, dir1.Exist())
			{
				got, err := file2.Read()
				assert.Nil(t, err)
				assert.Equal(t, "file2", got)
			}
			{
				got, err := file1.Read()
				assert.Nil(t, err)
				assert.Equal(t, "file1", got)
			}
		}
		prepareBackup := func(t *testing.T) *backup.Backup {
			prepare(t)
			b, err := backup.New(basePath.Join("dirbackup").DirPath(), srcDir.Path)
			assert.Nil(t, err)
			return b
		}

		t.Run("remove all", func(t *testing.T) {
			b := prepareBackup(t)
			defer b.Close()
			assert.Nil(t, b.Copy())
			assert.Nil(t, srcDir.Remove())
			assert.Nil(t, b.Restore())
			check(t)
		})

		t.Run("remove file", func(t *testing.T) {
			b := prepareBackup(t)
			defer b.Close()
			assert.Nil(t, b.Copy())
			assert.Nil(t, file1.Remove())
			assert.Nil(t, b.Restore())
			check(t)
		})

		t.Run("remove dir", func(t *testing.T) {
			b := prepareBackup(t)
			defer b.Close()
			assert.Nil(t, b.Copy())
			assert.Nil(t, dir1.Remove())
			assert.Nil(t, b.Restore())
			check(t)
		})

		t.Run("modify file1", func(t *testing.T) {
			b := prepareBackup(t)
			defer b.Close()
			assert.Nil(t, b.Copy())
			assert.Nil(t, file1.Write("modify"))
			assert.Nil(t, b.Restore())
			check(t)
		})

		t.Run("modify file2", func(t *testing.T) {
			b := prepareBackup(t)
			defer b.Close()
			assert.Nil(t, b.Copy())
			assert.Nil(t, file2.Write("modify"))
			assert.Nil(t, b.Restore())
			check(t)
		})
	})

	t.Run("file", func(t *testing.T) {
		const content = "zone"
		origin := basePath.Join("src.file")

		assert.Nil(t, origin.FilePath().Write(content))
		b, err := backup.New(basePath.Join("filebackup").DirPath(), origin)
		if !assert.Nil(t, err) {
			return
		}
		defer b.Close()

		t.Run("modify", func(t *testing.T) {
			assert.Nil(t, b.Copy())
			assert.Nil(t, origin.FilePath().Write("enoz"))
			assert.Nil(t, b.Restore())
			got, err := origin.FilePath().Read()
			assert.Nil(t, err)
			assert.Equal(t, content, got)
		})

		t.Run("remove", func(t *testing.T) {
			assert.Nil(t, b.Copy())
			assert.Nil(t, origin.FilePath().Remove())
			assert.Nil(t, b.Restore())
			got, err := origin.FilePath().Read()
			assert.Nil(t, err)
			assert.Equal(t, content, got)
		})

	})
}
