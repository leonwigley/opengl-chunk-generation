package main

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"
	"os"
	"runtime"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func init() {
	runtime.LockOSThread() // GLFW requires the main thread locked
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("runtime error: %v", err)
	}
}

func run() error {
	if err := glfw.Init(); err != nil {
		return fmt.Errorf("failed to initialize glfw: %w", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(800, 600, "Chunk Rendering", nil, nil)
	if err != nil {
		return err
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		return err
	}
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)


	program, err := createShaderProgram(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		return err
	}
	gl.UseProgram(program)

	texture, err := loadTexture("block.png")
	if err != nil {
		return err
	}
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.Uniform1i(gl.GetUniformLocation(program, gl.Str("texture1\x00")), 0)

	camera := NewCamera(mgl32.Vec3{0, 0, 10})
	projection := mgl32.Perspective(mgl32.DegToRad(45), 800.0/600.0, 0.1, 100.0)
	model := mgl32.Ident4()

	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	var (
		lastX, lastY = 400.0, 300.0
		firstMouse   = true
	)
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

	world := World{
		Chunks:      make(map[[2]int]*Chunk),
		ChunkRadius: 3,
	}
	initialChunk := NewFlatChunk()
	initialChunk.UploadMesh()
	world.Chunks[[2]int{0, 0}] = initialChunk
	world.UpdateChunks(camera.Position)

	for !window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := float32(currentTime - lastTime)
		lastTime = currentTime

		// Process keyboard input
		processInput(window, camera, deltaTime)

		// Rendering setup
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		view := camera.GetViewMatrix()

		gl.UseProgram(program)
		gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("projection\x00")), 1, false, &projection[0])
		gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("view\x00")), 1, false, &view[0])

		// Render all chunks
		for pos, chunk := range world.Chunks {
			model := mgl32.Translate3D(float32(pos[0]*ChunkSize), 0, float32(pos[1]*ChunkSize))
			gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("model\x00")), 1, false, &model[0])
			gl.BindVertexArray(chunk.VAO)
			gl.DrawArrays(gl.TRIANGLES, 0, chunk.VertexCount)
		}

		// Set lighting uniforms
		lightDirLoc := gl.GetUniformLocation(program, gl.Str("lightDir\x00"))
		viewPosLoc := gl.GetUniformLocation(program, gl.Str("viewPos\x00"))
		gl.Uniform3f(lightDirLoc, 0.5, -1.0, 0.3)
		gl.Uniform3f(viewPosLoc, camera.Position.X(), camera.Position.Y(), camera.Position.Z())

		window.SwapBuffers()
		glfw.PollEvents()
	}

	return nil
}

func processInput(window *glfw.Window, camera *Camera, deltaTime float32) {
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
}

// You can keep your existing vertexShaderSource and fragmentShaderSource strings here:

const vertexShaderSource = `
#version 460 core
layout(location = 0) in vec3 position;
layout(location = 1) in vec2 texCoord;
layout(location = 2) in vec3 normal;

out vec2 TexCoord;
out vec3 Normal;
out vec3 FragPos;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

void main() {
    FragPos = vec3(model * vec4(position, 1.0));
    Normal = mat3(transpose(inverse(model))) * normal;

    gl_Position = projection * view * vec4(FragPos, 1.0);
    TexCoord = texCoord;
}
`

