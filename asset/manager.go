package asset

import (
	"errors"
	"fmt"
	"image"
	_ "image/png" // for png textures
	"os"

	gl "github.com/go-gl/gl"
)

// ShaderSet is a tuple of shader handles
type ShaderSet struct {
	Vs uint32
	Fs uint32
	Gs uint32
}

// Manager stores Materials, Meshes, Shaders, and Textures.
type Manager struct {
	Materials map[string]*Material
	Meshes    map[string]*Mesh
	Shaders   map[string]uint32
	Programs  map[ShaderSet]uint32
	Textures  map[string]*Texture

	Parent *Manager
}

// NewManager creates and initializes a new Manager
func NewManager(parent *Manager) *Manager {
	var am = &Manager{
		Materials: make(map[string]*Material),
		Meshes:    make(map[string]*Mesh),
		Shaders:   make(map[string]uint32),
		Programs:  make(map[ShaderSet]uint32),
		Textures:  make(map[string]*Texture),
		Parent:    parent,
	}

	return am
}

// AddMaterial adds a Material to the manager. If the Material's name is already
// in use, the operation fails and an error is returned.
func (am *Manager) AddMaterial(m *Material) error {
	if _, ok := am.Materials[m.Name]; ok {
		return fmt.Errorf("asset.Manager.AddMaterial error: material '%s' already exists", m.Name)
	}

	Logger.Printf("Manager: adding Material '%s'\n", m.Name)
	am.Materials[m.Name] = m

	return nil
}

// GetMaterial searches for a Material. If it exists it is returned, otherwise
// nil and false are returned.
func (am *Manager) GetMaterial(name string) (*Material, bool) {
	if m, ok := am.Materials[name]; ok {
		return m, true
	}

	if am.Parent != nil {
		return am.Parent.GetMaterial(name)
	}

	return nil, false
}

// AddMesh adds a Mesh to the Manager. If the Mesh's name is already in use, the
// operation fails and an error is returned.
func (am *Manager) AddMesh(m *Mesh) error {
	if _, ok := am.Meshes[m.Name]; ok {
		return fmt.Errorf("asset.Manager.AddMesh error: Mesh '%s' already exists", m.Name)
	}

	Logger.Printf("Manager: adding Mesh '%s'\n", m.Name)
	am.Meshes[m.Name] = m

	return nil
}

// GetMesh searches for a Mesh. If it exists it is returned, otherwise nil and
// false are returned.
func (am *Manager) GetMesh(name string) (*Mesh, bool) {
	if m, ok := am.Meshes[name]; ok {
		return m, true
	}

	if am.Parent != nil {
		return am.Parent.GetMesh(name)
	}

	return nil, false
}

// AddShader adds a Shader to the Manager. If the Shader's name is already in
// use, the operation fails and an error is returned.
func (am *Manager) AddShader(name string, shader uint32) error {
	if _, ok := am.GetShader(name); ok {
		return fmt.Errorf("asset.Manager.AddShader error: Shader '%s' already exists", name)
	}

	Logger.Printf("Manager: adding Shader '%s'\n", name)
	am.Shaders[name] = shader

	return nil
}

// GetShader searches for a Shader. If it exists it is returned, otherwise 0 and
// false are returned.
func (am *Manager) GetShader(name string) (uint32, bool) {
	if shader, ok := am.Shaders[name]; ok {
		return shader, true
	}

	if am.Parent != nil {
		return am.Parent.GetShader(name)
	}

	return 0, false
}

// LoadShader loads a shader from the file 'name' and returns it. If it already
// exists, it and a nil error is returned. 'typ' indicates the type of shader
// and must be either gl.VERTEX_SHADER, gl.FRAGMENT_SHADER, or
// gl.GEOMETRY_SHADER.
func (am *Manager) LoadShader(typ uint32, name string) (uint32, error) {
	if shader, ok := am.GetShader(name); ok {
		return shader, nil
	}

	Logger.Printf("asset.Manager.LoadShader: loading Shader '%s'\n", name)

	var shader, err = newShader("assets/shaders/"+name, typ)
	if err != nil {
		Logger.Print("asset.Manager.LoadShader: failed")
		return 0, err
	}

	Logger.Print("asset.Manager.LoadShader: shader loaded")

	am.AddShader(name, shader)

	return shader, nil
}

// AddProgram adds a Program to the Manager. If the Program's name is already in
// use, the operation fails and an error is returned.
func (am *Manager) AddProgram(set ShaderSet, prog uint32) error {
	if _, ok := am.GetProgram(set); ok {
		return fmt.Errorf("asset.Manager.AddProgram error: Program '%v' already exists", set)
	}

	Logger.Printf("Manager: adding Program '%v'\n", set)
	am.Programs[set] = prog

	return nil
}

