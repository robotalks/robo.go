package nav

import (
	"time"

	"github.com/robotalks/robo.go/pkg/l1/msgs"
	"github.com/robotalks/robo.go/pkg/sim"
)

type turnState struct {
	startPose sim.Pose2D
	startTime time.Time
	speed     float64
}

func newTurnState(pose sim.Pose2D, now time.Time, msg *msgs.Nav2DTurn) state {
	if msg.Speed == 0 {
		return nil
	}
	return &turnState{
		startPose: pose,
		startTime: now,
		speed:     float64(msg.Speed),
	}
}

func (s *turnState) estimate(now time.Time) (sim.Pose2D, state) {
	if s.speed == 0 {
		return s.startPose, nil
	}
	pose := s.startPose
	pose.Orientation = pose.Orientation.AddRadians(now.Sub(s.startTime).Seconds() * s.speed)
	return pose, s
}
