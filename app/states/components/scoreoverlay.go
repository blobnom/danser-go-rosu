package components

import (
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/difficulty"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/graphics/font"
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"math/rand"
	"strconv"
)

type Overlay interface {
	Update(int64)
	DrawBeforeObjects(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64)
	DrawNormal(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64)
	DrawHUD(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64)
	IsBroken(cursor *graphics.Cursor) bool
	NormalBeforeCursor() bool
}

type ScoreOverlay struct {
	font           *font.Font
	lastTime       int64
	combo          int64
	newCombo       int64
	maxCombo       int64
	newComboScale  *animation.Glider
	newComboScaleB *animation.Glider
	newComboFadeB  *animation.Glider
	leftScale      *animation.Glider
	lastLeft       bool
	lastLeftC      int64
	rightScale     *animation.Glider
	lastRight      bool
	lastRightC     int64
	oldScore       int64
	scoreGlider    *animation.Glider
	ppGlider       *animation.Glider
	ruleset        *osu.OsuRuleSet
	cursor         *graphics.Cursor
	combobreak     *bass.Sample
	music          *bass.Track
	nextEnd        int64
	countLeft      int
	countRight     int
	results        *sprite.SpriteManager
}

func NewScoreOverlay(ruleset *osu.OsuRuleSet, cursor *graphics.Cursor) *ScoreOverlay {
	overlay := new(ScoreOverlay)
	overlay.results = sprite.NewSpriteManager()
	overlay.ruleset = ruleset
	overlay.cursor = cursor
	overlay.font = font.GetFont("Exo 2 Bold")
	overlay.newComboScale = animation.NewGlider(1)
	overlay.newComboScaleB = animation.NewGlider(1)
	overlay.newComboFadeB = animation.NewGlider(1)
	overlay.leftScale = animation.NewGlider(0.9)
	overlay.rightScale = animation.NewGlider(0.9)
	overlay.scoreGlider = animation.NewGlider(0)
	overlay.ppGlider = animation.NewGlider(0)
	overlay.combobreak = bass.NewSample("assets/sounds/combobreak.wav")

	discord.UpdatePlay(cursor.Name)

	ruleset.SetListener(func(cursor *graphics.Cursor, time int64, number int64, position vector.Vector2d, result osu.HitResult, comboResult osu.ComboResult, pp float64, score1 int64) {

		if result == osu.HitResults.Hit100 || result == osu.HitResults.Hit50 || result == osu.HitResults.Miss {
			overlay.AddResult(time, result, position)
		}

		if comboResult == osu.ComboResults.Increase {
			overlay.newComboScaleB.Reset()
			overlay.newComboScaleB.AddEventS(float64(time), float64(time+300), 1.7, 1.1)

			overlay.newComboFadeB.Reset()
			overlay.newComboFadeB.AddEventS(float64(time), float64(time+300), 0.6, 0.0)

			overlay.animate(time)

			overlay.combo = overlay.newCombo
			overlay.newCombo++
			overlay.nextEnd = time + 300
		} else if comboResult == osu.ComboResults.Reset {
			if overlay.newCombo > 20 {
				overlay.combobreak.Play()
			}
			overlay.newCombo = 0
			overlay.combo = 0
		}

		_, _, score, _ := overlay.ruleset.GetResults(overlay.cursor)

		overlay.scoreGlider.Reset()
		overlay.scoreGlider.AddEvent(float64(time), float64(time+1000), float64(score))
		overlay.ppGlider.Reset()
		overlay.ppGlider.AddEvent(float64(time), float64(time+1000), pp)

		overlay.oldScore = score
	})
	return overlay
}

func (overlay *ScoreOverlay) animate(time int64) {
	overlay.newComboScale.Reset()
	overlay.newComboScale.AddEventSEase(float64(time), float64(time+50), 1.0, 1.2, easing.InQuad)
	overlay.newComboScale.AddEventSEase(float64(time+50), float64(time+100), 1.2, 1.0, easing.OutQuad)
}

