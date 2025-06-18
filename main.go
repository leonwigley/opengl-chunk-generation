package main

import (
	"fmt"
	"log"
	"math"
	"runtime"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func init() {
	runtime.LockOSThread() // Required for GLFW
}

func main() {
	runtime.LockOSThread()

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(800, 600, "something", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}
	gl.Enable(gl.DEPTH_TEST)

	// Create shader
	vertexShader := `
    #version 460
    in vec3 position;
    uniform mat4 model;
    uniform mat4 view;
    uniform mat4 projection;
    void main() {
        gl_Position = projection * view * model * vec4(position, 1.0);
    }`
	fragmentShader := `
    #version 460
    out vec4 fragColor;
    void main() {
        fragColor = vec4(0.5, 0.8, 0.2, 1.0);
    }`
	program, err := createShaderProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Fatalf("Shader error: %v", err)
	}
	gl.UseProgram(program)

	// Cube data
	vertices := []float32{
		-0.5, -0.5, -0.5, 0.5, -0.5, -0.5, 0.5, 0.5, -0.5, -0.5, 0.5, -0.5,
		-0.5, -0.5, 0.5, 0.5, -0.5, 0.5, 0.5, 0.5, 0.5, -0.5, 0.5, 0.5,
	}
	indices := []uint32{
		0, 1, 2, 2, 3, 0,
		4, 5, 6, 6, 7, 4,
		0, 4, 7, 7, 3, 0,
		1, 5, 6, 6, 2, 1,
		3, 2, 6, 6, 7, 3,
		0, 1, 5, 5, 4, 0,
	}

	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	gl.EnableVertexAttribArray(0)

	// Camera setup
	camera := NewCamera(mgl32.Vec3{0, 0, 3})
	projection := mgl32.Perspective(mgl32.DegToRad(45), 800.0/600.0, 0.1, 100.0)
	model := mgl32.Ident4()

	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	var lastX, lastY float64 = 400, 300
	firstMouse := true
	window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		if firstMouse {
			lastX, lastY = xpos, ypos
			firstMouse = false
		}
		xoffset := xpos - lastX
		yoffset := ypos - lastY
		lastX = xpos
		lastY = ypos
		camera.ProcessMouse(xoffset, yoffset)
	})

	lastTime := glfw.GetTime()

	// Main loop
	for !window.ShouldClose() {
		// Timing
		currentTime := glfw.GetTime()
		deltaTime := float32(currentTime - lastTime)
		lastTime = currentTime

		// Input
		if window.GetKey(glfw.KeyEscape) == glfw.Press {
			window.SetShouldClose(true)
		}
		if window.GetKey(glfw.KeyW) == glfw.Press {
			camera.ProcessKeyboard("forward", deltaTime)
		}
		if window.GetKey(glfw.KeyS) == glfw.Press {
			camera.ProcessKeyboard("backward", deltaTime)
		}
		if window.GetKey(glfw.KeyA) == glfw.Press {
			camera.ProcessKeyboard("left", deltaTime)
		}
		if window.GetKey(glfw.KeyD) == glfw.Press {
			camera.ProcessKeyboard("right", deltaTime)
		}

		// Rendering
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		view := camera.GetViewMatrix()

		gl.UseProgram(program)
		gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("projection\x00")), 1, false, &projection[0])
		gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("view\x00")), 1, false, &view[0])
		gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("model\x00")), 1, false, &model[0])

		gl.BindVertexArray(vao)
		gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, nil)

		window.SwapBuffers()
		glfw.PollEvents()
	}
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
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		logStr := make([]byte, logLength)
		gl.GetProgramInfoLog(program, logLength, nil, &logStr[0])
		return 0, fmt.Errorf("failed to link program: %v", string(logStr))
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)
	return program, nil
}

type Camera struct {
	Position mgl32.Vec3
	Front    mgl32.Vec3
	Up       mgl32.Vec3
	Right    mgl32.Vec3
	WorldUp  mgl32.Vec3

	Yaw   float32
	Pitch float32

	Speed       float32
	Sensitivity float32
}

func NewCamera(position mgl32.Vec3) *Camera {
	cam := &Camera{
		Position:    position,
		WorldUp:     mgl32.Vec3{0, 1, 0},
		Yaw:         -90,
		Pitch:       0,
		Speed:       3.0,
		Sensitivity: 0.1,
	}
	cam.updateVectors()
	return cam
}

func (c *Camera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position, c.Position.Add(c.Front), c.Up)
}

func (c *Camera) ProcessKeyboard(dir string, deltaTime float32) {
	velocity := c.Speed * deltaTime
	switch dir {
	case "forward":
		c.Position = c.Position.Add(c.Front.Mul(velocity))
	case "backward":
		c.Position = c.Position.Sub(c.Front.Mul(velocity))
	case "left":
		c.Position = c.Position.Sub(c.Right.Mul(velocity))
	case "right":
		c.Position = c.Position.Add(c.Right.Mul(velocity))
	}
}

func (c *Camera) ProcessMouse(xoffset, yoffset float64) {
	c.Yaw += float32(xoffset) * c.Sensitivity
	c.Pitch -= float32(yoffset) * c.Sensitivity // invert y

	if c.Pitch > 89 {
		c.Pitch = 89
	}
	if c.Pitch < -89 {
		c.Pitch = -89
	}
	c.updateVectors()
}

func (c *Camera) updateVectors() {
	yawRad := mgl32.DegToRad(c.Yaw)
	pitchRad := mgl32.DegToRad(c.Pitch)

	front := mgl32.Vec3{
		float32(math.Cos(float64(yawRad)) * math.Cos(float64(pitchRad))),
		float32(math.Sin(float64(pitchRad))),
		float32(math.Sin(float64(yawRad)) * math.Cos(float64(pitchRad))),
	}
	c.Front = front.Normalize()
	c.Right = c.Front.Cross(c.WorldUp).Normalize()
	c.Up = c.Right.Cross(c.Front).Normalize()
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	cSource, free := gl.Strs(source + "\x00") // Convert to C string and null-terminate
	gl.ShaderSource(shader, 1, cSource, nil)
	free() // Free C memory after use
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		logStr := make([]byte, logLength)
		gl.GetShaderInfoLog(shader, logLength, nil, &logStr[0])
		return 0, fmt.Errorf("failed to compile shader: %v", string(logStr))
	}
	return shader, nil
}
