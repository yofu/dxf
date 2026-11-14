package entity

import (
	"math"

	"github.com/yofu/dxf/format"
	"github.com/yofu/dxf/handle"
)

// Display flags of images (code 70).
const (
	IMAGE_SHOW             = 1 // show image
	IMAGE_SHOW_NOT_ALIGNED = 2 // show image when not aligned with screen
	IMAGE_USE_CLIPPING     = 4 // use clipping boundary
	IMAGE_TRANSPARENT      = 8 // transparency is on
)

// Clipping boundary types of images (code 71).
const (
	IMAGE_CLIP_RECTANGLE = 1
	IMAGE_CLIP_POLYGON   = 2
)

// ImageDefinition represents IMAGEDEF objects (OBJECTS section): the
// image file of images. Use Drawing.AddImageDefinition to create one.
type ImageDefinition struct {
	FileName        string    // 1: file name (relative or absolute path)
	Size            []float64 // 10, 20: image size in pixels
	PixelSize       []float64 // 11, 21: default size of one pixel in drawing units
	Loaded          bool      // 280
	ResolutionUnits int       // 281: 0 none, 2 centimeters, 5 inch

	handle   int
	src      int
	owner    handle.Handler   // ACAD_IMAGE_DICT dictionary
	reactors []handle.Handler // IMAGEDEF_REACTOR objects of the images (set when writing)
}

// NewImageDefinition creates an image definition of file filename with
// size in pixels.
func NewImageDefinition(filename string, width, height int) *ImageDefinition {
	return &ImageDefinition{
		FileName:  filename,
		Size:      []float64{float64(width), float64(height)},
		PixelSize: []float64{0.01, 0.01},
		Loaded:    true,
	}
}

// IsObject is for object.Object interface.
func (d *ImageDefinition) IsObject() bool { return true }

// Format writes data to formatter.
func (d *ImageDefinition) Format(f format.Formatter) {
	f.WriteString(0, "IMAGEDEF")
	f.WriteHex(5, d.handle)
	f.WriteString(102, "{ACAD_REACTORS")
	if d.owner != nil {
		f.WriteHex(330, d.owner.Handle())
	}
	for _, r := range d.reactors {
		f.WriteHex(330, r.Handle())
	}
	f.WriteString(102, "}")
	if d.owner != nil {
		f.WriteHex(330, d.owner.Handle())
	} else {
		f.WriteString(330, "0")
	}
	f.WriteString(100, "AcDbRasterImageDef")
	f.WriteInt(90, 0)
	f.WriteString(1, d.FileName)
	f.WriteFloat(10, coord(d.Size, 0))
	f.WriteFloat(20, coord(d.Size, 1))
	f.WriteFloat(11, coord(d.PixelSize, 0))
	f.WriteFloat(21, coord(d.PixelSize, 1))
	f.WriteInt(280, boolInt(d.Loaded))
	f.WriteInt(281, d.ResolutionUnits)
}

// Handle returns the handle.
func (d *ImageDefinition) Handle() int { return d.handle }

// SetHandle sets the handle (the handle read is kept).
func (d *ImageDefinition) SetHandle(v *int) { d.handle = handle.Next(d.src, v) }

// SetOwner sets the owner dictionary.
func (d *ImageDefinition) SetOwner(h handle.Handler) { d.owner = h }

// SetReactors sets the IMAGEDEF_REACTOR objects of the images.
func (d *ImageDefinition) SetReactors(hs ...handle.Handler) { d.reactors = hs }

// SourceHandle returns the handle in the file it was read from ("" if new).
func (d *ImageDefinition) SourceHandle() string { return handle.Format(d.src) }

// SetSourceHandle sets the handle in the file it was read from;
// it is kept when the drawing is written.
func (d *ImageDefinition) SetSourceHandle(h string) { d.src = handle.Parse(h) }

