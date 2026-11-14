package dxf

import (
	"github.com/yofu/dxf/drawing"
	"github.com/yofu/dxf/entity"
	"github.com/yofu/dxf/proxygraphic"
	"github.com/yofu/dxf/table"
)

// ProxyGraphics returns the entities drawn by the proxy graphics of an
// entity read that this package does not know (ACAD_PROXY_ENTITY or a
// custom entity: entity.Unknown), with the layers and linetypes of drawing
// d, as ezdxf's proxy graphics support.
func ProxyGraphics(d *drawing.Drawing, e entity.Entity) (entity.Entities, error) {
	u, ok := e.(*entity.Unknown)
	if !ok {
		return nil, ErrNoVirtualEntities
	}
	data, err := proxygraphic.Data(u.Data)
	if err != nil {
		return nil, err
	}
	opt := proxygraphic.Options{Layer: e.Layer()}
	if ts, ok := d.Sections[drawing.TABLES].(table.Tables); ok {
		for _, st := range ts[table.LAYER].Entries() {
			if l, ok := st.(*table.Layer); ok {
				opt.Layers = append(opt.Layers, l)
			}
		}
		for _, st := range ts[table.LTYPE].Entries() {
			if lt, ok := st.(*table.LineType); ok {
				opt.LineTypes = append(opt.LineTypes, lt)
			}
		}
	}
	return proxygraphic.Decode(data, opt)
}
