package asset

import (
	"fmt"

	gl "github.com/go-gl/gl"
)

// Material is a collection of textures and shaders and their related data.
type Material struct {
	Name     string     // material name
	Textures []*Texture // list of textures
	Samplers []string   // list of sampler uniform names

	Prog uint32 // shader program

	AttribLocs  map[string]uint32 // vertex attrib locations
	UniformLocs map[string]int32  // other uniform locations
}

// NewMaterial creates an empty Material.
func NewMaterial(name string) *Material {
	return &Material{
		Name:        name,
		AttribLocs:  make(map[string]uint32),
		UniformLocs: make(map[string]int32),
	}
}

// AddTextures adds Textures to the Material
func (mat *Material) AddTextures(textures ...*Texture) {
	mat.Textures = append(mat.Textures, textures...)
}

// AddSamplers adds sampler names to the Material
func (mat *Material) AddSamplers(samplers ...string) {
	mat.Samplers = append(mat.Samplers, samplers...)
}

// SetProgram sets the shader program
func (mat *Material) SetProgram(prog uint32) {
	mat.Prog = prog
}

// BindAttribLoc manually binds an attribute location handle to an attribute
// name.
func (mat *Material) BindAttribLoc(attrib string, loc uint32) error {
	var attr = ([]uint8)(attrib)
	gl.BindAttribLocation(mat.Prog, loc, &attr[0])

	switch gl.GetError() {
	case gl.INVALID_VALUE:
		return fmt.Errorf("Material '%s' error: attrib '%s' location '%d' is greater than GL_MAX_VERTEX_ATTRIBS", mat.Name, attrib, loc)
	case gl.INVALID_OPERATION:
		return fmt.Errorf("Material '%s' error: attrib '%s' begins with reserved prefix 'gl_'", mat.Name, attrib)
	case gl.NO_ERROR:
		fallthrough
	default:
		return nil
	}
}

// InitUniformLocs initializes a table of uniform location handles, given a
// set of uniform names.
func (mat *Material) InitUniformLocs(uniforms ...string) error {
	if mat.Prog == 0 {
		return fmt.Errorf("Material error: material '%s' has no shader program from which to get uniform locations", mat.Name)
	}

	for _, name := range uniforms {
		var (
			bytes = ([]uint8)(name)
			loc   = gl.GetUniformLocation(mat.Prog, &bytes[0])
		)
		if loc == -1 {
			return fmt.Errorf("Material error: material '%s' has no uniform '%s'", mat.Name, name)
		}
		mat.UniformLocs[name] = loc
	}

	return nil
}

// Use binds the Material's textures and shader program for rendering.
func (mat *Material) Use() {
	for i, tex := range mat.Textures {
		tex.Use(uint32(i))
	}
	gl.UseProgram(mat.Prog)
}

// Release unbinds the Material's shader program and textures.
func (mat *Material) Release() {
	gl.UseProgram(0)

	for _, tex := range mat.Textures {
		tex.Release()
	}
}

// Clean deassociates the shader program from the material
func (mat *Material) Clean() {
	mat.Prog = 0
}
