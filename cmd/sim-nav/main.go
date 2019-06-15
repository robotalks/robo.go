package main

//go-build: CGO_ENABLED=0

import (
	"flag"

	fx "github.com/robotalks/robo.go/pkg/framework"
	"github.com/robotalks/robo.go/pkg/l1"
	env "github.com/robotalks/robo.go/pkg/l1/env/controller"
	navbot "github.com/robotalks/robo.go/pkg/sim/bots/nav"
	"github.com/robotalks/robo.go/pkg/sim/visualization/see"
)

const (
	imageSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="-150 -150 300 300">
		<g>
			<rect x="-100" y="-100" width="160" height="50" rx="5" />
			<rect x="-100" y="50" width="160" height="50" rx="5" />
			<circle cx="-45" cy="0" r="45" fill="none" stroke="black" stroke-width="2" />
			<path d="M 15 -50 L 100 0 L 15 50 Z" />
		</g>
	</svg>`
)

func init() {
	env.SetControllerType("sim-nav", l1.ControllerMeta{Description: "Simulation: navigation"})
	env.SetupFlags()
	see.SetupFlags()
	navbot.SetupFlags()
}

func main() {
	flag.Parse()

	env := env.NewConfig().MustNewEnv()
	bot := navbot.NewConfig().NewController(env)
	vis := see.NewConfig().NewAdapter()
	vis.Mapper = see.MapObjectFunc(func(obj see.VisibleObject) []see.Object {
		return []see.Object{
			see.ObjectFrom("image", obj).With("src", "data:image/svg+xml;utf8,"+imageSVG),
		}
	})
	vis.Subscribe(bot)

	fx.NewLoop().
		Add(env, bot, vis).
		RunOrFail()
}
