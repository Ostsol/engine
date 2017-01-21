package asset

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	gl "github.com/go-gl/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

// AttribLenError creates an attribute length error
func AttribLenError(name string, length, Dims int) error {
	return fmt.Errorf("AttribArray error: '%s' length %d is not a multiple of elements %d", name, length, Dims)
}

// AttribArray is a vertex attribute array. The raw attribute data is not
// retained and must be stored separately, if at all.
type AttribArray struct {
	Name string // attrib location name for linking with shader
	Dims int    // number dimensions per Attribute
	Type uint32 // OpenGL datatype of Attribute elements
	Buf  uint32 // the OpenGL buffer handle
	Len  int    // the length of the buffer in elements
	Cap  int    // the maximum capacity of the buffer in elements
}

// NewAttribArray creates a new AttribArray. 'data' must be a numeric slice.
// Currently supported types are:
//   []float32, []float64,
//   []uint8
func NewAttribArray(name string, dims int, data interface{}, usage uint32) (*AttribArray, error) {
	var val = reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("AttribArray error: '%s' data is not a slice", name)
	}

	var (
		size int
		typ  uint32
		ptr  unsafe.Pointer
		l    = val.Len()
	)

	if l == 0 {
		return nil, fmt.Errorf("AttribArray error: '%s' length is zero", name)
	}

	switch v := data.(type) {
	case []uint8:
		size = 1
		typ = gl.UNSIGNED_BYTE
		ptr = unsafe.Pointer(&v[0])
	case []float32:
		size = 4
		typ = gl.FLOAT
		ptr = unsafe.Pointer(&v[0])
	case []mgl32.Vec2:
		size = 4
		typ = gl.FLOAT
		ptr = unsafe.Pointer(&v[0])
		if dims != 2 {
			panic("AttribArray error: dimension mismatch")
		}
		l *= 2
	case []mgl32.Vec3:
		size = 4
		typ = gl.FLOAT
		ptr = unsafe.Pointer(&v[0])
		if dims != 3 {
			panic("AttribArray error: dimension mismatch")
		}
		l *= 3
	case []float64:
		size = 8
		typ = gl.DOUBLE
		ptr = unsafe.Pointer(&v[0])
	case []mgl64.Vec2:
		size = 8
		typ = gl.DOUBLE
		ptr = unsafe.Pointer(&v[0])
		if dims != 2 {
			panic("AttribArray error: dimension mismatch")
		}
		l *= 2
	case []mgl64.Vec3:
		size = 8
		typ = gl.DOUBLE
		ptr = unsafe.Pointer(&v[0])
		if dims != 3 {
			panic("AttribArray error: dimension mismatch")
		}
		l *= 3
	default:
		return nil, fmt.Errorf("AttribArray error: unhandled element type for '%s'", name)
	}

	if l%dims != 0 {
		return nil, AttribLenError(name, l, dims)
	}

	var buf uint32
	gl.GenBuffers(1, &buf)
	var arr = &AttribArray{
		Name: name,
		Dims: dims,
		Type: typ,
		Buf:  buf,
		Len:  l,
		Cap:  l,
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, arr.Buf)
	gl.BufferData(gl.ARRAY_BUFFER, arr.Len*size, ptr, usage)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	return arr, nil
}

