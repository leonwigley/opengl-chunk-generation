package debug

import (
	"fmt"
	"image"
	_ "image/png"
	"os"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
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
}

func NewDebug(window *glfw.Window) (*Debug, error) {
	width, height := window.GetFramebufferSize()
	d := &Debug{
		Enabled:       false,
		FPS:           0,
		frameCount:    0,
		lastFrameTime: glfw.GetTime(),
		window:        window,
		ortho:         mgl32.Ortho(0, float32(width), 0, float32(height), -1, 1),
	}

	// Compile text shader
	program, err := createShaderProgram(textVertexShaderSource, textFragmentShaderSource)
	if err != nil {
		return nil, err
	}
	d.program = program

	// Load font texture
	fontTexture, err := loadTexture("textures/font.png")
	if err != nil {
		return nil, err
	}
	d.fontTexture = fontTexture

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
	gl.BindTexture(gl.TEXTURE_2D, d.fontTexture)
	gl.UniformMatrix4fv(gl.GetUniformLocation(d.program, gl.Str("projection\x00")), 1, false, &d.ortho[0])
	gl.BindVertexArray(d.vao)

	width, height := d.window.GetFramebufferSize()
	coords := fmt.Sprintf("X: %.1f Y: %.1f Z: %.1f", playerPos.X(), playerPos.Y(), playerPos.Z())
	fpsText := fmt.Sprintf("FPS: %.1f", d.FPS)
	renderText(coords, 10, float32(height)-20, 16, d.program)
	renderText(fpsText, 10, float32(height)-40, 16, d.program)
}

func (d *Debug) Cleanup() {
	gl.DeleteProgram(d.program)
	gl.DeleteTextures(1, &d.fontTexture)
	gl.DeleteVertexArrays(1, &d.vao)
	gl.DeleteBuffers(1, &d.vbo)
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

func loadTexture(path string) (uint32, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("failed to open texture: %w", err)
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, fmt.Errorf("failed to decode texture: %w", err)
	}

	rgba := image.NewRGBA(img.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			rgba.Set(x, img.Bounds().Dy()-y-1, img.At(x, y))
		}
	}

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y),
		0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	return texture, nil
}

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
		0, 0, 0, 1, // Bottom-left
		1, 0, 1, 1, // Bottom-right
		1, 1, 1, 0, // Top-right
		0, 0, 0, 1, // Bottom-left
		1, 1, 1, 0, // Top-right
		0, 1, 0, 0, // Top-left
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

func renderText(text string, x, y, size float32, program uint32) {
	gl.UseProgram(program)
	charWidth := size
	charHeight := size
	for i, char := range text {
		ascii := int(char)
		u0 := float32((ascii % 16) / 16.0)
		v0 := float32((ascii / 16) / 16.0)
		u1 := u0 + 1.0/16.0
		v1 := v0 + 1.0/16.0
		offset := mgl32.Vec2{x + float32(i)*charWidth, y}
		scale := mgl32.Vec2{charWidth, charHeight}
		gl.Uniform2fv(gl.GetUniformLocation(program, gl.Str("offset\x00")), 1, &offset[0])
		gl.Uniform2fv(gl.GetUniformLocation(program, gl.Str("scale\x00")), 1, &scale[0])
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
	}
}
