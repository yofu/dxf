package object

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yofu/dxf/format"
)

// Plot paper units (code 72).
const (
	PLOT_INCHES      = 0
	PLOT_MILLIMETERS = 1
	PLOT_PIXELS      = 2
)

// PlotSettings is the page setup of a layout (AcDbPlotSettings).
type PlotSettings struct {
	PageSetupName    string  // 1
	Device           string  // 2: printer or configuration file (e.g. "DWG To PDF.pc3")
	PaperSize        string  // 4: canonical media name
	PlotView         string  // 6
	MarginLeft       float64 // 40
	MarginBottom     float64 // 41
	MarginRight      float64 // 42
	MarginTop        float64 // 43
	PaperWidth       float64 // 44
	PaperHeight      float64 // 45
	OriginX          float64 // 46: plot origin (offset)
	OriginY          float64 // 47
	WindowMin        []float64
	WindowMax        []float64
	ScaleNumerator   float64 // 142: real world units
	ScaleDenominator float64 // 143: drawing units
	Flags            int     // 70
	PaperUnits       int     // 72: PLOT_INCHES, PLOT_MILLIMETERS, PLOT_PIXELS
	Rotation         int     // 73: 0, 1 (90 degrees), 2 (180), 3 (270)
	PlotType         int     // 74: 0 last screen display, 1 extents, 2 limits, 3 view, 4 window, 5 layout
	StyleSheet       string  // 7: plot style table (e.g. "monochrome.ctb")
	StandardScale    int     // 75: standard scale type (0: scaled to fit)
	ScaleFactor      float64 // 147
	PaperImageOrigin []float64
	// Extra holds the other group codes as read (shade plot settings, ...).
	Extra [][2]string
}

// DefaultPlotSettings returns the page setup written for new layouts.
func DefaultPlotSettings(model bool) *PlotSettings {
	ps := &PlotSettings{Device: "none_device", WindowMin: []float64{0, 0}, WindowMax: []float64{0, 0},
		ScaleNumerator: 1, ScaleDenominator: 1, Flags: 688, PaperUnits: PLOT_MILLIMETERS, PlotType: 5,
		ScaleFactor: 1, PaperImageOrigin: []float64{0, 0}}
	if model {
		ps.Flags, ps.PlotType = 1712, 0
	}
	return ps
}

// PlotSettingsFrom returns the page setup of the group codes of
// AcDbPlotSettings.
func PlotSettingsFrom(data [][2]string) *PlotSettings {
	ps := DefaultPlotSettings(false)
	ps.Flags, ps.PlotType = 0, 0
	for _, p := range data {
		v := strings.TrimSpace(p[1])
		fv, _ := strconv.ParseFloat(v, 64)
		iv, _ := strconv.Atoi(v)
		switch p[0] {
		case "1":
			ps.PageSetupName = p[1]
		case "2":
			ps.Device = p[1]
		case "4":
			ps.PaperSize = p[1]
		case "6":
			ps.PlotView = p[1]
		case "40":
			ps.MarginLeft = fv
		case "41":
			ps.MarginBottom = fv
		case "42":
			ps.MarginRight = fv
		case "43":
			ps.MarginTop = fv
		case "44":
			ps.PaperWidth = fv
		case "45":
			ps.PaperHeight = fv
		case "46":
			ps.OriginX = fv
		case "47":
			ps.OriginY = fv
		case "48":
			ps.WindowMin[0] = fv
		case "49":
			ps.WindowMin[1] = fv
		case "140":
			ps.WindowMax[0] = fv
		case "141":
			ps.WindowMax[1] = fv
		case "142":
			ps.ScaleNumerator = fv
		case "143":
			ps.ScaleDenominator = fv
		case "70":
			ps.Flags = iv
		case "72":
			ps.PaperUnits = iv
		case "73":
			ps.Rotation = iv
		case "74":
			ps.PlotType = iv
		case "7":
			ps.StyleSheet = p[1]
		case "75":
			ps.StandardScale = iv
		case "147":
			ps.ScaleFactor = fv
		case "148":
			ps.PaperImageOrigin[0] = fv
		case "149":
			ps.PaperImageOrigin[1] = fv
		default:
			ps.Extra = append(ps.Extra, p)
		}
	}
	return ps
}

