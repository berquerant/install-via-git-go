package filepathx_test

import (
	"berquerant/install-via-git-go/filepathx"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	baseDir := t.TempDir()

	path, err := filepathx.NewPath(baseDir)
	assert.Nil(t, err)
	assert.Equal(t, baseDir, path.String())

	t.Run("NewPath relative", func(t *testing.T) {
		const relPath = "rel"
		path, err := filepathx.NewPath(relPath)
		assert.Nil(t, err)
		currentDir, err := os.Getwd()
		assert.Nil(t, err)
		assert.Equal(t, filepath.Join(currentDir, relPath), path.String())
	})

	t.Run("Ensure", func(t *testing.T) {
		assert.ErrorIs(t, path.FilePath().Ensure(), filepathx.ErrNotFile)
		p := path.Join("testfile").FilePath()
		assert.Nil(t, p.Ensure())
		assert.Nil(t, p.Ensure()) // noop

		t.Run("ReadWrite", func(t *testing.T) {
			got, err := p.Read()
			assert.Nil(t, err)
			assert.Equal(t, "", got)

			assert.Nil(t, p.Write("str"))
			got, err = p.Read()
			assert.Nil(t, err)
			assert.Equal(t, "str", got)

			assert.Nil(t, p.Ensure()) // noop
			got, err = p.Read()
			assert.Nil(t, err)
			assert.Equal(t, "str", got)
		})
	})

	t.Run("Join", func(t *testing.T) {
		for _, tc := range []struct {
			want string
			p    filepathx.Path
		}{
			{
				want: path.String(),
				p:    path.Join("/root"),
			},
			{
				want: filepath.Join(path.String(), "relpath"),
				p:    path.Join("relpath"),
			},
		} {
			assert.Equal(t, tc.want, tc.p.String())
		}
	})

	t.Run("Ext", func(t *testing.T) {
		for _, tc := range []struct {
			want string
			ok   bool
			p    filepathx.Path
		}{
			{
				p: path,
			},
			{
				want: "el",
				ok:   true,
				p:    path.Join("init.el"),
			},
			{
				want: "zst",
				ok:   true,
				p:    path.Join("init.el.zst"),
			},
		} {
			got, ok := tc.p.Ext()
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.want, got)
		}
	})
}
