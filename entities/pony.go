package entities

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Pony represents a pony entity with multiple parts (body, head, legs, etc.).
type Pony struct {
	Position  mgl32.Vec3 // World position of the pony's center
	Velocity  mgl32.Vec3 // Movement velocity
	Height    float32    // Overall height (bounding box, ~body + head)
	Width     float32    // Overall width (bounding box, ~body length)
	Color     mgl32.Vec3 // Default color (unused, parts have own colors)
	Parts     []PonyPart // List of body parts (body, head, legs, etc.)
	Program   uint32     // OpenGL shader program
	AnimState AnimState  // Animation state (for future use)
}

// PonyPart represents a single part of the pony (e.g., body, head).
type PonyPart struct {
	VAO, VBO    uint32     // OpenGL buffers for cube mesh
	VertexCount int32      // Number of vertices (36 for cube)
	ModelMatrix mgl32.Mat4 // Computed during rendering
	Color       mgl32.Vec3 // RGB color for block rendering
	Pivot       mgl32.Vec3 // Rotation center (for future animation)
	Scale       mgl32.Vec3 // Size (width, height, depth)
	Offset      mgl32.Vec3 // Position relative to Pony.Position
}

// AnimState holds animation data (for future use).
type AnimState struct {
	Time      float32 // Animation timer
	WalkCycle float32 // 0 to 1 for walk cycle
}

// NewPony creates a new pony with predefined parts (body, head, neck, tail, legs).
func NewPony(pos, vel mgl32.Vec3) (*Pony, error) {
	// Initialize pony with bounding box (Height: body + head, Width: body length)
	pony := &Pony{
		Position:  pos,
		Velocity:  vel,
		Height:    2.0,                       // Body (1) + head (0.8) + offset
		Width:     2.0,                       // Body length
		Color:     mgl32.Vec3{0.6, 0.4, 0.2}, // Default brown (unused)
		AnimState: AnimState{Time: 0, WalkCycle: 0},
	}

	// Create shader program (reuse logic from debug/debug.go)
	vertexShader := `
        #version 460 core
        layout(location = 0) in vec3 pos;
        uniform mat4 model;
        uniform mat4 view;
        uniform mat4 projection;
        void main() {
            gl_Position = projection * view * model * vec4(pos, 1.0);
        }
    `
	fragmentShader := `
        #version 460 core
        out vec4 fragColor;
        uniform vec3 partColor;
        void main() {
            fragColor = vec4(partColor, 1.0);
        }
    `
	program, err := createShaderProgram(vertexShader, fragmentShader)
	if err != nil {
		return nil, fmt.Errorf("failed to create shader: %w", err)
	}
	pony.Program = program

	// Define pony parts with sizes, offsets, and colors
	parts := []PonyPart{
		// Body: Centered at Pony.Position
		{
			Scale:  mgl32.Vec3{2, 1, 0.8}, // Long, high, wide
			Offset: mgl32.Vec3{0, 0, 0},
			Color:  mgl32.Vec3{0.6, 0.4, 0.2}, // Brown
			Pivot:  mgl32.Vec3{0, 0, 0},       // Center (for future rotation)
		},
		// Head: Forward and up from body
		{
			Scale:  mgl32.Vec3{0.8, 0.8, 0.8},
			Offset: mgl32.Vec3{1.2, 0.8, 0},
			Color:  mgl32.Vec3{0.7, 0.5, 0.3}, // Light brown
			Pivot:  mgl32.Vec3{0, 0, 0},
		},
		// Neck: Between body and head
		{
			Scale:  mgl32.Vec3{0.4, 0.6, 0.4},
			Offset: mgl32.Vec3{0.8, 0.6, 0},
			Color:  mgl32.Vec3{0.6, 0.4, 0.2}, // Brown
			Pivot:  mgl32.Vec3{0, 0, 0},
		},
		// Tail: At rear of body
		{
			Scale:  mgl32.Vec3{0.3, 0.8, 0.3},
			Offset: mgl32.Vec3{-0.8, 0.4, 0},
			Color:  mgl32.Vec3{1, 1, 0}, // Yellow
			Pivot:  mgl32.Vec3{0, 0, 0},
		},
		// Front-left leg
		{
			Scale:  mgl32.Vec3{0.4, 1.2, 0.4},
			Offset: mgl32.Vec3{0.8, -0.6, 0.3},
			Color:  mgl32.Vec3{0.5, 0.3, 0.1}, // Dark brown
			Pivot:  mgl32.Vec3{0, 0.6, 0},     // Top of leg (for animation)
		},
		// Front-right leg
		{
			Scale:  mgl32.Vec3{0.4, 1.2, 0.4},
			Offset: mgl32.Vec3{0.8, -0.6, -0.3},
			Color:  mgl32.Vec3{0.5, 0.3, 0.1},
			Pivot:  mgl32.Vec3{0, 0.6, 0},
		},
		// Back-left leg
		{
			Scale:  mgl32.Vec3{0.4, 1.2, 0.4},
			Offset: mgl32.Vec3{-0.8, -0.6, 0.3},
			Color:  mgl32.Vec3{0.5, 0.3, 0.1},
			Pivot:  mgl32.Vec3{0, 0.6, 0},
		},
		// Back-right leg
		{
			Scale:  mgl32.Vec3{0.4, 1.2, 0.4},
			Offset: mgl32.Vec3{-0.8, -0.6, -0.3},
			Color:  mgl32.Vec3{0.5, 0.3, 0.1},
			Pivot:  mgl32.Vec3{0, 0.6, 0},
		},
	}

	// Set up cube mesh for each part
	for i := range parts {
		vao, vbo := setupCubeMesh()
		parts[i].VAO = vao
		parts[i].VBO = vbo
		parts[i].VertexCount = 36 // 6 faces * 2 triangles * 3 vertices
	}
	pony.Parts = parts

	return pony, nil
}