// Update updates the data in the AttribArray. The data must be of the same
// type as the original data and no longer.
func (arr *AttribArray) Update(data interface{}) error {
	var val = reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		panic("asset.AttribArray.Update error: data is not a slice")
	}

	var (
		size int
		typ  uint32
		ptr  unsafe.Pointer
		l    = val.Len()
	)

	arr.Len = l
	if l == 0 {
		return nil
	}
	if l > arr.Cap {
		panic("asset.AttribArray.Update error: data length is longer than buffer")
	}

	switch v := data.(type) {
	case []uint8:
		size = 1
		typ = gl.UNSIGNED_BYTE
		ptr = unsafe.Pointer(&v[0])
	case []float32:
		size = 4
		typ = gl.FLOAT
		ptr = unsafe.Pointer(&v[0])
	case []mgl32.Vec2:
		size = 4
		typ = gl.FLOAT
		ptr = unsafe.Pointer(&v[0])
		if arr.Dims != 2 {
			panic("AttribArray error: dimension mismatch")
		}
		arr.Len *= 2
	case []mgl32.Vec3:
		size = 4
		typ = gl.FLOAT
		ptr = unsafe.Pointer(&v[0])
		if arr.Dims != 3 {
			panic("AttribArray error: dimension mismatch")
		}
		arr.Len *= 3
	case []float64:
		size = 8
		typ = gl.DOUBLE
		ptr = unsafe.Pointer(&v[0])
	case []mgl64.Vec2:
		size = 8
		typ = gl.DOUBLE
		ptr = unsafe.Pointer(&v[0])
		if arr.Dims != 2 {
			panic("AttribArray error: dimension mismatch")
		}
		arr.Len *= 2
	case []mgl64.Vec3:
		size = 8
		typ = gl.DOUBLE
		ptr = unsafe.Pointer(&v[0])
		if arr.Dims != 3 {
			panic("AttribArray error: dimension mismatch")
		}
		arr.Len *= 3
	default:
		return fmt.Errorf("AttribArray error: unhandled element type for '%s'", arr.Name)
	}

	if arr.Type != typ {
		panic("asset.AttribArray.Update error: data type does not match array type")
	}

	if arr.Len%arr.Dims != 0 {
		panic("asset.AttribArray.Update error: invalid data length")
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, arr.Buf)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, arr.Len*size, ptr)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	return nil
}

// Init initializes the AttribArray within the provided vertex array.
func (arr *AttribArray) Init(loc uint32) {
	gl.EnableVertexAttribArray(loc)
	gl.BindBuffer(gl.ARRAY_BUFFER, arr.Buf)
	gl.VertexAttribPointer(loc, int32(arr.Dims), arr.Type, false, 0, nil)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

// Attribs returns the number of attributes in the array
func (arr *AttribArray) Attribs() int {
	return arr.Len / arr.Dims
}

// Clean deletes the array buffer
func (arr *AttribArray) Clean() {
	if arr == nil {
		return
	}
	gl.DeleteBuffers(1, &arr.Buf)
	arr.Buf = 0
}

// ElementArray is an attribute array specialized for element indices.
type ElementArray struct {
	Type uint32
	Buf  uint32
	Len  int
	Cap  int
}

// NewElementArray creates an ElementArray
func NewElementArray(data interface{}, usage uint32) (*ElementArray, error) {
	var v = reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		panic("asset.NewElementArray error: data is not a slice")
	}

	var l = v.Len()
	if l == 0 {
		return nil, errors.New("asset.NewElementArray error: data length is zero")
	}

	var buf uint32
	gl.GenBuffers(1, &buf)
	var arr = &ElementArray{
		Buf: buf,
		Len: l,
		Cap: v.Cap(),
	}

	var size int

	switch data.(type) {
	case []uint8:
		arr.Type = gl.UNSIGNED_BYTE
		size = 1
	case []uint32:
		arr.Type = gl.UNSIGNED_INT
		size = 4
	default:
		panic("asset.NewElementArray error: unhandled data type")
	}

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, arr.Buf)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, arr.Len*size, unsafe.Pointer(v.Index(0).Addr().Pointer()), usage)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)

	return arr, nil
}

// Update updates the data within the ElementArray. The data must be of the
// same type as the original and no longer.
func (arr *ElementArray) Update(data interface{}) {
	var v = reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		panic("AttribArray error: data is not a slice")
	}

	var l = v.Len()
	if l > arr.Cap {
		panic("ElementArray error: data is larger than array")
	}

	arr.Len = l

	var (
		size int
		typ  uint32
		ptr  unsafe.Pointer
	)

	switch v := data.(type) {
	case []uint8:
		size = 1
		typ = gl.UNSIGNED_BYTE
		ptr = unsafe.Pointer(&v[0])
	case []uint32:
		size = 4
		typ = gl.UNSIGNED_INT
		ptr = unsafe.Pointer(&v[0])
	}

	if arr.Type != typ {
		panic("asset.ElementArray.Update error: data type does not match array type")
	}

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, arr.Buf)
	gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, arr.Len*size, ptr)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
}

// Init binds the element array.
func (arr *ElementArray) Init() {
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, arr.Buf)
}

// Clean deletes the array buffer.
func (arr *ElementArray) Clean() {
	if arr == nil {
		return
	}
	gl.DeleteBuffers(1, &arr.Buf)
	arr.Buf = 0
}
