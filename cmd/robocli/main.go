package main

import (
	"github.com/robotalks/robo.go/pkg/cli/sh"
	env "github.com/robotalks/robo.go/pkg/l1/env/connector"

	_ "github.com/robotalks/robo.go/pkg/cli/cmds/all"
)

//go-build: CGO_ENABLED=0

func init() {
	env.SetupFlags()
}

func main() {
	sh.Main()
}
