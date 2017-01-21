package asset

import (
	"fmt"

	"github.com/go-gl/gl/v4.5-core/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
)

// attribMap defines the handles for each attribute name.
// TODO: Make it so I don't have to define a handle for each texcoord.
var attribMap = map[string]uint32{
	"pos":       0,
	"color":     1,
	"normal":    2,
	"texcoord0": 3,
	"texcoord1": 4,
	"texcoord2": 5,
}

// Uniforms is a map of uniform names with their values
type Uniforms map[string]interface{}

// Mesh is a collection of AttribArrays
type Mesh struct {
	Name      string                  // Mesh name
	Attribs   map[string]*AttribArray // map of vertex attribute arrays
	Elements  *ElementArray
	Array     uint32 // OpenGL vertex array handle
	Primitive uint32 // OpenGL primitive
	Vertices  int    // number of vertex attribute sets
}

// NewMesh returns an empty Mesh
func NewMesh(name string) *Mesh {
	return &Mesh{
		Name:     name,
		Attribs:  make(map[string]*AttribArray),
		Vertices: -1,
	}
}

// AddArrays adds AttribArrays to the Mesh. Each AttribArray must have the same
// number of attributes elements.
func (m *Mesh) AddArrays(arrays ...*AttribArray) error {
	for _, arr := range arrays {
		if arr == nil {
			continue
		}

		if m.Vertices == -1 {
			m.Vertices = arr.Attribs()
		} else if arr.Attribs() != m.Vertices {
			return fmt.Errorf("Mesh '%s' error: AttribArray sizes are inconsistent.", m.Name)
		}
		m.Attribs[arr.Name] = arr
	}

	return nil
}

// Init creates a vertex array and attaches each vertex attribute array to it.
func (m *Mesh) Init() error {
	gl.GenVertexArrays(1, &m.Array)
	gl.BindVertexArray(m.Array)
	defer gl.BindVertexArray(0)

	for _, arr := range m.Attribs {
		arr.Init(attribMap[arr.Name])
	}

	m.Elements.Init()

	return nil
}

// Clean deletes the vertex array and all attached attribute arrays.
func (m *Mesh) Clean() {
	gl.DeleteVertexArrays(1, &m.Array)
	m.Array = 0
	for _, attr := range m.Attribs {
		attr.Clean()
	}
	m.Elements.Clean()
}

// DrawUniforms draws the Mesh, given a Material and a set of uniforms.
func (m *Mesh) DrawUniforms(material *Material, uniforms Uniforms) {
	material.Use()

	for name, value := range uniforms {
		var loc = material.UniformLocs[name]
		if loc < 0 {
			continue
		}

		switch val := value.(type) {
		case int32:
			gl.Uniform1i(loc, val)
		case float32:
			gl.Uniform1f(loc, val)
		case mgl.Vec2:
			gl.Uniform2fv(loc, 1, &val[0])
		case mgl.Vec3:
			gl.Uniform3fv(loc, 1, &val[0])
		case mgl.Vec4:
			gl.Uniform4fv(loc, 1, &val[0])
		case mgl.Mat4:
			gl.UniformMatrix4fv(loc, 1, false, &val[0])
		default:
			panic("Mesh.DrawUniforms: unhandled uniform type")
		}
	}

	gl.BindVertexArray(m.Array)

	gl.DrawElements(m.Primitive, int32(m.Elements.Len), m.Elements.Type, nil)

	gl.BindVertexArray(0)

	material.Release()
}
