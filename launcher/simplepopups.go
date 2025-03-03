package launcher

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/wieku/danser-go/build"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/platform"
	"github.com/wieku/danser-go/framework/util"
	"strconv"
)

func drawTimeMenu(bld *builder) {
	imgui.Checkbox("Skip map's beginning", &bld.skip)

	start := &bld.start
	end := &bld.end

	sliderIntReset("Start time", start, 0, end.ogValue-1, util.FormatSeconds(int(start.value)))

	if start.value >= end.value {
		end.value = start.value + 1
	}

	iPos1 := imgui.CursorPosY()

	sliderIntReset("End time", end, 1, end.ogValue, util.FormatSeconds(int(end.value)))

	iPos2 := imgui.CursorPosY()

	if start.value >= end.value {
		start.value = end.value - 1
	}

	imgui.Dummy(vec2(0, iPos2-iPos1))

	sliderIntReset("Audio offset", &bld.offset, -300, 300, "%dms")
}

func drawSpeedMenu(bld *builder) {
	sliderFloatResetStep("Speed", &bld.speed, 0.1, 3, 0.05, "%.2f")
	imgui.Spacing()

	sliderFloatResetStep("Pitch", &bld.pitch, 0.1, 3, 0.05, "%.2f")
	imgui.Spacing()
}

func drawParamMenu(bld *builder) {
	sliderFloatReset("Approach Rate (AR)", &bld.ar, 0, 10, "%.1f")
	imgui.Spacing()

	if bld.currentMode == Play || bld.currentMode == DanserReplay {
		sliderFloatReset("Overall Difficulty (OD)", &bld.od, 0, 10, "%.1f")
		imgui.Spacing()
	}

	sliderFloatReset("Circle Size (CS)", &bld.cs, 0, 10, "%.1f")
	imgui.Spacing()

	if bld.currentMode == Play || bld.currentMode == DanserReplay {
		sliderFloatReset("Health Drain (HP)", &bld.hp, 0, 10, "%.1f")
		imgui.Spacing()
	}
}

func drawCDMenu(bld *builder) {
	if imgui.BeginTable("dfa", 2) {
		imgui.TableNextColumn()

		imgui.Text("Mirrored cursors:")

		imgui.TableNextColumn()

		imgui.SetNextItemWidth(imgui.TextLineHeight() * 4.5)

		if imgui.InputIntV("##mirrors", &bld.mirrors, 1, 1, 0) {
			if bld.mirrors < 1 {
				bld.mirrors = 1
			}
		}

		imgui.TableNextColumn()

		imgui.Text("Tag cursors:")

		imgui.TableNextColumn()

		imgui.SetNextItemWidth(imgui.TextLineHeight() * 4.5)

		if imgui.InputIntV("##tags", &bld.tags, 1, 1, 0) {
			if bld.tags < 1 {
				bld.tags = 1
			}
		}

		imgui.EndTable()
	}
}

func drawRecordMenu(bld *builder) {
	if imgui.BeginTable("rfa", 2) {
		imgui.TableSetupColumnV("c1rfa", imgui.TableColumnFlagsWidthFixed, 0, uint(0))
		imgui.TableSetupColumnV("c2rfa", imgui.TableColumnFlagsWidthFixed, imgui.TextLineHeight()*7, uint(1))

		imgui.TableNextColumn()

		imgui.AlignTextToFramePadding()
		imgui.Text("Output name:")

		imgui.TableNextColumn()

		imgui.SetNextItemWidth(-1)

		imgui.InputTextV("##oname", &bld.outputName, imgui.InputTextFlagsCallbackCharFilter, imguiPathFilter)

		if bld.currentPMode == Screenshot {
			imgui.TableNextColumn()

			imgui.AlignTextToFramePadding()
			imgui.Text("Screenshot at:")

			imgui.TableNextColumn()

			if imgui.BeginTableV("rrfa", 2, 0, vec2(-1, 0), -1) {
				imgui.TableSetupColumnV("c1rrfa", imgui.TableColumnFlagsWidthStretch, 0, uint(0))
				imgui.TableSetupColumnV("c2rrfa", imgui.TableColumnFlagsWidthFixed, imgui.CalcTextSize("s", false, 0).X+imgui.CurrentStyle().CellPadding().X*2, uint(1))

				imgui.TableNextColumn()

				imgui.SetNextItemWidth(-1)

				valText := strconv.FormatFloat(float64(bld.ssTime), 'f', 3, 64)
				prevText := valText

				if imgui.InputText("##sstime", &valText) {
					parsed, err := strconv.ParseFloat(valText, 64)
					if err != nil {
						valText = prevText
					} else {
						parsed = mutils.ClampF(parsed, 0, float64(bld.end.ogValue))
						bld.ssTime = float32(parsed)
					}
				}

				imgui.TableNextColumn()

				imgui.AlignTextToFramePadding()
				imgui.Text("s")

				imgui.EndTable()
			}
		}

		imgui.EndTable()
	}
}

