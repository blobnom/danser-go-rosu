package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/vector"
)

type MultiPointMover interface {
	Reset(diff *difficulty.Difficulty, id int)
	SetObjects(objs []objects.IHitObject) int
	Update(time float64) vector.Vector2f
	GetEndTime() float64
}