func (ps *PlotSettings) format(f format.Formatter) {
	f.WriteString(100, "AcDbPlotSettings")
	f.WriteString(1, ps.PageSetupName)
	f.WriteString(2, ps.Device)
	f.WriteString(4, ps.PaperSize)
	f.WriteString(6, ps.PlotView)
	f.WriteFloat(40, ps.MarginLeft)
	f.WriteFloat(41, ps.MarginBottom)
	f.WriteFloat(42, ps.MarginRight)
	f.WriteFloat(43, ps.MarginTop)
	f.WriteFloat(44, ps.PaperWidth)
	f.WriteFloat(45, ps.PaperHeight)
	f.WriteFloat(46, ps.OriginX)
	f.WriteFloat(47, ps.OriginY)
	f.WriteFloat(48, coord2(ps.WindowMin, 0))
	f.WriteFloat(49, coord2(ps.WindowMin, 1))
	f.WriteFloat(140, coord2(ps.WindowMax, 0))
	f.WriteFloat(141, coord2(ps.WindowMax, 1))
	f.WriteFloat(142, ps.ScaleNumerator)
	f.WriteFloat(143, ps.ScaleDenominator)
	f.WriteInt(70, ps.Flags)
	f.WriteInt(72, ps.PaperUnits)
	f.WriteInt(73, ps.Rotation)
	f.WriteInt(74, ps.PlotType)
	f.WriteString(7, ps.StyleSheet)
	f.WriteInt(75, ps.StandardScale)
	f.WriteFloat(147, ps.ScaleFactor)
	f.WriteFloat(148, coord2(ps.PaperImageOrigin, 0))
	f.WriteFloat(149, coord2(ps.PaperImageOrigin, 1))
	for _, p := range ps.Extra {
		writeRawPair(f, p)
	}
}

func coord2(v []float64, i int) float64 {
	if i < len(v) {
		return v[i]
	}
	return 0
}

// writeRawPair writes a group code/value pair kept as read; values of
// float group codes are written by the formatter (as ASCII and binary
// sources give different texts).
func writeRawPair(f format.Formatter, p [2]string) {
	c, err := strconv.Atoi(strings.TrimSpace(p[0]))
	if err != nil {
		return
	}
	if isFloatCode(c) {
		if v, err := strconv.ParseFloat(strings.TrimSpace(p[1]), 64); err == nil {
			f.WriteFloat(c, v)
			return
		}
	}
	f.WriteString(c, p[1])
}

// isFloatCode reports whether group code c has a float value.
func isFloatCode(c int) bool {
	return (c >= 10 && c <= 59) || (c >= 110 && c <= 149) || (c >= 210 && c <= 239) || (c >= 460 && c <= 469) || (c >= 1010 && c <= 1059)
}

// PageSetup sets the page setup of a paper space layout as ezdxf does:
// paper width and height in units (PLOT_MILLIMETERS or PLOT_INCHES),
// margins (top, right, bottom, left), rotation (0 to 3, by 90 degrees) and
// scale (1:scale; drawing units per paper unit). The limits of the layout
// become the paper size.
func (l *Layout) PageSetup(width, height float64, units int, margins [4]float64, rotation int, scale float64, device string) {
	ps := l.PlotSettings
	if ps == nil {
		ps = DefaultPlotSettings(l.IsModel())
		l.PlotSettings = ps
	}
	if device == "" {
		device = "DWG To PDF.pc3"
	}
	if scale <= 0 {
		scale = 1
	}
	unit := "MM"
	if units == PLOT_INCHES {
		unit = "Inches"
	}
	ps.PageSetupName = ""
	ps.Device = device
	ps.PaperSize = fmt.Sprintf("USER_(%.2f_x_%.2f_%s)", width, height, unit)
	ps.PaperWidth, ps.PaperHeight = width, height
	ps.MarginTop, ps.MarginRight, ps.MarginBottom, ps.MarginLeft = margins[0], margins[1], margins[2], margins[3]
	ps.PaperUnits = units
	ps.Rotation = rotation & 3
	ps.ScaleNumerator, ps.ScaleDenominator, ps.ScaleFactor = 1, scale, 1/scale
	ps.StandardScale = 16 // 1:1 is a standard scale, custom scales use the numerator and denominator
	ps.PlotType = 5       // layout
	l.LimMin = []float64{-margins[3] * scale, -margins[2] * scale}
	l.LimMax = []float64{(width - margins[3]) * scale, (height - margins[2]) * scale}
}
