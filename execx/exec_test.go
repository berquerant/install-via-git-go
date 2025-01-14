package execx_test

import (
	"berquerant/install-via-git-go/execx"
	"berquerant/install-via-git-go/filepathx"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	baseDir := t.TempDir()

	p, err := filepathx.NewPath(baseDir)
	assert.Nil(t, err)
	withDir := execx.WithDir(p.DirPath())

	t.Run("command", func(t *testing.T) {
		for _, tc := range []struct {
			title  string
			args   []string
			env    execx.Env
			stdout string
			stderr string
		}{
			{
				title:  "echo",
				args:   []string{"echo", "me"},
				stdout: "me\n",
			},
			{
				title:  "echo env",
				env:    execx.EnvFromSlice([]string{"nternationalizatio=18"}),
				args:   []string{"echo", "i${nternationalizatio}n"},
				stdout: "i18n\n",
			},
		} {
			tc := tc
			t.Run(tc.title, func(t *testing.T) {
				r, err := execx.NewCommand(tc.args...).Execute(context.TODO(), withDir, execx.WithEnv(tc.env))
				if !assert.Nil(t, err) {
					return
				}
				assert.Equal(t, tc.stdout, r.Stdout)
				assert.Equal(t, tc.stderr, r.Stderr)
			})
		}
	})

	t.Run("script from strings success", func(t *testing.T) {
		r, err := execx.NewExecutorFromStrings([]string{
			"echo one",
			"echo two",
		}, "bash").Execute(context.TODO())
		assert.Nil(t, err)
		assert.Equal(t, "one\ntwo\n", r.Stdout)
	})

	t.Run("script from strings failfast", func(t *testing.T) {
		t.Run("touched", func(t *testing.T) {
			tmpfile := filepath.Join(t.TempDir(), "touched")
			_, err := execx.NewExecutorFromStrings([]string{
				"echo one",
				fmt.Sprintf("touch %s", tmpfile),
				"which unknowncommand",
			}, "bash").Execute(context.TODO())
			assert.NotNil(t, err)
			assert.FileExists(t, tmpfile)
		})

		t.Run("not touched", func(t *testing.T) {
			tmpfile := filepath.Join(t.TempDir(), "not_touched")
			_, err := execx.NewExecutorFromStrings([]string{
				"echo one",
				"which unknowncommand",
				fmt.Sprintf("touch %s", tmpfile),
			}, "bash").Execute(context.TODO())
			assert.NotNil(t, err)
			p, _ := filepathx.NewPath(tmpfile)
			assert.False(t, p.FilePath().Exist())
		})
	})

	t.Run("script", func(t *testing.T) {
		for _, tc := range []struct {
			title  string
			script string
			env    execx.Env
			stdout string
			stderr string
		}{
			{
				title: "stderr",
				script: `echo out
echo err >&2`,
				stdout: "out\n",
				stderr: "err\n",
			},
			{
				title:  "env",
				env:    execx.EnvFromSlice([]string{"nternationalizatio=18"}),
				script: `echo i${nternationalizatio}n >&2`,
				stderr: "i18n\n",
			},
			{
				title:  "refer env from env",
				env:    execx.EnvFromSlice([]string{"A=100", "B=$A/200"}),
				script: `echo $B`,
				stdout: "100/200\n",
			},
		} {
			tc := tc
			t.Run(tc.title, func(t *testing.T) {
				r, err := execx.NewRawScript(tc.script, "bash").Execute(context.TODO(), withDir, execx.WithEnv(tc.env))
				if !assert.Nil(t, err) {
					return
				}
				assert.Equal(t, tc.stdout, r.Stdout)
				assert.Equal(t, tc.stderr, r.Stderr)
			})
		}
	})

	t.Run("cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		cancel()
		_, err := execx.NewCommand("sleep", "1").Execute(ctx, withDir)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
