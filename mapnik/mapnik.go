package mapnik

// #include "mapnik_c_api.h"
import "C"

import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

func init() {
	// register default datasources path and fonts path like the python bindings do
	var err error
	err = RegisterDatasources(pluginPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mapnik: ", err)
	}
	err = RegisterFonts(fontPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mapnik: ", err)
	}
}

func RegisterDatasources(path string) error {
	if C.mapnik_register_datasources(C.CString(path)) != 0 {
		return lastRegisterError()
	}
	return nil
}

func RegisterFonts(path string) error {
	if C.mapnik_register_fonts(C.CString(path)) != 0 {
		return lastRegisterError()
	}
	return nil
}

func lastRegisterError() error {
	return errors.New("mapnik: " + C.GoString(C.mapnik_register_last_error()))
}

// Point in 2D space
type Coord struct {
	X, Y float64
}

// Projection from one reference system to the other
type Projection struct {
	p *C.struct_mapnik_projection_t
}

func (p *Projection) Free() {
	C.mapnik_projection_free(p.p)
	p.p = nil
}

func (p Projection) Forward(coord Coord) Coord {
	c := C.mapnik_coord_t{C.double(coord.X), C.double(coord.Y)}
	c = C.mapnik_projection_forward(p.p, c)
	return Coord{float64(c.x), float64(c.y)}
}

// Map base type
type Map struct {
	m *C.struct_mapnik_map_t
}

func NewMap(width, height uint32) *Map {
	return &Map{C.mapnik_map(C.uint(width), C.uint(height))}
}

func (m *Map) lastError() error {
	return errors.New("mapnik: " + C.GoString(C.mapnik_map_last_error(m.m)))
}

func (m *Map) Load(stylesheet string) error {
	if C.mapnik_map_load(m.m, C.CString(stylesheet)) != 0 {
		return m.lastError()
	}
	return nil
}

func (m *Map) Resize(width, height uint32) {
	C.mapnik_map_resize(m.m, C.uint(width), C.uint(height))
}

func (m *Map) Free() {
	C.mapnik_map_free(m.m)
	m.m = nil
}

func (m *Map) SRS() string {
	return C.GoString(C.mapnik_map_get_srs(m.m))
}

func (m *Map) SetSRS(srs string) {
	C.mapnik_map_set_srs(m.m, C.CString(srs))
}

func (m *Map) ZoomAll() error {
	if C.mapnik_map_zoom_all(m.m) != 0 {
		return m.lastError()
	}
	return nil
}

func (m *Map) ZoomToMinMax(minx, miny, maxx, maxy float64) {
	bbox := C.mapnik_bbox(C.double(minx), C.double(miny), C.double(maxx), C.double(maxy))
	defer C.mapnik_bbox_free(bbox)
	C.mapnik_map_zoom_to_box(m.m, bbox)
}

func (m *Map) RenderToFile(path string) error {
	if C.mapnik_map_render_to_file(m.m, C.CString(path)) != 0 {
		return m.lastError()
	}
	return nil
}

func (m *Map) RenderToMemoryPng() ([]byte, error) {
	i := C.mapnik_map_render_to_image(m.m)
	if i == nil {
		return nil, m.lastError()
	}
	defer C.mapnik_image_free(i)
	b := C.mapnik_image_to_png_blob(i)
	defer C.mapnik_image_blob_free(b)
	return C.GoBytes(unsafe.Pointer(b.ptr), C.int(b.len)), nil
}

func (m *Map) Projection() Projection {
	p := Projection{}
	p.p = C.mapnik_map_projection(m.m)
	return p
}

func (m *Map) SetBufferSize(s int) {
	C.mapnik_map_set_buffer_size(m.m, C.int(s))
}
