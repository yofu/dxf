package dxf

import (
	"errors"
	"fmt"
	"strings"

	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/table"
)

// ErrUnknownEntity is returned by ParseEntity for entity types this package does not support.
var ErrUnknownEntity = errors.New("unknown entity type")

// ParseEntities parses ENTITIES section.
// Unsupported entity types are skipped.
func ParseEntities(d *drawing.Drawing, line int, data [][2]string) error {
	es, err := parseEntityList(d, data)
	if err != nil {
		var pe *entityError
		if errors.As(err, &pe) {
			return fmt.Errorf("line %d: %w", line+2*pe.index, pe.err)
		}
		return err
	}
	for _, e := range es {
		d.AddEntity(e)
	}
	return nil
}

type entityError struct {
	index int // index of the first group of the entity in data
	err   error
}

func (e *entityError) Error() string { return e.err.Error() }

func (e *entityError) Unwrap() error { return e.err }

// parseEntityList parses a sequence of entities.
// VERTEX entities following a POLYLINE are attached to it until SEQEND.
// Entity types not supported by this package are skipped.
func parseEntityList(d *drawing.Drawing, data [][2]string) (entity.Entities, error) {
	es := entity.New()
	var poly *entity.Polyline
	var ins *entity.Insert                     // INSERT followed by ATTRIB entities
	handles := map[string]entity.Entity{}      // original handles (5)
	annotations := map[*entity.Leader]string{} // LEADER -> annotation handle (340)
	start := 0
	flush := func(end int) error {
		chunk := data[start:end]
		if len(chunk) == 0 || chunk[0][0] != entityTypeCode {
			return nil // groups before the first entity (should not happen)
		}
		switch strings.ToUpper(strings.TrimSpace(chunk[0][1])) {
		case "SEQEND":
			if h, ok := value(chunk, "5"); ok {
				if poly != nil {
					poly.SetEndSourceHandle(h)
				} else if ins != nil {
					ins.SetEndSourceHandle(h)
				}
			}
			poly = nil
			ins = nil
			return nil
		case "ATTRIB":
			e, err := ParseEntity(d, chunk)
			if err != nil {
				return &entityError{start, err}
			}
			if ins != nil {
				setSourceHandle(e, chunk)
				ins.AddAttribute(e.(*entity.Attribute))
			}
			// an ATTRIB without INSERT is invalid and skipped
			return nil
		case "VERTEX":
			e, err := ParseVertex(d, chunk)
			if err == nil {
				err = parseCommon(d, e, chunk)
			}
			if err != nil {
				return &entityError{start, err}
			}
			setSourceHandle(e, chunk)
			if poly != nil {
				poly.AppendVertex(e.(*entity.Vertex))
			} else {
				es = es.Add(e)
			}
			return nil
		}
		poly = nil
		ins = nil
		e, err := ParseEntity(d, chunk)
		if errors.Is(err, ErrUnknownEntity) {
			u, ok := parseUnknown(d, chunk)
			if !ok {
				return nil // not preserved (refers to other objects)
			}
			if err := parseCommon(d, u, chunk); err != nil {
				return &entityError{start, err}
			}
			es = es.Add(u)
			if h, ok := value(chunk, "5"); ok {
				handles[strings.TrimSpace(h)] = u
				u.SetSourceHandle(strings.TrimSpace(h))
			}
			return nil
		}
		if err != nil {
			return &entityError{start, err}
		}
		switch c := e.(type) {
		case *entity.Polyline:
			poly = c
		case *entity.Insert:
			ins = c
		case *entity.Leader:
			if h, ok := value(chunk, "340"); ok {
				annotations[c] = strings.TrimSpace(h)
			}
		}
		if h, ok := value(chunk, "5"); ok {
			handles[strings.TrimSpace(h)] = e
			if sh, ok := e.(interface{ SetSourceHandle(string) }); ok {
				sh.SetSourceHandle(strings.TrimSpace(h))
			}
		}
		es = es.Add(e)
		return nil
	}
	for i, dt := range data {
		if dt[0] == entityTypeCode && i > start {
			if err := flush(i); err != nil {
				return es, err
			}
			start = i
		}
	}
	if err := flush(len(data)); err != nil {
		return es, err
	}
	for l, h := range annotations {
		if a, ok := handles[h]; ok {
			l.Annotation = a
		}
	}
	return es, nil
}

// ParseEntity parses each entity.
func ParseEntity(d *drawing.Drawing, data [][2]string) (entity.Entity, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("no data")
	}
	if data[0][0] != entityTypeCode {
		return nil, fmt.Errorf("invalid group code: %q", data[0][0])
	}
	f, err := ParseEntityFunc(data[0][1])
	if err != nil {
		return nil, err
	}
	e, err := f(d, data)
	if err != nil {
		return e, err
	}
	return e, parseCommon(d, e, data)
}

