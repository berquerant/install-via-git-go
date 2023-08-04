package execx

import (
	"fmt"
	"os"
	"strings"
)

type Env map[string]string

func EnvFromEnviron() Env {
	return EnvFromSlice(os.Environ())
}

func NewEnv() Env {
	return Env(make(map[string]string))
}

func EnvFromMap(envMap map[string]string) Env {
	if envMap == nil {
		return Env(make(map[string]string))
	}
	return Env(envMap)
}

func EnvFromSlice(envSlice []string) Env {
	env := NewEnv()
	for _, x := range envSlice {
		xs := strings.SplitN(x, "=", 2)
		if len(xs) != 2 {
			continue
		}
		env.Set(xs[0], xs[1])
	}
	return env
}

func (e Env) Get(key string) (string, bool) {
	v, ok := e[key]
	return v, ok
}

func (e Env) Set(key, value string) {
	e[key] = value
}

func (e Env) Add(other Env) Env {
	result := Env(make(map[string]string))
	for k, v := range e {
		result[k] = v
	}
	for k, v := range other {
		result[k] = v
	}
	return result
}

// IntoSlice converts into os.Environ format.
func (e Env) IntoSlice() []string {
	var (
		i      int
		result = make([]string, len(e))
	)
	for k, v := range e {
		result[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	return result
}

func (e Env) get(key string) string {
	return e[key]
}

const expandMaxAttempts = 10

// Expand expands environment variables in target.
func (e Env) Expand(target string) string {
	var (
		result string
		count  int
	)
	for result = os.Expand(target, e.get); result != target && count < expandMaxAttempts; count++ {
		target = result
		result = os.Expand(result, e.get)
	}
	return result
}

func (e Env) ExpandStrings(target []string) []string {
	result := make([]string, len(target))
	for i, t := range target {
		result[i] = e.Expand(t)
	}
	return result
}
