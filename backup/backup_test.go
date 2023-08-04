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

		for _, tc := range []struct {
			title  string
			change func(t *testing.T)
		}{
			{
				title: "remove all",
				change: func(t *testing.T) {
					assert.Nil(t, srcDir.Remove())
				},
			},
			{
				title: "remove file",
				change: func(t *testing.T) {
					_ = file1.Remove()
				},
			},
			{
				title: "remove dir",
				change: func(t *testing.T) {
					assert.Nil(t, dir1.Remove())
				},
			},
			{
				title: "modify file1",
				change: func(t *testing.T) {
					assert.Nil(t, file1.Write("change"))
				},
			},
			{
				title: "modify file2",
				change: func(t *testing.T) {
					assert.Nil(t, file2.Write("change"))
				},
			},
		} {
			tc := tc
			t.Run(tc.title, func(t *testing.T) {
				t.Run("copy", func(t *testing.T) {
					b := prepareBackup(t)
					defer b.Close()
					assert.Nil(t, b.Copy())
					tc.change(t)
					assert.Nil(t, b.Restore())
					check(t)
				})
				t.Run("move", func(t *testing.T) {
					b := prepareBackup(t)
					defer b.Close()
					assert.Nil(t, b.Move())
					tc.change(t)
					assert.Nil(t, b.Restore())
					check(t)
				})
			})
		}
	})

	t.Run("file", func(t *testing.T) {
		const content = "zone"
		origin := basePath.Join("src.file")
		prepare := func(t *testing.T) {
			assert.Nil(t, origin.FilePath().Write(content))
		}
		prepareBackup := func(t *testing.T) *backup.Backup {
			prepare(t)
			b, err := backup.New(basePath.Join("filebackup").DirPath(), origin)
			assert.Nil(t, err)
			return b
		}
		check := func(t *testing.T) {
			got, err := origin.FilePath().Read()
			assert.Nil(t, err)
			assert.Equal(t, content, got)
		}

		t.Run("move", func(t *testing.T) {
			t.Run("modify", func(t *testing.T) {
				b := prepareBackup(t)
				defer b.Close()
				assert.Nil(t, b.Move())
				assert.Nil(t, origin.FilePath().Write("enoz"))
				assert.Nil(t, b.Restore())
				check(t)
			})
		})

		t.Run("copy", func(t *testing.T) {
			t.Run("modify", func(t *testing.T) {
				b := prepareBackup(t)
				defer b.Close()
				assert.Nil(t, b.Copy())
				assert.Nil(t, origin.FilePath().Write("enoz"))
				assert.Nil(t, b.Restore())
				check(t)
			})
			t.Run("remove", func(t *testing.T) {
				b := prepareBackup(t)
				defer b.Close()
				assert.Nil(t, b.Copy())
				assert.Nil(t, origin.FilePath().Remove())
				assert.Nil(t, b.Restore())
				check(t)
			})
		})
	})
}
