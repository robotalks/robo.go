package nav

import (
	"math"
	"time"

	"github.com/robotalks/robo.go/pkg/l1/msgs"
	"github.com/robotalks/robo.go/pkg/sim"
)

type driveState struct {
	startPose        sim.Pose2D
	startTime        time.Time
	lastEstimateTime time.Time
	desiredSpeed     float64
	currentSpeed     float64

	accelStartSpeed float64
	accelEndTime    time.Time
	accel           float64
}

func newDriveState(old state, pose sim.Pose2D, now time.Time, msg *msgs.Nav2DDrive) state {
	s := &driveState{
		startPose:        pose,
		startTime:        now,
		lastEstimateTime: now,
		desiredSpeed:     float64(msg.Speed),
		accel:            math.Abs(float64(msg.Accelation)),
	}
	if ds, ok := old.(*driveState); ok && ds != nil {
		s.currentSpeed = ds.currentSpeed
	}
	if s.accel != 0 {
		s.accelStartSpeed = s.currentSpeed
		speedDiff := math.Abs(s.desiredSpeed - s.currentSpeed)
		if speedDiff > 0 {
			s.accelEndTime = s.startTime.Add(time.Duration(speedDiff*1000000/s.accel) * time.Microsecond)
		}
		if s.currentSpeed > s.desiredSpeed {
			s.accel = -s.accel
		}
	} else {
		s.currentSpeed = s.desiredSpeed
	}
	if s.desiredSpeed == 0 && s.currentSpeed == s.desiredSpeed {
		return nil
	}
	return s
}

func (s *driveState) estimate(now time.Time) (sim.Pose2D, state) {
	pose := s.startPose
	if s.lastEstimateTime.Before(s.accelEndTime) {
		nowAccel := now
		if now.After(s.accelEndTime) {
			nowAccel = s.accelEndTime
		}
		secs := nowAccel.Sub(s.startTime).Seconds()
		pose.Pos2D.OffsetBy(pose.Orientation.Project(secs*s.accelStartSpeed + s.accel*secs*secs/2))
		speedIncr := secs * s.accel
		if s.accelStartSpeed < s.desiredSpeed {
			s.currentSpeed = s.accelStartSpeed + speedIncr
		} else {
			s.currentSpeed = s.accelStartSpeed - speedIncr
		}
		s.lastEstimateTime = nowAccel
		if s.accelEndTime.After(nowAccel) {
			// acceleration not completed.
			return pose, s
		}
		// acceleration completed, update startPose and startTime and leave to
		// the following logic to perform post-acceleration estimate.
		s.startPose, s.startTime, s.currentSpeed = pose, s.accelEndTime, s.desiredSpeed
		if !now.After(nowAccel) {
			// here now == nowAccel
			return pose, s
		}
	}
	var next state
	if s.currentSpeed != 0 {
		pose.Pos2D.OffsetBy(pose.Orientation.Project(now.Sub(s.startTime).Seconds() * s.currentSpeed))
		next = s
	}
	s.lastEstimateTime = now
	return pose, next
}
