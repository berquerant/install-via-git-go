package main

import (
	"berquerant/install-via-git-go/cmd"
	"berquerant/install-via-git-go/exit"
)

func main() {
	if err := cmd.Execute(); err != nil {
		exit.Fail()
	}
}
