package drawing

import (
	"errors"

	"github.com/yofu/dxf/entity"
)

// add applies the current properties to e and adds it.
func (d *Drawing) add(e entity.Entity) {
	d.applyCurrent(e)
	d.AddEntity(e)
}

// XLine creates an XLINE (infinite line) through point in direction.
func (d *Drawing) XLine(point, direction []float64) (*entity.XLine, error) {
	x := entity.NewXLine(point, direction)
	d.add(x)
	return x, nil
}

// Ray creates a RAY from point in direction.
func (d *Drawing) Ray(point, direction []float64) (*entity.XLine, error) {
	x := entity.NewRay(point, direction)
	d.add(x)
	return x, nil
}

// Trace creates a TRACE with 4 corner points (in the order of SOLID).
func (d *Drawing) Trace(points [][]float64) (*entity.Solid, error) {
	if len(points) != 4 {
		return nil, errors.New("trace needs 4 points")
	}
	s := entity.NewSolid()
	s.Type = entity.TRACE
	for i, p := range points {
		copy(s.Points[i], p)
	}
	d.add(s)
	return s, nil
}

// Shape creates a SHAPE of name (a shape of a shape file loaded with
// AddShapeFile) at insert with size.
func (d *Drawing) Shape(name string, insert []float64, size float64) (*entity.Shape, error) {
	s := entity.NewShape(name, insert, size)
	d.add(s)
	return s, nil
}

// Tolerance creates a TOLERANCE (feature control frame) with content at
// insert, using the current dimension style "Standard".
func (d *Drawing) Tolerance(content string, insert []float64) (*entity.Tolerance, error) {
	tl := entity.NewTolerance(content, insert)
	d.add(tl)
	return tl, nil
}

// Light creates a LIGHT of lightType (entity.LIGHT_POINT, ...) at location.
func (d *Drawing) Light(name string, lightType int, location []float64) (*entity.Light, error) {
	l := entity.NewLight(name, lightType, location)
	d.add(l)
	return l, nil
}

// Helix creates a HELIX around the Z axis through base with radius, pitch
// (height of one turn) and turns, counterclockwise if ccw.
func (d *Drawing) Helix(base []float64, radius, pitch, turns float64, ccw bool) (*entity.Helix, error) {
	if radius <= 0 || turns <= 0 {
		return nil, errors.New("helix: radius and turns must be positive")
	}
	h := entity.NewHelix(radius, pitch, turns, ccw)
	if len(base) > 0 {
		h.ApplyTransform(entity.Move(base[0], coordAt(base, 1), coordAt(base, 2)))
	}
	d.add(h)
	return h, nil
}

func coordAt(v []float64, i int) float64 {
	if i < len(v) {
		return v[i]
	}
	return 0
}

// Mesh creates a MESH of vertices and faces (vertex indices).
func (d *Drawing) Mesh(vertices [][]float64, faces [][]int) (*entity.Mesh, error) {
	for _, f := range faces {
		for _, i := range f {
			if i < 0 || i >= len(vertices) {
				return nil, errors.New("mesh: vertex index out of range")
			}
		}
	}
	m := entity.NewMesh(vertices, faces)
	d.add(m)
	return m, nil
}

// MPolygon creates a solid MPOLYGON with the closed polyline paths
// (the first is external, the following are holes).
func (d *Drawing) MPolygon(paths ...[][]float64) (*entity.MPolygon, error) {
	if len(paths) == 0 {
		return nil, errors.New("mpolygon needs a path")
	}
	m := entity.NewMPolygon()
	for _, p := range paths {
		m.AddPath(p, nil)
	}
	d.add(m)
	return m, nil
}

// Body creates an entity with ACIS data: name is "3DSOLID", "BODY",
// "REGION" or a surface type, sat the ACIS data as SAT text lines.
func (d *Drawing) Body(name string, sat []string) (*entity.Body, error) {
	if entity.EntityTypeValue(name) != entity.BODY {
		return nil, errors.New("body: unknown ACIS entity type " + name)
	}
	b := entity.NewBody(name, sat)
	d.add(b)
	return b, nil
}

// classes of entity types (CLASSES section): C++ class, application,
// proxy flags, entity (as ezdxf writes them)
var entityClasses = map[string][]string{
	"HELIX":                  {"HELIX", "AcDbHelix", "ObjectDBX Classes", "4095", "1"},
	"MPOLYGON":               {"MPOLYGON", "AcDbMPolygon", "AcMPolygonObj15", "1025", "1"},
	"ARC_DIMENSION":          {"ARC_DIMENSION", "AcDbArcDimension", "ObjectDBX Classes", "1025", "1"},
	"LARGE_RADIAL_DIMENSION": {"LARGE_RADIAL_DIMENSION", "AcDbRadialDimensionLarge", "ObjectDBX Classes", "1025", "1"},
	"SURFACE":                {"SURFACE", "AcDbSurface", "ObjectDBX Classes", "4095", "1"},
	"PLANESURFACE":           {"PLANESURFACE", "AcDbPlaneSurface", "ObjectDBX Classes", "4095", "1"},
	"EXTRUDEDSURFACE":        {"EXTRUDEDSURFACE", "AcDbExtrudedSurface", "ObjectDBX Classes", "4095", "1"},
	"LOFTEDSURFACE":          {"LOFTEDSURFACE", "AcDbLoftedSurface", "ObjectDBX Classes", "0", "1"},
	"REVOLVEDSURFACE":        {"REVOLVEDSURFACE", "AcDbRevolvedSurface", "ObjectDBX Classes", "0", "1"},
	"SWEPTSURFACE":           {"SWEPTSURFACE", "AcDbSweptSurface", "ObjectDBX Classes", "0", "1"},
}

// entityTypeName returns the DXF type name of e.
func entityTypeName(e entity.Entity) string {
	switch x := e.(type) {
	case *entity.Body:
		return x.TypeName()
	case *entity.Helix:
		return "HELIX"
	case *entity.MPolygon:
		return "MPOLYGON"
	case *entity.Dimension:
		switch x.Type {
		case entity.DIM_ARC:
			return "ARC_DIMENSION"
		case entity.DIM_LARGE_RADIAL:
			return "LARGE_RADIAL_DIMENSION"
		}
	}
	return ""
}

// ensureEntityClasses adds the classes of the entity types used.
func (d *Drawing) ensureEntityClasses() {
	d.forEachEntity(func(e entity.Entity) {
		if c, ok := entityClasses[entityTypeName(e)]; ok {
			d.ensureClass(c)
		}
	})
}
