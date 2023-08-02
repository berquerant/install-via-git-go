package lock_test

import (
	"berquerant/install-via-git-go/filepathx"
	"berquerant/install-via-git-go/lock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileKeeper(t *testing.T) {
	baseDir := t.TempDir()

	p, err := filepathx.NewPath(baseDir)
	assert.Nil(t, err)

	for _, tc := range []struct {
		name              string
		init              string
		next              string
		wantAfterCommit   string
		wantAfterRollback string
	}{
		{
			name: "empty",
		},
		{
			name:              "commitonly",
			next:              "next",
			wantAfterCommit:   "next",
			wantAfterRollback: "next",
		},
		{
			name:              "rollbackonly",
			init:              "init",
			wantAfterCommit:   "init",
			wantAfterRollback: "init",
		},
		{
			name:              "rollback",
			init:              "init",
			next:              "next",
			wantAfterCommit:   "next",
			wantAfterRollback: "init",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			path := p.Join(tc.name).FilePath()
			assert.Nil(t, path.Ensure())
			defer path.Remove()
			assert.Nil(t, path.Write(tc.init))

			k := lock.NewFileKeeper(path)
			assert.Equal(t, tc.init, k.Pair().Current)
			k.Pair().Next = tc.next

			{
				assert.Nil(t, k.Commit())
				got, err := path.Read()
				assert.Nil(t, err)
				assert.Equal(t, tc.wantAfterCommit, got)
			}
			{
				assert.Nil(t, k.Rollback())
				got, err := path.Read()
				assert.Nil(t, err)
				assert.Equal(t, tc.wantAfterRollback, got)
			}
		})
	}
}