func (overlay *ScoreOverlay) AddResult(time int64, result osu.HitResult, position vector.Vector2d) {
	var tex *texture.TextureRegion

	switch result {
	case osu.HitResults.Hit100:
		tex = graphics.Hit100
	case osu.HitResults.Hit50:
		tex = graphics.Hit50
	case osu.HitResults.Miss:
		tex = graphics.Hit0
	}

	if tex == nil {
		return
	}

	sprite := sprite.NewSpriteSingle(tex, -float64(time), position, bmath.Origin.Centre)

	fadeIn := float64(time + difficulty.ResultFadeIn)
	postEmpt := float64(time + difficulty.PostEmpt)
	fadeOut := postEmpt + float64(difficulty.ResultFadeOut)

	sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(time), fadeIn, 0.0, 1.0))
	sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, postEmpt, fadeOut, 1.0, 0.0))

	sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time), float64(time+difficulty.ResultFadeIn*0.8), 0.6, 1.1))
	sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, fadeIn, float64(time+difficulty.ResultFadeIn*1.2), 1.1, 0.9))
	sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time+difficulty.ResultFadeIn*1.2), float64(time+difficulty.ResultFadeIn*1.4), 0.9, 1.0))

	if result == osu.HitResults.Miss {
		rotation := rand.Float64()*0.3 - 0.15

		sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Rotate, easing.Linear, float64(time), fadeIn, 0.0, rotation))
		sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Rotate, easing.Linear, fadeIn, fadeOut, rotation, rotation*2))

		sprite.AddTransformUnordered(animation.NewSingleTransform(animation.MoveY, easing.Linear, float64(time), fadeOut, position.Y-5, position.Y+40))
	}

	sprite.SortTransformations()
	sprite.AdjustTimesToTransformations()
	sprite.ResetValuesToTransforms()

	overlay.results.Add(sprite)
}

func (overlay *ScoreOverlay) Update(time int64) {

	if input.Win.GetKey(glfw.KeySpace) == glfw.Press {
		start := overlay.ruleset.GetBeatMap().HitObjects[0].GetBasicData().StartTime
		if start-time > 7000 {
			overlay.music.SetPosition((float64(start) - 2000) / 1000)
		}
	}

	for sTime := overlay.lastTime + 1; sTime <= time; sTime++ {
		overlay.newComboScale.Update(float64(sTime))
		overlay.newComboScaleB.Update(float64(sTime))
		overlay.newComboFadeB.Update(float64(sTime))
		overlay.scoreGlider.Update(float64(sTime))
		overlay.ppGlider.Update(float64(sTime))
	}

	if overlay.combo != overlay.newCombo && overlay.nextEnd < time+140 {
		overlay.animate(time)
	}

	if overlay.combo != overlay.newCombo && overlay.newComboFadeB.GetValue() < 0.01 {
		overlay.combo = overlay.newCombo
	}

	overlay.results.Update(time)

	left := overlay.cursor.LeftButton
	right := overlay.cursor.RightButton

	if !overlay.lastLeft && left {
		overlay.leftScale.AddEventSEase(float64(time), float64(time+100), 0.9, 0.65, easing.InQuad)
		overlay.lastLeftC = time + 100
		overlay.countLeft++
	}

	if overlay.lastLeft && !left {
		cTime := time
		if overlay.lastLeftC > cTime {
			cTime = overlay.lastLeftC
		}
		overlay.leftScale.AddEventSEase(float64(cTime), float64(cTime+100), 0.65, 0.9, easing.OutQuad)
	}

	if !overlay.lastRight && right {
		overlay.rightScale.AddEventSEase(float64(time), float64(time+100), 0.9, 0.65, easing.InQuad)
		overlay.lastRightC = time + 100
		overlay.countRight++
	}

	if overlay.lastRight && !right {
		cTime := time
		if overlay.lastRightC > cTime {
			cTime = overlay.lastRightC
		}
		overlay.rightScale.AddEventSEase(float64(cTime), float64(cTime+100), 0.65, 0.9, easing.OutQuad)
	}

	overlay.lastLeft = left
	overlay.lastRight = right

	overlay.leftScale.Update(float64(time))
	overlay.rightScale.Update(float64(time))

	overlay.lastTime = time
}

