package execx

import (
	ex "github.com/berquerant/execx"
)

type Env = ex.Env

func NewEnv() Env {
	return ex.NewEnv()
}

func EnvFromSlice(ss []string) Env {
	return ex.EnvFromSlice(ss)
}

func EnvFromMap(d map[string]string) Env {
	e := NewEnv()
	e.Merge(Env(d))
	return e
}
