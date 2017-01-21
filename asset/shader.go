package asset

import (
	"errors"
	"io/ioutil"
	"os"

	gl "github.com/go-gl/gl"
)

func newShader(file string, typ uint32) (uint32, error) {
	var (
		f      *os.File
		err    error
		s      uint32
		buf    []uint8
		//bufptr *uint8
		//ln     int32
	)

	if f, err = os.Open(file); err != nil {
		return 0, err
	}
	defer f.Close()

	if buf, err = ioutil.ReadAll(f); err != nil {
		return 0, err
	}

	buf = append(buf, 0)

	var source, free = gl.Strs(string(buf))
	//bufptr = &buf[0]
	//ln = int32(len(buf))

	s = gl.CreateShader(typ)
	gl.ShaderSource(s, 1, source, nil)
	free()
	gl.CompileShader(s)

	var infoLogLen int32
	gl.GetShaderiv(s, gl.INFO_LOG_LENGTH, &infoLogLen)

	if infoLogLen > 1 {
		Logger.Printf("asset.newShader error: error compiling '%s'", file)
		var log = make([]byte, infoLogLen)
		gl.GetShaderInfoLog(s, infoLogLen, nil, &log[0])
		return 0, errors.New(string(log))
	}

	return s, nil
}
