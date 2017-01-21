package asset

import (
	"errors"
	"image"
	"image/draw"
	"reflect"
	"unsafe"

	gl "github.com/go-gl/gl"
)

// Texture encapsulates texture state
type Texture struct {
	Name string
	Tex  uint32
	Buf  uint32
	W, H int
}

// NewTexture creates a new texture, but does no GL allocation
func NewTexture(name string, w, h int) *Texture {
	var tex, buf uint32

	gl.GenTextures(1, &tex)
	gl.GenBuffers(1, &buf)

	var t = &Texture{
		Name: name,
		Tex:  tex,
		Buf:  buf,
		W:    w, H: h,
	}

	gl.BindTexture(gl.TEXTURE_2D, tex)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

	gl.BindTexture(gl.TEXTURE_2D, 0)

	return t
}

// NewTextureFromImage creates a new Texture from Image data
func NewTextureFromImage(name string, img image.Image) (*Texture, error) {
	var (
		bounds = img.Bounds()
		t      = NewTexture(name, bounds.Dx(), bounds.Dy())
	)

	if err := t.LoadImage(img, 0); err != nil {
		t.Clean()
		return nil, err
	}

	return t, nil
}

// LoadSubImage updates a portion of a texture from a given Image
func (t *Texture) LoadSubImage(img image.Image, offset image.Point, level int32) error {
	var bounds = img.Bounds()

	if bounds.Dx()+offset.X > t.W || bounds.Dy()+offset.Y > t.H {
		return errors.New("asset.Texture.LoadSubImage error: image out of bounds")
	}

	switch imgfmt := img.(type) {
	case *image.RGBA:
		return t.LoadSubRGBA(imgfmt, offset, level)
	default:
		var cpy = image.NewRGBA(bounds)
		draw.Draw(cpy, bounds, img, image.ZP, draw.Src)
		return t.LoadSubRGBA(cpy, offset, level)
	}
}

// LoadImage updates a texture from a given Image
func (t *Texture) LoadImage(img image.Image, level int32) error {
	var bounds = img.Bounds()

	if bounds.Dx() != t.W>>uint(level) || bounds.Dy() != t.H>>uint(level) {
		return errors.New("asset.Texture.LoadImage error: invalid image size")
	}

	switch imgfmt := img.(type) {
	case *image.RGBA:
		return t.LoadRGBA(imgfmt, level)
	default:
		var cpy = image.NewRGBA(bounds)
		draw.Draw(cpy, bounds, img, image.ZP, draw.Src)
		return t.LoadRGBA(cpy, level)
	}
}

// LoadRGBA updates a texture from a given RGBA image
func (t *Texture) LoadRGBA(img *image.RGBA, level int32) error {
	var bounds = img.Bounds()

	gl.BindTexture(gl.TEXTURE_2D, t.Tex)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		level, gl.RGBA,
		int32(bounds.Dx()), int32(bounds.Dy()), 0,
		gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&img.Pix[0]),
	)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.BindBuffer(gl.PIXEL_UNPACK_BUFFER, t.Buf)
	gl.BufferData(gl.PIXEL_UNPACK_BUFFER, len(img.Pix), nil, gl.STREAM_DRAW)

	return nil
}

// LoadSubRGBA updates a portion of a texture from a given RGBA image
func (t *Texture) LoadSubRGBA(img *image.RGBA, offset image.Point, level int32) error {
	var bounds = img.Bounds()

	gl.BindBuffer(gl.PIXEL_UNPACK_BUFFER, t.Buf)
	gl.BufferData(gl.PIXEL_UNPACK_BUFFER, t.W*t.H*4, nil, gl.STREAM_DRAW)

	var ln int32
	gl.GetBufferParameteriv(gl.PIXEL_UNPACK_BUFFER, gl.BUFFER_SIZE, &ln)

	var ptr = uintptr(gl.MapBuffer(gl.PIXEL_UNPACK_BUFFER, gl.WRITE_ONLY))

	if ptr == 0 {
		return errors.New("Assets.Texture.LoadSubRGBA error: could not map buffer")
	}
	var pbuf = *(*[]uint8)(unsafe.Pointer(&reflect.SliceHeader{
		Data: ptr,
		Len:  int(ln),
		Cap:  int(ln),
	}))

	for y := 0; y < bounds.Dy(); y++ {
		var (
			i = (y+offset.Y)*t.W*4 + offset.X*4
			j = y * img.Stride
		)
		copy(pbuf[i:], img.Pix[j:j+bounds.Dx()*4])
	}

	gl.UnmapBuffer(gl.PIXEL_UNPACK_BUFFER)

	gl.BindTexture(gl.TEXTURE_2D, t.Tex)

	gl.TexSubImage2D(
		gl.TEXTURE_2D,
		level,
		0, 0, //offset.X, offset.Y,
		int32(bounds.Dx()), int32(bounds.Dy()),
		gl.RGBA, gl.UNSIGNED_BYTE, nil,
	)

	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.BindBuffer(gl.PIXEL_UNPACK_BUFFER, 0)

	return nil
}

// Use binds texture state
func (t *Texture) Use(i uint32) {
	gl.Enable(gl.TEXTURE_2D)
	gl.ActiveTexture(gl.TEXTURE0 + i)
	gl.BindTexture(gl.TEXTURE_2D, t.Tex)
}

// Release unbinds texture state
func (t *Texture) Release() {
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.Disable(gl.TEXTURE_2D)
}

// Clean deletes texture state
func (t *Texture) Clean() {
	gl.DeleteTextures(1, &t.Tex)
	gl.DeleteBuffers(1, &t.Buf)
}
