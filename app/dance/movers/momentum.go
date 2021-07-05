package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

// https://github.com/TechnoJo4/osu/blob/master/osu.Game.Rulesets.Osu/Replays/Movers/MomentumMover.cs

type MomentumMover struct {
	bz        *curves.Bezier
	last      vector.Vector2f
	startTime float64
	endTime   float64
	first     bool
	wasStream bool
	diff      *difficulty.Difficulty
	id        int
}

func NewMomentumMover() MultiPointMover {
	return &MomentumMover{last: vector.NewVec2f(0, 0), first: true}
}

func (bm *MomentumMover) Reset(diff *difficulty.Difficulty, id int) {
	bm.diff = diff
	bm.first = true
	bm.last = vector.NewVec2f(0, 0)
	bm.id = id
}

func same(mods difficulty.Modifier, o1 objects.IHitObject, o2 objects.IHitObject, skipStackAngles bool) bool {
	return o1.GetStackedStartPositionMod(mods) == o2.GetStackedStartPositionMod(mods) || (skipStackAngles && o1.GetStartPosition() == o2.GetStartPosition())
}

func anorm(a float32) float32 {
	pi2 := 2 * math32.Pi
	a = math32.Mod(a, pi2)
	if a < 0 {
		a += pi2
	}

	return a
}

func anorm2(a float32) float32 {
	a = anorm(a)
	if a > math32.Pi {
		a = -(2 * math32.Pi - a)
	}

	return a
}

func (bm *MomentumMover) SetObjects(objs []objects.IHitObject) int {
	ms := settings.CursorDance.MoverSettings.Momentum[bm.id%len(settings.CursorDance.MoverSettings.Momentum)]

	i := 0

	end := objs[i+0]
	start := objs[i+1]

	hasNext := false
	var next objects.IHitObject
	if len(objs) > 2 {
		if _, ok := objs[i+2].(*objects.Circle); ok {
			hasNext = true
		}
		next = objs[i+2]
	}

	endPos := end.GetStackedEndPositionMod(bm.diff.Mods)
	startPos := start.GetStackedStartPositionMod(bm.diff.Mods)

	dst := endPos.Dst(startPos)

	var a2 float32
	fromLong := false
	for i++; i < len(objs); i++ {
		o := objs[i]
		if s, ok := o.(objects.ILongObject); ok {
			a2 = s.GetStartAngleMod(bm.diff.Mods)
			fromLong = true
			break
		}
		if i == len(objs) - 1 {
			a2 = bm.last.AngleRV(endPos)
			break
		}
		if !same(bm.diff.Mods, o, objs[i+1], ms.SkipStackAngles) {
			a2 = o.GetStackedStartPositionMod(bm.diff.Mods).AngleRV(objs[i+1].GetStackedStartPositionMod(bm.diff.Mods))
			break
		}
	}

	var sq1, sq2 float32
	if next != nil {
		nextPos := next.GetStackedStartPositionMod(bm.diff.Mods)
		sq1 = endPos.DstSq(startPos)
		sq2 = startPos.DstSq(nextPos)
	}

	// stream detection logic stolen from spline mover
	stream := false
	if hasNext && !fromLong && ms.StreamRestrict {
		min := float32(25.0)
		max := float32(10000.0)

		if sq1 >= min && sq1 <= max && bm.wasStream || (sq2 >= min && sq2 <= max) {
			stream = true
		}
	}

	bm.wasStream = stream

	var a1 float32
	if s, ok := end.(objects.ILongObject); ok {
		a1 = s.GetEndAngleMod(bm.diff.Mods)
	} else if bm.first {
		a1 = a2 + math.Pi
	} else {
		a1 = endPos.AngleRV(bm.last)
	}


	mult := ms.DistanceMultOut

	ac := a2 - startPos.AngleRV(endPos)
	area := float32(ms.RestrictArea * math.Pi / 180.0)

	if area > 0 && stream && anorm(ac) < anorm((2 * math32.Pi) - area) {
		a := endPos.AngleRV(startPos)

		sangle := float32(0.5 * math.Pi)
		if anorm(a1 - a) > math32.Pi {
			a2 = a - sangle
		} else {
			a2 = a + sangle
		}

		mult = ms.StreamMult
	} else if !fromLong && area > 0 && math32.Abs(anorm2(ac)) < area {
		a := startPos.AngleRV(endPos)

		offset := float32(ms.RestrictAngle * math.Pi / 180.0)
		if (anorm(a2 - a) < offset) != ms.RestrictInvert {
			a2 = a + offset
		} else {
			a2 = a - offset
		}

		mult = ms.DistanceMult
	} else if next != nil && !fromLong {
		r := sq1 / (sq1 + sq2)
		a := endPos.AngleRV(startPos)
		a2 = a + r * anorm2(a2 - a)
	}

	endTime := end.GetEndTime()
	startTime := start.GetStartTime()
	duration := startTime - endTime

	if ms.DurationTrigger > 0 && duration >= ms.DurationTrigger {
		mult *= ms.DurationMult * (duration / ms.DurationTrigger)
	}

	p1 := vector.NewVec2fRad(a1, dst * float32(mult)).Add(endPos)
	p2 := vector.NewVec2fRad(a2, dst * float32(mult)).Add(startPos)

	if !same(bm.diff.Mods, end, start, ms.SkipStackAngles) {
		bm.last = p2
		bm.bz = curves.NewBezierNA([]vector.Vector2f{endPos, p1, p2, startPos})
	} else {
		bm.bz = curves.NewBezierNA([]vector.Vector2f{endPos, startPos})
	}

	bm.endTime = end.GetEndTime()
	bm.startTime = start.GetStartTime()
	bm.first = false

	return 2
}

func (bm *MomentumMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-bm.endTime)/float32(bm.startTime-bm.endTime), 0, 1)
	return bm.bz.PointAt(t)
}

func (bm *MomentumMover) GetEndTime() float64 {
	return bm.startTime
}