// GetProgram searches for a Program. If it exists it is returned, otherwise 0
// and false are returned.
func (am *Manager) GetProgram(set ShaderSet) (uint32, bool) {
	if prog, ok := am.Programs[set]; ok {
		return prog, true
	}

	if am.Parent != nil {
		return am.Parent.GetProgram(set)
	}

	return 0, false
}

// LoadProgram attempts to generate and return a Program based on the given
// Shader files. The parameters correspond to the vertex shader, fragment
// shader, and geometry shader respectively. The geometry shader is optional.
// If a Program with those Shaders already exists, it and a nil error are
// returned.
func (am *Manager) LoadProgram(vfile, ffile, gfile string) (uint32, error) {
	var (
		set ShaderSet
		err error
	)

	if set.Vs, err = am.LoadShader(gl.VERTEX_SHADER, vfile); err != nil {
		return 0, err
	}
	if set.Fs, err = am.LoadShader(gl.FRAGMENT_SHADER, ffile); err != nil {
		return 0, err
	}
	if len(gfile) > 0 {
		if set.Gs, err = am.LoadShader(gl.GEOMETRY_SHADER, gfile); err != nil {
			return 0, err
		}
	}

	if prog, ok := am.GetProgram(set); ok {
		return prog, nil
	}

	Logger.Printf("Manager: loading Program '%v'\n", set)

	var prog = gl.CreateProgram()
	gl.AttachShader(prog, set.Vs)
	gl.AttachShader(prog, set.Fs)
	if set.Gs > 0 {
		gl.AttachShader(prog, set.Gs)
	}
	gl.LinkProgram(prog)

	var infoLogLen int32
	gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &infoLogLen)

	if infoLogLen > 1 {
		var log = make([]uint8, infoLogLen)
		gl.GetProgramInfoLog(prog, infoLogLen, nil, &log[0])
		return 0, errors.New(string(log))
	}

	am.AddProgram(set, prog)

	return prog, nil
}

// AddTexture adds a Texture to the Manager. If the Texture's name is already in
// use, the operation fails and an error is returned.
func (am *Manager) AddTexture(t *Texture) error {
	if _, ok := am.Textures[t.Name]; ok {
		return fmt.Errorf("asset.Manager.AddTexture error: texture %s already exists", t.Name)
	}

	am.Textures[t.Name] = t

	return nil
}

// GetTexture searches for a Texture. If it exists it is returned, otherwise
// nil and false are returned.
func (am *Manager) GetTexture(name string) (*Texture, bool) {
	if tex, ok := am.Textures[name]; ok {
		return tex, ok
	}

	if am.Parent != nil {
		return am.Parent.GetTexture(name)
	}

	return nil, false
}

// LoadTexture attempts to load a Texture from the given file 'name'. If it
// already exists, it is returned.
func (am *Manager) LoadTexture(name string) (*Texture, error) {
	if tex, ok := am.GetTexture(name); ok {
		return tex, nil
	}

	var (
		err error
		f   *os.File
	)

	if f, err = os.Open("assets/textures/" + name); err != nil {
		return nil, err
	}
	defer f.Close()

	var img image.Image

	if img, _, err = image.Decode(f); err != nil {
		return nil, err
	}

	var tex *Texture

	if tex, err = NewTextureFromImage(name, img); err != nil {
		return nil, err
	}

	am.AddTexture(tex)

	return tex, nil
}

// LoadTextures attempts to load a set of textures from the given file names.
// If at any point an error occurs, nothing is returned save for the error.
func (am *Manager) LoadTextures(names ...string) ([]*Texture, error) {
	var (
		textures = make([]*Texture, len(names))
		err      error
	)

	for i, name := range names {
		textures[i], err = am.LoadTexture(name)
		if err != nil {
			return nil, err
		}
	}

	return textures, nil
}

// Clean ensures that all objects themselves cleaned and are removed from
// memory.
func (am *Manager) Clean() {
	for name, m := range am.Materials {
		Logger.Printf("Manager: deleting Material '%s'\n", name)
		m.Clean()
		delete(am.Materials, name)
	}
	for name, m := range am.Meshes {
		Logger.Printf("Manager: deleting Mesh '%s'\n", name)
		m.Clean()
		delete(am.Meshes, name)
	}
	for set, prog := range am.Programs {
		Logger.Printf("Manager: deleting Program '%v'\n", set)
		gl.DeleteProgram(prog)
		delete(am.Programs, set)
	}
	for name, shader := range am.Shaders {
		Logger.Printf("Manager: deleting Shader '%s'\n", name)
		gl.DeleteShader(shader)
		delete(am.Shaders, name)
	}
	for name, tex := range am.Textures {
		Logger.Printf("Manager: deleting Texture '%s'\n", name)
		tex.Clean()
		delete(am.Textures, name)
	}
}
