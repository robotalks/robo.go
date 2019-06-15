package nav

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/robotalks/robo.go/pkg/l1/msgs"
	"github.com/robotalks/robo.go/pkg/sim"
)

func TestDriveEstimate(t *testing.T) {
	testCases := []struct {
		name   string
		from   *driveState
		speed  float32
		accel  float32
		after  time.Duration
		expect float64
	}{
		{
			name:   "no accel",
			speed:  1,
			after:  time.Second,
			expect: 1,
		},
		{
			name:   "no accel reverse",
			speed:  -1,
			after:  time.Second,
			expect: -1,
		},
		{
			name:   "before accel ends",
			speed:  2,
			accel:  1,
			after:  time.Second,
			expect: 0.5,
		},
		{
			name:   "at accel ends",
			speed:  2,
			accel:  1,
			after:  2 * time.Second,
			expect: 2,
		},
		{
			name:   "after accel ends",
			speed:  2,
			accel:  1,
			after:  3 * time.Second,
			expect: 4,
		},
		{
			name: "reduce speed before accel ends",
			from: &driveState{
				currentSpeed: 2,
			},
			speed:  0,
			accel:  1,
			after:  time.Second,
			expect: 1.5,
		},
		{
			name: "reduce speed at accel ends",
			from: &driveState{
				currentSpeed: 2,
			},
			speed:  0,
			accel:  1,
			after:  2 * time.Second,
			expect: 2,
		},
		{
			name: "reduce speed after accel ends",
			from: &driveState{
				currentSpeed: 2,
			},
			speed:  0,
			accel:  1,
			after:  3 * time.Second,
			expect: 2,
		},
		{
			name: "reduce speed after accel ends and reverse",
			from: &driveState{
				currentSpeed: 2,
			},
			speed:  -1,
			accel:  1,
			after:  4 * time.Second,
			expect: 0.5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var msg msgs.Nav2DDrive
			msg.Speed = tc.speed
			msg.Accelation = tc.accel
			var baseTime time.Time
			pose, _ := newDriveState(tc.from, sim.Pose2D{}, baseTime, &msg).
				estimate(baseTime.Add(tc.after))
			require.Equal(t, tc.expect, pose.X)
		})
	}
}
