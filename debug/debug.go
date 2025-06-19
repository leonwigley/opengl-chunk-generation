package debug

import (
	"fmt"
	"image"
	"os"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Debug struct {
	Enabled       bool
	FPS           float64
	frameCount    int
	lastFrameTime float64
	program       uint32
	fontTexture   uint32
	vao, vbo      uint32
	ortho         mgl32.Mat4
	window        *glfw.Window
	fontFace      font.Face
	fontDPI       float64
	fontSize      float64
}

func NewDebug(window *glfw.Window) (*Debug, error) {
	width, height := window.GetFramebufferSize()
	d := &Debug{
		Enabled:       false,
		FPS:           0,
		frameCount:    0,
		lastFrameTime: glfw.GetTime(),
		window:        window,
		ortho:         mgl32.Ortho(0, float32(width), 0, float32(height), -1, 1), // Standard y-axis
		fontDPI:       100,
		fontSize:      20,
	}

	// Load TTF font
	fontData, err := os.ReadFile("fonts/DejaVuSans.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to read font: %w", err)
	}
	fnt, err := truetype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}
	d.fontFace = truetype.NewFace(fnt, &truetype.Options{Size: d.fontSize, DPI: d.fontDPI})

	// Compile text shader
	program, err := createShaderProgram(textVertexShaderSource, textFragmentShaderSource)
	if err != nil {
		return nil, err
	}
	d.program = program

	// Setup text quad
	d.vao, d.vbo = setupTextQuad()

	// Update ortho projection on resize
	window.SetFramebufferSizeCallback(func(w *glfw.Window, width, height int) {
		if width == 0 || height == 0 {
			return
		}
		gl.Viewport(0, 0, int32(width), int32(height))
		d.ortho = mgl32.Ortho(0, float32(width), 0, float32(height), -1, 1)
	})

	return d, nil
}

func (d *Debug) Toggle() {
	d.Enabled = !d.Enabled
}

func (d *Debug) Update(deltaTime float32) {
	currentTime := glfw.GetTime()
	d.frameCount++
	if currentTime-d.lastFrameTime >= 1.0 {
		d.FPS = float64(d.frameCount) / (currentTime - d.lastFrameTime)
		d.frameCount = 0
		d.lastFrameTime = currentTime
	}
}

func (d *Debug) Render(playerPos mgl32.Vec3) {
	if !d.Enabled {
		return
	}
	gl.UseProgram(d.program)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.UniformMatrix4fv(gl.GetUniformLocation(d.program, gl.Str("projection\x00")), 1, false, &d.ortho[0])
	gl.BindVertexArray(d.vao)

	_, height := d.window.GetFramebufferSize()
	coords := fmt.Sprintf("X: %.1f Y: %.1f Z: %.1f", playerPos.X(), playerPos.Y(), playerPos.Z())
	fpsText := fmt.Sprintf("FPS: %.1f", d.FPS)
	d.renderText(coords, 10, float32(height)-50, float32(d.fontSize))
	d.renderText(fpsText, 10, float32(height)-100, float32(d.fontSize))
}

func (d *Debug) Cleanup() {
	gl.DeleteProgram(d.program)
	gl.DeleteTextures(1, &d.fontTexture)
	gl.DeleteVertexArrays(1, &d.vao)
	gl.DeleteBuffers(1, &d.vbo)
	if d.fontFace != nil {
		d.fontFace.Close()
	}
}

func (d *Debug) renderText(text string, x, y, size float32) {
	gl.UseProgram(d.program)
	scaleFactor := size / float32(d.fontSize) * 2

	drawer := &font.Drawer{
		Dst:  nil,
		Src:  image.White,
		Face: d.fontFace,
	}

	xPos := x
	for _, char := range text {
		bounds, advance, ok := d.fontFace.GlyphBounds(char)
		if !ok {
			xPos += float32(advance.Ceil()) * scaleFactor
			continue
		}
		w := (bounds.Max.X - bounds.Min.X).Ceil()
		h := (bounds.Max.Y - bounds.Min.Y).Ceil()
		if w <= 0 || h <= 0 {
			xPos += float32(advance.Ceil()) * scaleFactor
			continue
		}

		// Render glyph to image
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		drawer.Dst = img
		drawer.Dot = fixed.P(0, -bounds.Min.Y.Ceil())
		drawer.DrawString(string(char))

		// Upload glyph to texture
		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

		// Render quad
		offset := mgl32.Vec2{xPos, y}
		scale := mgl32.Vec2{float32(w) * scaleFactor, float32(h) * scaleFactor}
		gl.Uniform2fv(gl.GetUniformLocation(d.program, gl.Str("offset\x00")), 1, &offset[0])
		gl.Uniform2fv(gl.GetUniformLocation(d.program, gl.Str("scale\x00")), 1, &scale[0])
		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		// Clean up texture
		gl.DeleteTextures(1, &texture)

		// Advance x position
		xPos += float32(advance.Ceil()) * scaleFactor
	}
}

const textVertexShaderSource = `
#version 460 core
layout(location = 0) in vec2 pos;
layout(location = 1) in vec2 texCoord;
out vec2 TexCoord;
uniform mat4 projection;
uniform vec2 offset;
uniform vec2 scale;
void main() {
    gl_Position = projection * vec4(pos * scale + offset, 0.0, 1.0);
    TexCoord = texCoord;
}
`

const textFragmentShaderSource = `
#version 460 core
in vec2 TexCoord;
out vec4 fragColor;
uniform sampler2D textTexture;
void main() {
    vec4 color = texture(textTexture, TexCoord);
    if (color.a < 0.1) discard;
    fragColor = vec4(1.0, 1.0, 1.0, color.a); // White text
}
`

func createShaderProgram(vertexSrc, fragmentSrc string) (uint32, error) {
	vertexShader, err := compileShader(vertexSrc, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	fragmentShader, err := compileShader(fragmentSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}
	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)

	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength+1)
		gl.GetProgramInfoLog(prog, logLength, nil, &log[0])
		return 0, fmt.Errorf("failed to link program: %s", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)
	return prog, nil
}

func compileShader(src string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(src + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength+1)
		gl.GetShaderInfoLog(shader, logLength, nil, &log[0])
		return 0, fmt.Errorf("failed to compile shader: %s", log)
	}
	return shader, nil
}

func setupTextQuad() (uint32, uint32) {
	vertices := []float32{
		0, 0, 0, 1, // Bottom-left (flipped)
		1, 0, 1, 1, // Bottom-right (flipped)
		1, 1, 1, 0, // Top-right (flipped)
		0, 0, 0, 1, // Bottom-left (flipped)
		1, 1, 1, 0, // Top-right (flipped)
		0, 1, 0, 0, // Top-left (flipped)
	}
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)
	gl.BindVertexArray(0)
	return vao, vbo
}