// ImageDefReactor represents IMAGEDEF_REACTOR objects: the link from an
// image definition to an image. Drawings create them when writing.
type ImageDefReactor struct {
	image  handle.Handler // 330: the image (owner)
	handle int
	src    int
}

// IsObject is for object.Object interface.
func (r *ImageDefReactor) IsObject() bool { return true }

// Format writes data to formatter.
func (r *ImageDefReactor) Format(f format.Formatter) {
	f.WriteString(0, "IMAGEDEF_REACTOR")
	f.WriteHex(5, r.handle)
	if r.image != nil {
		f.WriteHex(330, r.image.Handle())
	} else {
		f.WriteString(330, "0")
	}
	f.WriteString(100, "AcDbRasterImageDefReactor")
	f.WriteInt(90, 2)
	if r.image != nil {
		f.WriteHex(330, r.image.Handle())
	} else {
		f.WriteString(330, "0")
	}
}

// Handle returns the handle.
func (r *ImageDefReactor) Handle() int { return r.handle }

// SetHandle sets the handle (the handle read is kept).
func (r *ImageDefReactor) SetHandle(v *int) { r.handle = handle.Next(r.src, v) }

// SetSourceHandle sets the handle in the file it was read from.
func (r *ImageDefReactor) SetSourceHandle(h string) { r.src = handle.Parse(h) }

// SourceHandle returns the handle in the file it was read from ("" if new).
func (r *ImageDefReactor) SourceHandle() string { return handle.Format(r.src) }

// Image represents IMAGE and WIPEOUT entities (WIPEOUT has no definition
// and masks the area of its boundary). The image is the parallelogram
// Insert + i*U + j*V (0 <= i <= Size[0], 0 <= j <= Size[1]).
type Image struct {
	*entity
	Definition   *ImageDefinition // 340 (nil for WIPEOUT)
	Insert       []float64        // 10, 20, 30: lower left corner
	U            []float64        // 11, 21, 31: vector of one pixel along the image X axis
	V            []float64        // 12, 22, 32: vector of one pixel along the image Y axis
	Size         []float64        // 13, 23: image size in pixels (1, 1 for WIPEOUT)
	Flags        int              // 70: IMAGE_* display flags
	Clipping     bool             // 280
	Brightness   int              // 281: 0 to 100
	Contrast     int              // 282: 0 to 100
	Fade         int              // 283: 0 to 100
	BoundaryType int              // 71: IMAGE_CLIP_RECTANGLE or IMAGE_CLIP_POLYGON
	Boundary     [][]float64      // 91, 14, 24: clipping boundary in pixel coordinates (origin at the top left corner, Y down)
	ClipInside   bool             // 290: clip mode inside (AC1024 or later)

	reactor       *ImageDefReactor
	defHandle     string // 340 as read, until it is resolved to Definition
	reactorHandle string // 360 as read
}

// IsEntity is for Entity interface.
func (im *Image) IsEntity() bool { return true }

// NewImage creates an image of definition def at the origin, one drawing
// unit per pixel.
func NewImage(def *ImageDefinition) *Image {
	im := &Image{
		entity:       NewEntity(IMAGE),
		Definition:   def,
		Insert:       []float64{0, 0, 0},
		U:            []float64{1, 0, 0},
		V:            []float64{0, 1, 0},
		Size:         []float64{1, 1},
		Flags:        IMAGE_SHOW | IMAGE_SHOW_NOT_ALIGNED,
		Brightness:   50,
		Contrast:     50,
		BoundaryType: IMAGE_CLIP_RECTANGLE,
	}
	if def != nil {
		im.Size = append([]float64{}, def.Size...)
	}
	im.resetBoundary()
	return im
}

