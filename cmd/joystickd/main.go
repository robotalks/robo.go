package main

//go-build: CGO_ENABLED=0

import (
	"flag"

	"github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/joystick"
	"github.com/robotalks/robo.go/pkg/l1"
	env "github.com/robotalks/robo.go/pkg/l1/env/controller"
)

func init() {
	env.SetControllerType("joystick", l1.ControllerMeta{Description: "Joystick Controller"})
	env.SetupFlags()
	joystick.SetupFlags()
}

func main() {
	flag.Parse()

	env := env.NewConfig().MustNewEnv()
	ctl := joystick.NewConfig().NewController(env)
	framework.NewLoop().Add(env, ctl).RunOrFail()
}