const fragmentShaderSource = `
#version 460 core
in vec2 TexCoord;
in vec3 Normal;
in vec3 FragPos;

out vec4 fragColor;

uniform sampler2D texture1;
uniform vec3 lightDir;
uniform vec3 viewPos;

void main() {
    vec3 norm = normalize(Normal);
    vec3 lightDirection = normalize(-lightDir);

    float diff = max(dot(norm, lightDirection), 0.0);
    vec3 ambient = 0.1 * texture(texture1, TexCoord).rgb;
    vec3 diffuse = diff * texture(texture1, TexCoord).rgb;
    vec3 result = ambient + diffuse;

    fragColor = vec4(result, 1.0);
}
`

	program, err := createShaderProgram(vertexShader, fragmentShader)
	if err != nil {
		log.Fatalf("Shader error: %v", err)
	}
	gl.UseProgram(program)

	texture, err := loadTexture("block.png")
	if err != nil {
		log.Fatalf("texture error: %v", err)
	}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.Uniform1i(gl.GetUniformLocation(program, gl.Str("texture1\x00")), 0)

	// Camera setup
	camera := NewCamera(mgl32.Vec3{0, 0, 10})
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

	// Create world and one chunk
	world := World{
		Chunks:      make(map[[2]int]*Chunk),
		ChunkRadius: 3,
	}
	chunk := NewFlatChunk()
	chunk.UploadMesh()
	world.Chunks[[2]int{0, 0}] = chunk
	world.UpdateChunks(camera.Position)

	// Main loop
	for !window.ShouldClose() {
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

		for pos, chunk := range world.Chunks {
			model := mgl32.Translate3D(float32(pos[0]*ChunkSize), 0, float32(pos[1]*ChunkSize))
			gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("model\x00")), 1, false, &model[0])
			gl.BindVertexArray(chunk.VAO)
			gl.DrawArrays(gl.TRIANGLES, 0, chunk.VertexCount)
		}

		lightDirUniform := gl.GetUniformLocation(program, gl.Str("lightDir\x00"))
		viewPosUniform := gl.GetUniformLocation(program, gl.Str("viewPos\x00"))
		gl.Uniform3f(lightDirUniform, 0.5, -1.0, 0.3) // Example light direction (sunlight)
		gl.Uniform3f(viewPosUniform, camera.Position.X(), camera.Position.Y(), camera.Position.Z())

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

const ChunkSize = 16

type BlockID byte

const (
	BlockAir BlockID = iota
	BlockGrass
	BlockDirt
	BlockStone
)

type Chunk struct {
	Blocks      [ChunkSize][ChunkSize][ChunkSize]BlockID
	VAO         uint32
	VBO         uint32
	VertexCount int32
}

func NewFlatChunk() *Chunk {
	var c Chunk
	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			for y := 0; y < ChunkSize; y++ {
				if y == 3 {
					c.Blocks[x][y][z] = BlockGrass
				} else if y < 3 {
					c.Blocks[x][y][z] = BlockDirt
				} else {
					c.Blocks[x][y][z] = BlockAir
				}
			}
		}
	}	
	return &c
}

func (c *Chunk) GenerateMesh() []float32 {
	var mesh []float32
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkSize; y++ {
			for z := 0; z < ChunkSize; z++ {
				block := c.Blocks[x][y][z]
				if block == BlockAir {
					continue
				}
				// Right
				if x == ChunkSize-1 || c.Blocks[x+1][y][z] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "right")...)
				}
				// Left
				if x == 0 || c.Blocks[x-1][y][z] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "left")...)
				}
				// Top
				if y == ChunkSize-1 || c.Blocks[x][y+1][z] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "top")...)
				}
				// Bottom
				if y == 0 || c.Blocks[x][y-1][z] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "bottom")...)
				}
				// Front
				if z == ChunkSize-1 || c.Blocks[x][y][z+1] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "front")...)
				}
				// Back
				if z == 0 || c.Blocks[x][y][z-1] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "back")...)
				}
			}
		}
	}
	return mesh
}

func (c *Chunk) UploadMesh() {
	mesh := c.GenerateMesh()
	vao, vbo := UploadMesh(mesh)
	c.VAO = vao
	c.VBO = vbo
	c.VertexCount = int32(len(mesh) / 8) // 8 floats per vertex: x,y,z,nx,ny,nz,u,v
}

func UploadMesh(vertices []float32) (vao, vbo uint32) {
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	// normal attribute
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	// texCoord attribute
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(1)

	return
}

