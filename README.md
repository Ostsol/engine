# engine

This is an attempt at writing a 3D graphics engine in Go, currently using the [go-gl] suite of libraries. I am attempting to make this as modular as possible.

## Modules:

### asset
Asset is a high-level wrapper for basic OpenGL drawing functionality. This includes:
 - materials, which include texture and shader loading and management,
 - meshes, which includes vertex and element arrays

## TODO:
 - audio module using [OpenAl]
 - physics module, perhaps written natively

[//]: # References

[go-gl]: <https://github.com/go-gl/>
[OpenAl]: <https://www.openal.org>
