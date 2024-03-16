package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func fail(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

type testRunner struct {
	ivg   string
	based string
}

func (r *testRunner) parseSkeleton(t *testing.T) {
	skeletonCmd := exec.Command(r.ivg, "skeleton")
	parseCmd := exec.Command(r.ivg, "parse", "--config", "-")
	skeletonStdout, err := skeletonCmd.StdoutPipe()
	fail(t, err)
	parseCmd.Stdin = skeletonStdout
	parseCmd.Stdout = os.Stdout
	parseCmd.Stderr = os.Stderr

	fail(t, skeletonCmd.Start())
	fail(t, parseCmd.Start())
	fail(t, skeletonCmd.Wait())
	fail(t, parseCmd.Wait())
}

type testRunnerArg struct {
	branch    string
	installed bool
	commit    string
	opt       []string
}

func (r *testRunner) installSelf(t *testing.T, arg *testRunnerArg) {
	installCheck, err := os.CreateTemp(r.based, "install_check")
	const lockFile = ".lock"
	fail(t, err)
	configPath := filepath.Join(r.based, "self.yml")
	const configTemplate = `uri: https://github.com/berquerant/install-via-git-go.git
branch: %[1]s
lock: %[2]s
install:
  - go build -v -o ivgself
  - ./ivgself help
  - ./ivgself skeleton > skull.yml
  - ./ivgself parse --config skull.yml
  - rm -f %[3]s`
	config := fmt.Sprintf(configTemplate,
		arg.branch,
		lockFile,
		installCheck.Name(),
	)
	f, err := os.Create(configPath)
	fail(t, err)
	defer func() {
		fail(t, f.Close())
	}()
	_, err = f.Write([]byte(config))
	fail(t, err)

	workDir := filepath.Join(r.based, "work")
	assert.Nil(t, run(r.ivg, append([]string{"run", "--config", configPath, "--workDir", workDir}, arg.opt...)...))

	if arg.installed {
		assert.NoFileExists(t, installCheck.Name())
	} else {
		assert.FileExists(t, installCheck.Name())
	}

	if arg.commit == "" {
		t.Log("skip check commit")
		return
	}

	lock, err := os.Open(filepath.Join(workDir, lockFile))
	fail(t, err)
	defer lock.Close()
	gotCommit, err := io.ReadAll(lock)
	fail(t, err)
	assert.Equal(t, arg.commit, string(gotCommit))
}

func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	based := t.TempDir()
	ivg := filepath.Join(based, "install-via-git")

	fail(t, compileBinary(ivg))

	t.Run("skeleton", func(t *testing.T) {
		skeleton := filepath.Join(based, "skeleton.yml")
		f, err := os.Create(skeleton)
		fail(t, err)

		t.Run("generate", func(t *testing.T) {
			defer func() {
				fail(t, f.Close())
			}()

			out, err := exec.Command(ivg, "skeleton").Output()
			assert.Nil(t, err)
			_, err = f.Write(out)
			assert.Nil(t, err)

			t.Run("parse", func(t *testing.T) {
				assert.Nil(t, run(ivg, "parse", "--config", skeleton))
			})
		})
	})

	runner := &testRunner{
		ivg:   ivg,
		based: based,
	}

	const (
		v0_8_0 = "a84439250e564509cc6b0d6736dd71ad94c4fec1"
		v0_9_0 = "6b6c6e3aa5482b59d769c1b6793f45fb8e1d45f7"
	)

	t.Run("parse from stdin", runner.parseSkeleton)
	t.Run("install self v0.8.0", func(t *testing.T) {
		arg := &testRunnerArg{
			branch:    "v0.8.0",
			installed: true,
			commit:    v0_8_0,
		}
		runner.installSelf(t, arg)
	})
	t.Run("install self v0.9.0 noop", func(t *testing.T) {
		arg := &testRunnerArg{
			branch:    "v0.9.0",
			installed: false,
			commit:    v0_8_0,
		}
		runner.installSelf(t, arg)
	})
	t.Run("install self v0.9.0 retry", func(t *testing.T) {
		arg := &testRunnerArg{
			branch:    "v0.9.0",
			installed: true,
			commit:    v0_8_0,
			opt:       []string{"--retry"},
		}
		runner.installSelf(t, arg)
	})
	t.Run("install self v0.9.0 update", func(t *testing.T) {
		arg := &testRunnerArg{
			branch:    "v0.9.0",
			installed: true,
			commit:    v0_9_0,
			opt:       []string{"--update"},
		}
		runner.installSelf(t, arg)
	})
	t.Run("install self main update", func(t *testing.T) {
		arg := &testRunnerArg{
			branch:    "main",
			installed: true,
			opt:       []string{"--update"},
		}
		runner.installSelf(t, arg)
	})
}

func compileBinary(path string) error {
	return run("go", "build", "-o", path, "-v")
}

func run(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type tempd string

func (d tempd) ensure() error {
	return os.MkdirAll(string(d), 0755)
}

func (d tempd) join(elem string) tempd {
	return tempd(filepath.Join(string(d), elem))
}