func createFace(x, y, z float32, face string) []float32 {
	var nx, ny, nz float32
	switch face {
	case "right":
		nx, ny, nz = 1, 0, 0
	case "left":
		nx, ny, nz = -1, 0, 0
	case "top":
		nx, ny, nz = 0, 1, 0
	case "bottom":
		nx, ny, nz = 0, -1, 0
	case "front":
		nx, ny, nz = 0, 0, 1
	case "back":
		nx, ny, nz = 0, 0, -1
	}

	switch face {
	case "right":
		return []float32{
			x + 1, y, z, nx, ny, nz, 1, 0,
			x + 1, y + 1, z, nx, ny, nz, 1, 1,
			x + 1, y + 1, z + 1, nx, ny, nz, 0, 1,

			x + 1, y, z, nx, ny, nz, 1, 0,
			x + 1, y + 1, z + 1, nx, ny, nz, 0, 1,
			x + 1, y, z + 1, nx, ny, nz, 0, 0,
		}
	case "left":
		return []float32{
			x, y, z, nx, ny, nz, 0, 0,
			x, y + 1, z + 1, nx, ny, nz, 1, 1,
			x, y + 1, z, nx, ny, nz, 0, 1,

			x, y, z, nx, ny, nz, 0, 0,
			x, y, z + 1, nx, ny, nz, 1, 0,
			x, y + 1, z + 1, nx, ny, nz, 1, 1,
		}
	case "top":
		return []float32{
			x, y + 1, z, nx, ny, nz, 0, 0,
			x + 1, y + 1, z, nx, ny, nz, 1, 0,
			x + 1, y + 1, z + 1, nx, ny, nz, 1, 1,

			x, y + 1, z, nx, ny, nz, 0, 0,
			x + 1, y + 1, z + 1, nx, ny, nz, 1, 1,
			x, y + 1, z + 1, nx, ny, nz, 0, 1,
		}
	case "bottom":
		return []float32{
			x, y, z, nx, ny, nz, 0, 0,
			x + 1, y, z + 1, nx, ny, nz, 1, 1,
			x + 1, y, z, nx, ny, nz, 1, 0,

			x, y, z, nx, ny, nz, 0, 0,
			x, y, z + 1, nx, ny, nz, 0, 1,
			x + 1, y, z + 1, nx, ny, nz, 1, 1,
		}
	case "front":
		return []float32{
			x, y, z + 1, nx, ny, nz, 0, 0,
			x + 1, y + 1, z + 1, nx, ny, nz, 1, 1,
			x + 1, y, z + 1, nx, ny, nz, 1, 0,

			x, y, z + 1, nx, ny, nz, 0, 0,
			x, y + 1, z + 1, nx, ny, nz, 0, 1,
			x + 1, y + 1, z + 1, nx, ny, nz, 1, 1,
		}
	case "back":
		return []float32{
			x, y, z, nx, ny, nz, 0, 0,
			x + 1, y, z, nx, ny, nz, 1, 0,
			x + 1, y + 1, z, nx, ny, nz, 1, 1,

			x, y, z, nx, ny, nz, 0, 0,
			x + 1, y + 1, z, nx, ny, nz, 1, 1,
			x, y + 1, z, nx, ny, nz, 0, 1,
		}
	}
	return nil
}

// ---- Helper functions: texture loading and shader compilation ----