// NewWipeout creates a WIPEOUT masking the polygon of vertices (XY plane).
func NewWipeout(vertices ...[]float64) *Image {
	im := NewImage(nil)
	im.Type = WIPEOUT
	im.Flags = IMAGE_SHOW | IMAGE_SHOW_NOT_ALIGNED | IMAGE_USE_CLIPPING
	im.Clipping = true
	if len(vertices) == 0 {
		return im
	}
	min := []float64{math.Inf(1), math.Inf(1)}
	max := []float64{math.Inf(-1), math.Inf(-1)}
	for _, v := range vertices {
		for i := 0; i < 2; i++ {
			min[i] = math.Min(min[i], v[i])
			max[i] = math.Max(max[i], v[i])
		}
	}
	w, h := max[0]-min[0], max[1]-min[1]
	if w == 0 {
		w = 1
	}
	if h == 0 {
		h = 1
	}
	im.Insert = []float64{min[0], min[1], coord(vertices[0], 2)}
	im.U = []float64{w, 0, 0}
	im.V = []float64{0, h, 0}
	var b [][]float64
	for _, v := range vertices {
		b = append(b, []float64{(v[0]-min[0])/w - 0.5, 0.5 - (v[1]-min[1])/h})
	}
	im.SetBoundary(b...)
	return im
}

// IsWipeout reports whether the entity is a WIPEOUT.
func (im *Image) IsWipeout() bool { return im.Type == WIPEOUT }

// DefinitionHandle returns the handle of the definition as read when it is
// not resolved to Definition ("" otherwise).
func (im *Image) DefinitionHandle() string { return im.defHandle }

// SetDefinitionHandle sets the handle of the definition (used when the
// definition is not resolved; Definition takes precedence).
func (im *Image) SetDefinitionHandle(h string) { im.defHandle = h }

// ReactorHandle returns the handle of the IMAGEDEF_REACTOR as read when it
// is not resolved ("" otherwise).
func (im *Image) ReactorHandle() string { return im.reactorHandle }

// SetReactorHandle sets the handle of the IMAGEDEF_REACTOR as read.
func (im *Image) SetReactorHandle(h string) { im.reactorHandle = h }

// Reactor returns the IMAGEDEF_REACTOR of the image (nil until written).
func (im *Image) Reactor() *ImageDefReactor { return im.reactor }

// SetReactor sets the IMAGEDEF_REACTOR of the image and its owner.
func (im *Image) SetReactor(r *ImageDefReactor) {
	im.reactor = r
	if r != nil {
		r.image = im
		im.reactorHandle = ""
	}
}

// SetSize sets the size of the image in drawing units (width along the
// image X axis) and its rotation (degrees).
func (im *Image) SetSize(width, height, rotation float64) {
	rad := rotation * math.Pi / 180
	sx, sy := coord(im.Size, 0), coord(im.Size, 1)
	if sx == 0 {
		sx = 1
	}
	if sy == 0 {
		sy = 1
	}
	im.U = []float64{width / sx * math.Cos(rad), width / sx * math.Sin(rad), 0}
	im.V = []float64{-height / sy * math.Sin(rad), height / sy * math.Cos(rad), 0}
}

// SetBoundary sets the clipping boundary in pixel coordinates (two
// vertices: rectangle; more: polygon, closed if needed) and turns
// clipping on. No vertices reset the boundary to the whole image.
func (im *Image) SetBoundary(vertices ...[]float64) {
	if len(vertices) == 0 {
		im.resetBoundary()
		im.Flags &^= IMAGE_USE_CLIPPING
		im.Clipping = false
		return
	}
	var b [][]float64
	for _, v := range vertices {
		b = append(b, []float64{v[0], v[1]})
	}
	im.BoundaryType = IMAGE_CLIP_RECTANGLE
	if len(b) > 2 {
		im.BoundaryType = IMAGE_CLIP_POLYGON
		if f, l := b[0], b[len(b)-1]; f[0] != l[0] || f[1] != l[1] {
			b = append(b, []float64{f[0], f[1]})
		}
	}
	im.Boundary = b
	im.Flags |= IMAGE_USE_CLIPPING
	im.Clipping = true
}

