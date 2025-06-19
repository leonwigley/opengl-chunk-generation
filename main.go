package main

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"something/debug"
	"something/player"
	"something/world"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func init() {
	runtime.LockOSThread()
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
	glfw.WindowHint(glfw.Resizable, glfw.True)

	window, err := glfw.CreateWindow(800, 600, "Leon's Unreal Forge", nil, nil)
	if err != nil {
		return err
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(0)
	window.SetSizeLimits(400, 300, glfw.DontCare, glfw.DontCare)

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

	texture, err := loadTexture("textures/grass.png")
	if err != nil {
		return err
	}

	debugMenu, err := debug.NewDebug(window)
	if err != nil {
		return err
	}
	defer debugMenu.Cleanup()

	player := player.NewPlayer(mgl32.Vec3{0, 10, 0})
	width, height := window.GetFramebufferSize()
	projection := mgl32.Perspective(mgl32.DegToRad(45), float32(width)/float32(height), 0.1, 100.0)

	cursorCaptured := true
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	var (
		lastX, lastY = 400.0, 300.0
		firstMouse   = true
	)
	window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		if !cursorCaptured {
			return
		}
		if firstMouse {
			lastX, lastY = xpos, ypos
			firstMouse = false
		}
		xoffset := xpos - lastX
		yoffset := lastY - ypos
		lastX = xpos
		lastY = ypos
		player.Camera.ProcessMouse(xoffset, yoffset)
	})

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if key == glfw.KeyEscape && action == glfw.Press {
			cursorCaptured = !cursorCaptured
			if cursorCaptured {
				window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
				firstMouse = true
			} else {
				window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
			}
		}
		if key == glfw.KeyF1 && action == glfw.Press {
			debugMenu.Toggle()
		}
		if key == glfw.KeyQ && action == glfw.Press {
			w.SetShouldClose(true)
		}
	})

	lastTime := glfw.GetTime()
	gameWorld := world.World{
		Chunks:      make(map[[2]int]*world.Chunk),
		ChunkRadius: 3,
	}
	initialChunk := world.NewChunk(0, 0)
	initialChunk.UploadMesh()
	gameWorld.Chunks[[2]int{0, 0}] = initialChunk
	gameWorld.UpdateChunks(player.Camera.Position)

	for !window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := float32(currentTime - lastTime)
		lastTime = currentTime

		player.Update(window, &gameWorld, deltaTime)
		debugMenu.Update(deltaTime)
		gameWorld.UpdateChunks(player.Camera.Position)

		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.UseProgram(program)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("projection\x00")), 1, false, &projection[0])
		viewMatrix := player.Camera.GetViewMatrix()
		gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("view\x00")), 1, false, &viewMatrix[0])
		for pos, chunk := range gameWorld.Chunks {
			model := mgl32.Translate3D(float32(pos[0]*world.ChunkSize), 0, float32(pos[1]*world.ChunkSize))
			gl.UniformMatrix4fv(gl.GetUniformLocation(program, gl.Str("model\x00")), 1, false, &model[0])
			gl.BindVertexArray(chunk.VAO)
			gl.DrawArrays(gl.TRIANGLES, 0, chunk.VertexCount)
		}
		gl.Uniform3f(gl.GetUniformLocation(program, gl.Str("lightDir\x00")), 0.5, -1.0, 0.3)
		gl.Uniform3f(gl.GetUniformLocation(program, gl.Str("viewPos\x00")), player.Camera.Position.X(), player.Camera.Position.Y(), player.Camera.Position.Z())

		debugMenu.Render(player.Position)

		window.SwapBuffers()
		glfw.PollEvents()
	}
	return nil
}

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
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
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