func loadTexture(path string) (uint32, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	for y := 0; y < rgba.Bounds().Dy(); y++ {
		for x := 0; x < rgba.Bounds().Dx(); x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y),
		0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

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

// --- Simple FPS camera for navigation ---
type Camera struct {
	Position mgl32.Vec3
	Front    mgl32.Vec3
	Up       mgl32.Vec3
	Right    mgl32.Vec3
	WorldUp  mgl32.Vec3

	Yaw   float32
	Pitch float32

	MovementSpeed float32
	MouseSens     float32
}

func NewCamera(position mgl32.Vec3) *Camera {
	c := &Camera{
		Position:      position,
		Front:         mgl32.Vec3{0, 0, -1},
		Up:            mgl32.Vec3{0, 1, 0},
		WorldUp:       mgl32.Vec3{0, 1, 0},
		Yaw:           -90,
		Pitch:         0,
		MovementSpeed: 5,
		MouseSens:     0.1,
	}
	c.updateCameraVectors()
	return c
}

func (c *Camera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position, c.Position.Add(c.Front), c.Up)
}
func (c *Camera) ProcessKeyboard(direction string, deltaTime float32) {
    velocity := c.MovementSpeed * deltaTime
    switch direction {
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
    xoffset *= float64(c.MouseSens)
    yoffset *= float64(c.MouseSens)

    c.Yaw += float32(xoffset)
    c.Pitch += float32(yoffset)

    if c.Pitch > 89 {
        c.Pitch = 89
    }
    if c.Pitch < -89 {
        c.Pitch = -89
    }

    c.updateCameraVectors()
}

func (c *Camera) updateCameraVectors() {
    front := mgl32.Vec3{
        float32(math.Cos(float64(mgl32.DegToRad(c.Yaw))) * math.Cos(float64(mgl32.DegToRad(c.Pitch)))),
        float32(math.Sin(float64(mgl32.DegToRad(c.Pitch)))),
        float32(math.Sin(float64(mgl32.DegToRad(c.Yaw))) * math.Cos(float64(mgl32.DegToRad(c.Pitch)))),
    }
    c.Front = front.Normalize()
    c.Right = c.Front.Cross(c.WorldUp).Normalize()
    c.Up = c.Right.Cross(c.Front).Normalize()
}

type World struct {
	Chunks      map[[2]int]*Chunk
	ChunkRadius int
}

func (w *World) UpdateChunks(playerPos mgl32.Vec3) {
	playerChunkX := int(math.Floor(float64(playerPos.X() / ChunkSize)))
	playerChunkZ := int(math.Floor(float64(playerPos.Z() / ChunkSize)))

	// Load new chunks around player
	for x := playerChunkX - w.ChunkRadius; x <= playerChunkX+w.ChunkRadius; x++ {
		for z := playerChunkZ - w.ChunkRadius; z <= playerChunkZ+w.ChunkRadius; z++ {
			key := [2]int{x, z}
			if _, exists := w.Chunks[key]; !exists {
				chunk := NewFlatChunk() // You can make it more complex per (x,z)
				chunk.UploadMesh()
				w.Chunks[key] = chunk
			}
		}
	}

	// Unload chunks far away (optional)
	for key := range w.Chunks {
		x, z := key[0], key[1]
		if x < playerChunkX-w.ChunkRadius || x > playerChunkX+w.ChunkRadius ||
			z < playerChunkZ-w.ChunkRadius || z > playerChunkZ+w.ChunkRadius {
			// Delete OpenGL buffers
			gl.DeleteVertexArrays(1, &w.Chunks[key].VAO)
			gl.DeleteBuffers(1, &w.Chunks[key].VBO)
			delete(w.Chunks, key)
		}
	}
}

func isSolid(blocks [][][]bool, x, y, z int) bool {
    sizeX := len(blocks)
    sizeY := len(blocks[0])
    sizeZ := len(blocks[0][0])

    if x < 0 || x >= sizeX || y < 0 || y >= sizeY || z < 0 || z >= sizeZ {
        return false
    }
    return blocks[x][y][z]
}

var faceOffsets = map[string][3]int{
    "right":  {1, 0, 0},
    "left":   {-1, 0, 0},
    "top":    {0, 1, 0},
    "bottom": {0, -1, 0},
    "front":  {0, 0, 1},
    "back":   {0, 0, -1},
}

func BuildMesh(blocks [][][]bool) []float32 {
    var mesh []float32
    sizeX := len(blocks)
    sizeY := len(blocks[0])
    sizeZ := len(blocks[0][0])

    for x := 0; x < sizeX; x++ {
        for y := 0; y < sizeY; y++ {
            for z := 0; z < sizeZ; z++ {
                if !blocks[x][y][z] {
                    continue
                }

                for face, offset := range faceOffsets {
                    nx := x + offset[0]
                    ny := y + offset[1]
                    nz := z + offset[2]

                    if !isSolid(blocks, nx, ny, nz) {
                        // Add the face only if neighbor is empty or out of bounds
                        mesh = append(mesh, createFace(float32(x), float32(y), float32(z), face)...)
                    }
                }
            }
        }
    }
    return mesh
}
