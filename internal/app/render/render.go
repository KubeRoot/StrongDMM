package render

import (
	"log"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/SpaiR/strongdmm/pkg/dm/dmmap"
	"github.com/SpaiR/strongdmm/pkg/platform"
)

var (
	initialized  bool
	program      uint32
	uniTransform int32
	indicesCache []uint32 // Reuse all the same buffer to avoid allocations.
)

type Render struct {
	State         *State
	bucket        *bucket
	vao, vbo, ebo uint32
}

func (r *Render) Dispose() {
	gl.DeleteVertexArrays(1, &r.vao)
	gl.DeleteBuffers(1, &r.vbo)
	gl.DeleteBuffers(1, &r.ebo)
	log.Println("[canvas] disposed")
}

func New(dmm *dmmap.Dmm) *Render {
	if !initialized {
		initShaderProgram()
		initUniforms()
	}

	canvas := &Render{
		State:  newState(),
		bucket: createBucket(dmm),
	}

	gl.UseProgram(program)
	canvas.initBuffers()
	canvas.fillArrayBuffer()
	gl.UseProgram(0)

	return canvas
}

func initShaderProgram() {
	vertexShader := `
#version 330 core

uniform mat4 Transform;

layout (location = 0) in vec2 in_pos;
layout (location = 1) in vec4 in_color;
layout (location = 2) in vec2 in_texture_uv;

out vec2 frag_texture_uv;
out vec4 frag_color;

void main() {
	frag_texture_uv = in_texture_uv;
	frag_color = in_color;
    gl_Position = Transform * vec4(in_pos, 1, 1);
}
` + "\x00"

	fragmentShader := `
#version 330 core

uniform sampler2D Texture;

in vec2 frag_texture_uv;
in vec4 frag_color;

out vec4 outputColor;

void main() {
    outputColor = frag_color * texture(Texture, frag_texture_uv);
}
` + "\x00"

	var err error
	if program, err = platform.NewShaderProgram(vertexShader, fragmentShader); err != nil {
		log.Fatal("[canvas] unable to create shader:", err)
	}
}

func initUniforms() {
	uniTransform = gl.GetUniformLocation(program, gl.Str("Transform\x00"))
}

func (r *Render) initBuffers() {
	gl.GenVertexArrays(1, &r.vao)
	gl.GenBuffers(1, &r.vbo)
	gl.GenBuffers(1, &r.ebo)
}

func (r *Render) fillArrayBuffer() {
	gl.BindVertexArray(r.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(r.bucket.Data)*4, gl.Ptr(r.bucket.Data), gl.STATIC_DRAW)
	r.initAttributes()
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.BindVertexArray(0)
}

func (*Render) initAttributes() {
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(0))

	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 4, gl.FLOAT, false, 8*4, gl.PtrOffset(2*4))

	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(6*4))
}

func (r *Render) Draw(width, height float32) {
	// Initialize OpenGL state.
	r.prepare()
	r.configureTransform(width, height)

	// Here we will place our active texture.
	var activeTexture uint32

	// Convert our width/height to scaled values.
	width = width / r.State.Scale
	height = height / r.State.Scale

	// Draw all bucket units
	for _, unit := range r.bucket.Units {
		// Ignore out of bounds units.
		if r.isUnitOutOfBounds(unit, width, height) {
			continue
		}

		texture := unit.sp.Texture()

		// Sort of texture batching.
		// More effectively would be to merge all textures into one atlas and do not switch texture at all.
		if texture != activeTexture {
			if activeTexture != 0 && len(indicesCache) > 0 {
				r.flushIndices()
			}

			gl.BindTexture(gl.TEXTURE_2D, texture)
			activeTexture = texture
		}

		// Push data into the same indices slice to avoid unnecessary allocations.
		unit.pushIndices(&indicesCache)
	}

	// If we have something to draw - draw it.
	if activeTexture != 0 && len(indicesCache) > 0 {
		r.flushIndices()
	}

	// Major cleanup for OpenGL state.
	r.cleanup()
}

func (r *Render) prepare() {
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.BlendEquation(gl.FUNC_ADD)
	gl.UseProgram(program)
	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.ebo)
	gl.ActiveTexture(gl.TEXTURE0)
}

func (r *Render) configureTransform(width, height float32) {
	view := mgl32.Ortho(0, width, 0, height, -1, 1).Mul4(mgl32.Scale2D(r.State.Scale, r.State.Scale).Mat4())
	model := mgl32.Ident4().Mul4(mgl32.Translate2D(r.State.ShiftX, r.State.ShiftY).Mat4())
	mtxTransform := view.Mul4(model)
	gl.UniformMatrix4fv(uniTransform, 1, false, &mtxTransform[0])
}

func (r *Render) isUnitOutOfBounds(u unit, w, h float32) bool {
	bx1, by1, bx2, by2 := u.x1+r.State.ShiftX, u.y1+r.State.ShiftY, u.x2+r.State.ShiftX, u.y2+r.State.ShiftY
	return bx1 > w || by1 > h || bx2 < 0 || by2 < 0
}

func (r *Render) flushIndices() {
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indicesCache)*4, gl.Ptr(indicesCache), gl.STATIC_DRAW)
	gl.DrawElements(gl.TRIANGLES, int32(len(indicesCache)), gl.UNSIGNED_INT, gl.PtrOffset(0))
	indicesCache = indicesCache[:0]
}

func (*Render) cleanup() {
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.UseProgram(0)
	gl.Disable(gl.BLEND)
}
