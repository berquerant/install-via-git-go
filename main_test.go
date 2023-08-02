package main

import (
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

	t.Run("install self", func(t *testing.T) {
		configPath := filepath.Join(based, "self.yml")

		t.Run("write config", func(t *testing.T) {
			const config = `uri: https://github.com/berquerant/install-via-git-go.git
lock: .lock
install:
  - go build -v -o ivgself
  - ./ivgself help
  - ./ivgself skeleton > skull.yml
  - ./ivgself parse --config skull.yml`
			f, err := os.Create(configPath)
			fail(t, err)
			defer func() {
				fail(t, f.Close())
			}()
			_, err = f.Write([]byte(config))
			fail(t, err)
		})

		t.Run("run", func(t *testing.T) {
			workDir := filepath.Join(based, "work")
			assert.Nil(t, run(ivg, "run", "--config", configPath, "--workDir", workDir))
		})
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
