package asset

import (
	"fmt"
	"reflect"

	gl "github.com/go-gl/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
)

// checkSlice confirms whether or not the data is a slice type and panics if it
// is not.
func checkSlice(name string, attr string, data interface{}) {
	if reflect.ValueOf(data).Kind() != reflect.Slice {
		panic(fmt.Errorf("MakeMesh error: Mesh '%s' '%s' data is not a slice", name, attr))
	}
}

// MakeMesh creates a mesh given a set of common attributes. 'pos' and 'elems'
// are mandatory. 'cols' must be a slice of RGBA values. 'nrms' must be a slice
// of 3D values. Each 'texcoord' must be a slice of 2D values.
func MakeMesh(name string, dims int, prim uint32, pos, cols, nrms interface{}, texcoords []interface{}, elems interface{}) (*Mesh, error) {
	var (
		posarr, colarr, nrmarr *AttribArray
		texcarr                = make([]*AttribArray, len(texcoords))
		elemarr                *ElementArray

		mesh = NewMesh(name)

		err error
	)

	defer func() {
		if err != nil {
			posarr.Clean()
			colarr.Clean()
			nrmarr.Clean()
			for _, arr := range texcarr {
				arr.Clean()
			}
			elemarr.Clean()
		}
	}()

	if pos == nil {
		panic(fmt.Errorf("MakeMesh error: Mesh '%s' must have a 'pos' attribute", name))
	} else {
		checkSlice(name, "pos", pos)

		posarr, err = NewAttribArray("pos", dims, pos, gl.STATIC_DRAW)
		if err != nil {
			return nil, err
		}
	}
	if cols != nil {
		checkSlice(name, "color", cols)

		colarr, err = NewAttribArray("color", 4, cols, gl.STATIC_DRAW)
		if err != nil {
			return nil, err
		}
	}
	if nrms != nil {
		checkSlice(name, "normal", cols)

		colarr, err = NewAttribArray("normal", 3, nrms, gl.STATIC_DRAW)
		if err != nil {
			return nil, err
		}
	}
	for i, texcoord := range texcoords {
		var texname = fmt.Sprintf("texcoord%d", i)
		checkSlice(name, texname, texcoord)

		texcarr[i], err = NewAttribArray(texname, 2, texcoord, gl.STATIC_DRAW)
		if err != nil {
			return nil, err
		}
	}
	if elems == nil {
		panic(fmt.Errorf("MakeMesh error: no element array for Mesh '%s'", name))
	} else {
		checkSlice(name, "elems", elems)

		elemarr, err = NewElementArray(elems, gl.STATIC_DRAW)
		if err != nil {
			return nil, err
		}
	}

	if err = mesh.AddArrays(posarr, colarr, nrmarr); err != nil {
		return nil, err
	}
	if err = mesh.AddArrays(texcarr...); err != nil {
		return nil, err
	}

	mesh.Elements = elemarr
	mesh.Primitive = prim

	return mesh, nil
}

// NewBox creates an uninitialized box Mesh with an origin offset about its
// geometric centre
func NewBox(name string, width, height float32, offset mgl.Vec2) (*Mesh, error) {
	var (
		hw = width * 0.5
		hh = height * 0.5
	)

	return MakeMesh(
		name, 2, gl.TRIANGLES,
		[]float32{
			-hw + offset[0], -hh + offset[1],
			-hw + offset[0], hh + offset[1],
			hw + offset[0], hh + offset[1],
			hw + offset[0], -hh + offset[1],
		},
		nil, nil, nil,
		[]uint8{0, 1, 2, 0, 2, 3},
	)
}

// TriConst is the distance from the base of a unit equilateral triangle to its
// centre. It is equal to sqrt(1/12) or tan(30)/2.
const TriConst = 0.28867513459481288225457439025098

// NewEqTriangle creates an uninitialized equilateral triangle Mesh with an
// origin offset about its geometric centre
func NewEqTriangle(name string, base float32, offset mgl.Vec2) (*Mesh, error) {
	var (
		h1 = base * TriConst
		hb = base * 0.5
	)

	return MakeMesh(
		name, 2, gl.TRIANGLES,
		[]float32{
			0 + offset[0], base - h1 + offset[1],
			hb + offset[0], -h1 + offset[1],
			-hb + offset[0], -h1 + offset[1],
		},
		nil, nil, nil,
		[]uint8{0, 1, 2},
	)
}