// Render draws the pony using OpenGL (stub, integrate with main.go).
func (p *Pony) Render(view, projection mgl32.Mat4) {
	gl.UseProgram(p.Program)
	viewLoc := gl.GetUniformLocation(p.Program, gl.Str("view\x00"))
	projLoc := gl.GetUniformLocation(p.Program, gl.Str("projection\x00"))
	modelLoc := gl.GetUniformLocation(p.Program, gl.Str("model\x00"))
	colorLoc := gl.GetUniformLocation(p.Program, gl.Str("partColor\x00"))
	gl.UniformMatrix4fv(viewLoc, 1, false, &view[0])
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	for _, part := range p.Parts {
		// Compute model matrix: translate to position + offset, apply scale
		model := mgl32.Translate3D(p.Position.X()+part.Offset.X(), p.Position.Y()+part.Offset.Y(), p.Position.Z()+part.Offset.Z()).
			Mul4(mgl32.Scale3D(part.Scale.X(), part.Scale.Y(), part.Scale.Z()))
		part.ModelMatrix = model

		// Set uniforms
		gl.UniformMatrix4fv(modelLoc, 1, false, &part.ModelMatrix[0])
		gl.Uniform3fv(colorLoc, 1, &part.Color[0])

		// Draw part
		gl.BindVertexArray(part.VAO)
		gl.DrawElements(gl.TRIANGLES, part.VertexCount, gl.UNSIGNED_INT, nil)
		gl.BindVertexArray(0)
	}
}

// setupCubeMesh creates a cube mesh for a PonyPart (VAO, VBO, indices).
func setupCubeMesh() (uint32, uint32) {
	// Define cube vertices (1x1x1, centered at origin)
	vertices := []float32{
		// Front face
		-0.5, -0.5, 0.5, 0.5, -0.5, 0.5, 0.5, 0.5, 0.5, -0.5, 0.5, 0.5,
		// Back face
		-0.5, -0.5, -0.5, -0.5, 0.5, -0.5, 0.5, 0.5, -0.5, 0.5, -0.5, -0.5,
		// Left face
		-0.5, -0.5, -0.5, -0.5, -0.5, 0.5, -0.5, 0.5, 0.5, -0.5, 0.5, -0.5,
		// Right face
		0.5, -0.5, -0.5, 0.5, 0.5, -0.5, 0.5, 0.5, 0.5, 0.5, -0.5, 0.5,
		// Top face
		-0.5, 0.5, -0.5, -0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, 0.5, -0.5,
		// Bottom face
		-0.5, -0.5, -0.5, 0.5, -0.5, -0.5, 0.5, -0.5, 0.5, -0.5, -0.5, 0.5,
	}

	// Define indices for 12 triangles (6 faces * 2 triangles)
	indices := []uint32{
		// Front
		0, 1, 2, 2, 3, 0,
		// Back
		4, 5, 6, 6, 7, 4,
		// Left
		8, 9, 10, 10, 11, 8,
		// Right
		12, 13, 14, 14, 15, 12,
		// Top
		16, 17, 18, 18, 19, 16,
		// Bottom
		20, 21, 22, 22, 23, 20,
	}

	// Set up VAO, VBO, and EBO
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	// Upload vertices
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Upload indices
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.BindVertexArray(0)
	return vao, vbo
}

// createShaderProgram compiles vertex and fragment shaders (adapted from debug/debug.go).
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

// compileShader compiles a single shader (adapted from debug/debug.go).
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