func drawAbout(dTex texture.Texture) {
	centerTable("about1", -1, func() {
		imgui.Image(imgui.TextureID(dTex.GetID()), vec2(100, 100))
	})

	centerTable("about2", -1, func() {
		imgui.Text("danser-go " + build.VERSION)
	})

	centerTable("about3", -1, func() {
		if imgui.Button("Check for updates") {
			checkForUpdates(true)
		}
	})

	imgui.Dummy(vec2(1, imgui.FrameHeight()))

	centerTable("about4.1", -1, func() {
		imgui.Text("Advanced visualisation multi-tool")
	})

	centerTable("about4.2", -1, func() {
		imgui.Text("for osu!")
	})

	imgui.Dummy(vec2(1, imgui.FrameHeight()))

	if imgui.BeginTableV("about5", 3, imgui.TableFlagsSizingStretchSame, vec2(-1, 0), -1) {
		imgui.TableNextColumn()

		centerTable("aboutgithub", -1, func() {
			if imgui.Button("GitHub") {
				platform.OpenURL("https://wieku.me/danser")
			}
		})

		imgui.TableNextColumn()

		centerTable("aboutdonate", -1, func() {
			if imgui.Button("Donate") {
				platform.OpenURL("https://wieku.me/donate")
			}
		})

		imgui.TableNextColumn()

		centerTable("aboutdiscord", -1, func() {
			if imgui.Button("Discord") {
				platform.OpenURL("https://wieku.me/lair")
			}
		})

		imgui.EndTable()
	}
}

func drawLauncherConfig() {
	imgui.PushStyleVarVec2(imgui.StyleVarCellPadding, vec2(imgui.CurrentStyle().CellPadding().X, 10))

	checkboxOption := func(text string, value *bool) {
		if imgui.BeginTableV(text+"table", 2, 0, vec2(-1, 0), -1) {
			imgui.TableSetupColumnV(text+"table1", imgui.TableColumnFlagsWidthStretch, 0, uint(0))
			imgui.TableSetupColumnV(text+"table2", imgui.TableColumnFlagsWidthFixed, 0, uint(1))

			imgui.TableNextColumn()

			pos1 := imgui.CursorPos()

			imgui.AlignTextToFramePadding()

			imgui.PushTextWrapPos()

			imgui.Text(text)

			imgui.PopTextWrapPos()

			pos2 := imgui.CursorPos()

			imgui.TableNextColumn()

			imgui.SetCursorPos(vec2(imgui.CursorPosX(), (pos1.Y+pos2.Y-imgui.FrameHeightWithSpacing())/2))
			imgui.Checkbox("##ck"+text, value)

			imgui.EndTable()
		}
	}

	checkboxOption("Check for updates on startup", &launcherConfig.CheckForUpdates)

	checkboxOption("Load latest replay on startup", &launcherConfig.LoadLatestReplay)

	checkboxOption("Speed up startup on slow HDDs.\nWon't detect deleted/updated\nmaps!", &launcherConfig.SkipMapUpdate)

	checkboxOption("Show JSON paths in config editor", &launcherConfig.ShowJSONPaths)

	checkboxOption("Show exported videos/images\nin explorer", &launcherConfig.ShowFileAfter)

	checkboxOption("Preview selected maps", &launcherConfig.PreviewSelected)

	imgui.AlignTextToFramePadding()
	imgui.Text("Preview volume")

	volume := int32(launcherConfig.PreviewVolume * 100)

	imgui.PushFont(Font16)

	imgui.SetNextItemWidth(-1)

	if sliderIntSlide("##previewvolume", &volume, 0, 100, "%d%%", 0) {
		launcherConfig.PreviewVolume = float64(volume) / 100
	}

	imgui.PopFont()

	imgui.PopStyleVar()
}
