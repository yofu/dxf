package object

import (
	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Layout represents LAYOUT object.
// Each of *Model_Space and *Paper_Space has a LAYOUT ("Model", "Layout1"),
// which is referred by its BLOCK_RECORD (code 340) and refers to it (code 330).
type Layout struct {
	handle      int
	src         int            // handle in the file it was read from (kept when writing)
	owner       handle.Handler // ACAD_LAYOUT dictionary
	Name        string         // 1 (AcDbLayout)
	TabOrder    int            // 71
	blockRecord handle.Handler // 330 (AcDbLayout)
	LimMin      []float64      // 10, 20
	LimMax      []float64      // 11, 21
	// Block is the name of the block of the layout:
	// "*Model_Space", "*Paper_Space" (the active paper space layout),
	// "*Paper_Space0", "*Paper_Space1", ...
	Block string
	// PlotSettings is the page setup (AcDbPlotSettings; nil: defaults).
	PlotSettings *PlotSettings
	// LayoutData holds the other group codes of AcDbLayout as read
	// (flags, UCS, extents, references; nil: defaults).
	LayoutData [][2]string
}

// IsObject is for Object interface.
func (l *Layout) IsObject() bool {
	return true
}

// NewLayout creates a new Layout.
func NewLayout(name string, taborder int) *Layout {
	return &Layout{
		Name:     name,
		TabOrder: taborder,
		LimMin:   []float64{0.0, 0.0},
		LimMax:   []float64{420.0, 297.0},
	}
}

// IsModel reports whether the layout is the model space layout.
func (l *Layout) IsModel() bool {
	return l.TabOrder == 0
}

// Format writes data to formatter.
func (l *Layout) Format(f format.Formatter) {
	f.WriteString(0, "LAYOUT")
	f.WriteHex(5, l.handle)
	if l.owner != nil {
		f.WriteString(102, "{ACAD_REACTORS")
		f.WriteHex(330, l.owner.Handle())
		f.WriteString(102, "}")
		f.WriteHex(330, l.owner.Handle())
	}
	ps := l.PlotSettings
	if ps == nil {
		ps = DefaultPlotSettings(l.IsModel())
	}
	ps.format(f)

	f.WriteString(100, "AcDbLayout")
	f.WriteString(1, l.Name)
	flags := "1"
	for _, p := range l.LayoutData {
		if p[0] == "70" {
			flags = p[1]
			break
		}
	}
	f.WriteString(70, flags)
	f.WriteInt(71, l.TabOrder)
	f.WriteFloat(10, l.LimMin[0])
	f.WriteFloat(20, l.LimMin[1])
	f.WriteFloat(11, l.LimMax[0])
	f.WriteFloat(21, l.LimMax[1])
	if l.LayoutData != nil {
		wroteRecord := false
		for _, p := range l.LayoutData {
			switch p[0] {
			case "70":
				continue
			case "330":
				if l.blockRecord != nil && !wroteRecord {
					f.WriteHex(330, l.blockRecord.Handle())
					wroteRecord = true
					continue
				}
			}
			writeRawPair(f, p)
		}
		if !wroteRecord && l.blockRecord != nil {
			f.WriteHex(330, l.blockRecord.Handle())
		}
		return
	}
	for _, code := range []int{12, 22, 32} { // insertion base point
		f.WriteFloat(code, 0.0)
	}
	for _, code := range []int{14, 24, 34} { // minimum extents
		f.WriteFloat(code, 1e20)
	}
	for _, code := range []int{15, 25, 35} { // maximum extents
		f.WriteFloat(code, -1e20)
	}
	f.WriteFloat(146, 0.0)
	for _, code := range []int{13, 23, 33} { // UCS origin
		f.WriteFloat(code, 0.0)
	}
	f.WriteFloat(16, 1.0) // UCS X axis
	f.WriteFloat(26, 0.0)
	f.WriteFloat(36, 0.0)
	f.WriteFloat(17, 0.0) // UCS Y axis
	f.WriteFloat(27, 1.0)
	f.WriteFloat(37, 0.0)
	f.WriteInt(76, 0)
	if l.blockRecord != nil {
		f.WriteHex(330, l.blockRecord.Handle())
	}
}

// String outputs data using default formatter.
func (l *Layout) String() string {
	f := format.NewASCII()
	return l.FormatString(f)
}

// FormatString outputs data using given formatter.
func (l *Layout) FormatString(f format.Formatter) string {
	l.Format(f)
	return f.Output()
}

// Handle returns a handle value.
func (l *Layout) Handle() int {
	return l.handle
}

// SetHandle sets a handle.
func (l *Layout) SetHandle(v *int) {
	l.handle = handle.Next(l.src, v)
}

// SetOwner sets the owner (ACAD_LAYOUT dictionary).
func (l *Layout) SetOwner(h handle.Handler) {
	l.owner = h
}

// SetBlockRecord sets the BLOCK_RECORD of the layout's block (code 330).
func (l *Layout) SetBlockRecord(h handle.Handler) {
	l.blockRecord = h
}

// SourceHandle returns the handle in the file it was read from ("" if new).
func (l *Layout) SourceHandle() string {
	return handle.Format(l.src)
}

// SetSourceHandle sets the handle in the file it was read from;
// it is kept when the drawing is written.
func (l *Layout) SetSourceHandle(h string) {
	l.src = handle.Parse(h)
}