func (overlay *ScoreOverlay) SetMusic(music *bass.Track) {
	overlay.music = music
}

func (overlay *ScoreOverlay) DrawBeforeObjects(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	cs := overlay.ruleset.GetBeatMap().Diff.CircleRadius
	sizeX := 512 + cs*2
	sizeY := 384 + cs*2

	batch.SetScale(sizeX/2, sizeY/2)
	batch.SetColor(0, 0, 0, 0.8)
	batch.SetTranslation(vector.NewVec2d(256, 192)) //bg
	batch.DrawUnit(graphics.Pixel.GetRegion())

	batch.SetColor(1, 1, 1, 1)
	batch.SetScale(sizeX/2, 0.3)
	batch.SetTranslation(vector.NewVec2d(256, -cs)) //top line
	batch.DrawUnit(graphics.Pixel.GetRegion())

	batch.SetTranslation(vector.NewVec2d(256, 384+cs)) //bottom line
	batch.DrawUnit(graphics.Pixel.GetRegion())

	batch.SetScale(0.3, sizeY/2)
	batch.SetTranslation(vector.NewVec2d(-cs, 192)) //left line
	batch.DrawUnit(graphics.Pixel.GetRegion())
	batch.SetTranslation(vector.NewVec2d(512+cs, 192)) //right line
	batch.DrawUnit(graphics.Pixel.GetRegion())
	batch.SetScale(1, 1)
}

func (overlay *ScoreOverlay) DrawNormal(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	scale := overlay.ruleset.GetBeatMap().Diff.CircleRadius / 64
	batch.SetScale(scale, scale)

	overlay.results.Draw(overlay.lastTime, batch)
}