func (im *Image) resetBoundary() {
	im.BoundaryType = IMAGE_CLIP_RECTANGLE
	im.Boundary = [][]float64{{-0.5, -0.5}, {coord(im.Size, 0) - 0.5, coord(im.Size, 1) - 0.5}}
}

// Format writes data to formatter.
func (im *Image) Format(f format.Formatter) {
	im.entity.Format(f)
	if im.IsWipeout() {
		f.WriteString(100, "AcDbWipeout")
	} else {
		f.WriteString(100, "AcDbRasterImage")
	}
	f.WriteInt(90, 0)
	writePoint(f, 10, im.Insert)
	writePoint(f, 11, im.U)
	writePoint(f, 12, im.V)
	f.WriteFloat(13, coord(im.Size, 0))
	f.WriteFloat(23, coord(im.Size, 1))
	switch {
	case im.IsWipeout():
		f.WriteString(340, "0")
	case im.Definition != nil:
		f.WriteHex(340, im.Definition.Handle())
	case im.defHandle != "":
		f.WriteString(340, im.defHandle)
	default:
		f.WriteString(340, "0")
	}
	f.WriteInt(70, im.Flags)
	f.WriteInt(280, boolInt(im.Clipping))
	f.WriteInt(281, im.Brightness)
	f.WriteInt(282, im.Contrast)
	f.WriteInt(283, im.Fade)
	switch {
	case im.IsWipeout():
		f.WriteString(360, "0")
	case im.reactor != nil:
		f.WriteHex(360, im.reactor.Handle())
	case im.reactorHandle != "":
		f.WriteString(360, im.reactorHandle)
	}
	f.WriteInt(71, im.BoundaryType)
	f.WriteInt(91, len(im.Boundary))
	for _, p := range im.Boundary {
		f.WriteFloat(14, p[0])
		f.WriteFloat(24, p[1])
	}
	if format.AtLeast(f, format.R2010) {
		f.WriteInt(290, boolInt(im.ClipInside))
	}
}

// String outputs data using default formatter.
func (im *Image) String() string {
	f := format.NewASCII()
	return im.FormatString(f)
}

// FormatString outputs data using given formatter.
func (im *Image) FormatString(f format.Formatter) string {
	im.Format(f)
	return f.Output()
}

// Corners returns the corners of the image (or of its clipping boundary
// when clipping is used) in WCS.
func (im *Image) Corners() [][]float64 {
	b := im.Boundary
	if im.Flags&IMAGE_USE_CLIPPING == 0 || len(b) < 2 {
		b = [][]float64{{-0.5, -0.5}, {coord(im.Size, 0) - 0.5, coord(im.Size, 1) - 0.5}}
	}
	if len(b) == 2 {
		b = [][]float64{b[0], {b[1][0], b[0][1]}, b[1], {b[0][0], b[1][1]}}
	}
	h := coord(im.Size, 1)
	var pts [][]float64
	for _, p := range b {
		// pixel coordinates: origin at the top left corner, Y down
		i, j := p[0]+0.5, h-p[1]-0.5
		q := make([]float64, 3)
		for k := 0; k < 3; k++ {
			q[k] = coord(im.Insert, k) + i*coord(im.U, k) + j*coord(im.V, k)
		}
		pts = append(pts, q)
	}
	return pts
}

// BBox returns the bounding box of the corners.
func (im *Image) BBox() ([]float64, []float64) {
	mins := []float64{math.Inf(1), math.Inf(1), math.Inf(1)}
	maxs := []float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for _, p := range im.Corners() {
		for k := 0; k < 3; k++ {
			mins[k] = math.Min(mins[k], p[k])
			maxs[k] = math.Max(maxs[k], p[k])
		}
	}
	return mins, maxs
}

// ApplyTransform transforms the image (any transformation: the pixel
// vectors are transformed).
func (im *Image) ApplyTransform(t Transform) error {
	transformPoint(t, im.Insert)
	transformVector(t, im.U)
	transformVector(t, im.V)
	return nil
}