// parseCommon parses group codes common to all entities
// that are not handled by each entity parser (62: color number).
func parseCommon(d *drawing.Drawing, e entity.Entity, data [][2]string) error {
	data, xdata := splitXData(data)
	if g, ok := e.(interface{ SetExtensionGroups([][2]string) }); ok {
		if gs := groups102(data); len(gs) > 0 {
			g.SetExtensionGroups(gs)
		}
	}
	if x, ok := e.(entity.XDataHolder); ok && len(xdata) > 0 {
		x.SetXData(append([][2]string(nil), xdata...))
	}
	lp, hasLine := e.(entity.LineProperties)
	for _, dt := range data {
		switch dt[0] {
		case "62":
			if err := setInt(dt, func(val int) { entity.SetColorIndex(e, val) }); err != nil {
				return err
			}
		case "6":
			if !hasLine {
				continue
			}
			name := strings.TrimSpace(dt[1])
			lt, err := d.LineType(name)
			if err != nil {
				// keep the reference even if the linetype is not defined
				lt = table.NewLineType(name, "")
			}
			lp.SetLineType(lt)
		case "48":
			if err := setFloat(dt, func(val float64) { e.SetLtscale(val) }); err != nil {
				return err
			}
		case "420":
			if err := setInt(dt, func(val int) { entity.SetRawColorValues(e, val, "", 0) }); err != nil {
				return err
			}
		case "430":
			entity.SetRawColorValues(e, -1, dt[1], 0)
		case "440":
			if err := setInt(dt, func(val int) { entity.SetRawColorValues(e, -1, "", val) }); err != nil {
				return err
			}
		case "60":
			if v, ok := e.(entity.Visibility); ok {
				if err := setInt(dt, func(val int) { v.SetVisible(val == 0) }); err != nil {
					return err
				}
			}
		case "67":
			if sp, ok := e.(entity.SpaceSelector); ok {
				if err := setInt(dt, func(val int) { sp.SetPaperSpace(val == 1) }); err != nil {
					return err
				}
			}
		case "39":
			if th, ok := e.(entity.Thick); ok {
				if err := setFloat(dt, func(val float64) { th.SetThickness(val) }); err != nil {
					return err
				}
			}
		case "370":
			if !hasLine {
				continue
			}
			if err := setInt(dt, func(val int) { lp.SetLineWeight(val) }); err != nil {
				return err
			}
		}
	}
	return nil
}

// ParseEntityFunc returns a function for parsing according to entity type string.
func ParseEntityFunc(t string) (func(*drawing.Drawing, [][2]string) (entity.Entity, error), error) {
	switch strings.ToUpper(strings.TrimSpace(t)) {
	case "LINE":
		return ParseLine, nil
	case "3DFACE":
		return Parse3DFace, nil
	case "LWPOLYLINE":
		return ParseLwPolyline, nil
	case "CIRCLE":
		return ParseCircle, nil
	case "ARC":
		return ParseArc, nil
	case "POLYLINE":
		return ParsePolyline, nil
	case "VERTEX":
		return ParseVertex, nil
	case "POINT":
		return ParsePoint, nil
	case "TEXT":
		return ParseText, nil
	case "SPLINE":
		return ParseSpline, nil
	case "INSERT":
		return ParseInsert, nil
	case "MTEXT":
		return ParseMText, nil
	case "ELLIPSE":
		return ParseEllipse, nil
	case "HATCH":
		return ParseHatch, nil
	case "SOLID":
		return ParseSolid, nil
	case "DIMENSION":
		return ParseDimension, nil
	case "ATTDEF":
		return ParseAttdef, nil
	case "ATTRIB":
		return ParseAttrib, nil
	case "VIEWPORT":
		return ParseViewportEntity, nil
	case "LEADER":
		return ParseLeader, nil
	case "PDFUNDERLAY", "PDFREFERENCE", "DWFUNDERLAY", "DGNUNDERLAY":
		return ParseUnderlay, nil
	case "IMAGE", "WIPEOUT":
		return ParseImage, nil
	case "XLINE", "RAY":
		return ParseXLine, nil
	case "MLINE":
		return ParseMLine, nil
	case "ACAD_TABLE":
		return ParseAcadTable, nil
	case "MULTILEADER", "MLEADER":
		return ParseMultiLeader, nil
	case "TRACE":
		return ParseTrace, nil
	case "SHAPE":
		return ParseShape, nil
	case "TOLERANCE":
		return ParseTolerance, nil
	case "LIGHT":
		return ParseLight, nil
	case "HELIX":
		return ParseHelix, nil
	case "MESH":
		return ParseMesh, nil
	case "MPOLYGON":
		return ParseMPolygon, nil
	case "BODY", "REGION", "3DSOLID", "SURFACE", "PLANESURFACE", "EXTRUDEDSURFACE", "LOFTEDSURFACE", "REVOLVEDSURFACE", "SWEPTSURFACE":
		return ParseBody, nil
	case "ARC_DIMENSION", "LARGE_RADIAL_DIMENSION":
		return ParseSpecialDimension, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownEntity, t)
	}
}