func (overlay *ScoreOverlay) DrawHUD(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	scale := settings.Graphics.GetHeightF() / 1080.0
	batch.SetScale(1, -1)
	batch.SetColor(1, 1, 1, overlay.newComboFadeB.GetValue()*alpha)
	graphics.Score.Draw(batch, 0, scale*80*overlay.newComboScaleB.GetValue()/2, scale*80*overlay.newComboScaleB.GetValue(), fmt.Sprintf("%dx", overlay.newCombo))
	batch.SetColor(1, 1, 1, alpha)
	graphics.Score.Draw(batch, 0, scale*80*overlay.newComboScale.GetValue()/2, scale*80*overlay.newComboScale.GetValue(), fmt.Sprintf("%dx", overlay.combo))

	acc, _, _, _ := overlay.ruleset.GetResults(overlay.cursor)

	accText := fmt.Sprintf("%0.2f%%", acc)

	scoreText := fmt.Sprintf("%08d", int64(overlay.scoreGlider.GetValue()))
	ppText := fmt.Sprintf("%0.2fpp", overlay.ppGlider.GetValue())

	graphics.Score.Draw(batch, settings.Graphics.GetWidthF()-graphics.Score.GetWidth(scale*70, scoreText), settings.Graphics.GetHeightF()-scale*70/2, scale*70, scoreText)

	hObjects := overlay.ruleset.GetBeatMap().HitObjects

	startTime := float64(hObjects[0].GetBasicData().StartTime)
	endTime := float64(hObjects[len(hObjects)-1].GetBasicData().EndTime)
	musicPos := 0.0
	if overlay.music != nil {
		musicPos = overlay.music.GetPosition() * 1000
	}

	progress := math.Min(1.0, math.Max(0.0, (musicPos-startTime)/(endTime-startTime)))
	//log.Println(progress)
	batch.SetColor(0.2, 1, 0.2, alpha)

	batch.SetSubScale(scale*150*progress, scale*3)
	batch.SetTranslation(vector.NewVec2d(settings.Graphics.GetWidthF()+scale*(-5-300+progress*150), settings.Graphics.GetHeightF()-scale*(70+1.5)))
	batch.DrawUnit(graphics.Pixel.GetRegion())

	batch.SetColor(1, 1, 1, alpha)
	batch.SetSubScale(1, 2)

	batch.SetSubScale(scale*12, scale*12)
	batch.SetTranslation(vector.NewVec2d(settings.Graphics.GetWidthF()-scale*32*5.1, settings.Graphics.GetHeightF()-scale*70-scale*32/2-scale*4))
	_, _, _, grade := overlay.ruleset.GetResults(overlay.cursor)
	if grade != osu.NONE {
		batch.DrawUnit(*graphics.GradeTexture[int64(grade)])
	}

	graphics.Score.Draw(batch, settings.Graphics.GetWidthF()-graphics.Score.GetWidth(scale*32, accText), settings.Graphics.GetHeightF()-scale*70-scale*32/2-scale*4, scale*32, accText)
	batch.SetScale(1, 1)
	overlay.font.DrawMonospaced(batch, settings.Graphics.GetWidthF()-overlay.font.GetWidthMonospaced(scale*30, ppText)*1.1, 45*scale+scale*30/2, scale*30, ppText)
	batch.SetScale(1, -1)

	scll := scale / 20 * settings.Graphics.GetHeightF()

	batch.SetTranslation(vector.NewVec2d(settings.Graphics.GetWidthF()-scll/2, settings.Graphics.GetHeightF()/2))
	batch.SetScale(scll/2, scll/2)
	batch.SetSubScale(1, 2)
	batch.SetColor(0.2, 0.2, 0.2, alpha)
	batch.DrawUnit(graphics.Pixel.GetRegion())

	counterScl := 0.8 * scll / 2

	batch.SetTranslation(vector.NewVec2d(settings.Graphics.GetWidthF()-scll/2, settings.Graphics.GetHeightF()/2+scll/2))
	batch.SetColor(1, 1, 1, alpha)
	batch.SetSubScale(overlay.leftScale.GetValue(), overlay.leftScale.GetValue())
	batch.DrawUnit(graphics.Pixel.GetRegion())
	leftT := strconv.Itoa(overlay.countLeft)
	len1 := overlay.font.GetWidthMonospaced(counterScl*overlay.leftScale.GetValue(), leftT)
	batch.SetColor(0.8, 0.8, 0.8, alpha)
	overlay.font.DrawMonospaced(batch, settings.Graphics.GetWidthF()-scll/2-len1/2*1.15, settings.Graphics.GetHeightF()/2+scll/2-counterScl/3*overlay.leftScale.GetValue()*1.15, 0.8*overlay.leftScale.GetValue(), leftT)
	batch.SetColor(0, 0, 0, alpha)
	overlay.font.DrawMonospaced(batch, settings.Graphics.GetWidthF()-scll/2-len1/2, settings.Graphics.GetHeightF()/2+scll/2-counterScl/3*overlay.leftScale.GetValue(), 0.8*overlay.leftScale.GetValue(), leftT)

	batch.SetTranslation(vector.NewVec2d(settings.Graphics.GetWidthF()-scll/2, settings.Graphics.GetHeightF()/2-scll/2))
	batch.SetColor(1, 1, 1, alpha)
	batch.SetSubScale(overlay.rightScale.GetValue(), overlay.rightScale.GetValue())
	batch.DrawUnit(graphics.Pixel.GetRegion())
	rightT := strconv.Itoa(overlay.countRight)
	len2 := overlay.font.GetWidthMonospaced(counterScl*overlay.rightScale.GetValue(), rightT)

	batch.SetColor(0.8, 0.8, 0.8, alpha)
	overlay.font.DrawMonospaced(batch, settings.Graphics.GetWidthF()-scll/2-len2/2*1.15, settings.Graphics.GetHeightF()/2-scll/2-counterScl/3*overlay.rightScale.GetValue()*1.15, 0.8*overlay.rightScale.GetValue(), rightT)
	batch.SetColor(0, 0, 0, alpha)
	overlay.font.DrawMonospaced(batch, settings.Graphics.GetWidthF()-scll/2-len2/2, settings.Graphics.GetHeightF()/2-scll/2-counterScl/3*overlay.rightScale.GetValue(), 0.8*overlay.rightScale.GetValue(), rightT)
}

func (overlay *ScoreOverlay) IsBroken(cursor *graphics.Cursor) bool {
	return false
}

func (overlay *ScoreOverlay) NormalBeforeCursor() bool {
	return true
}
