package main

import (
	"fmt" // Added
	"log"
	"runtime"
	"something/debug"
	"something/entities"
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

	window, err := glfw.CreateWindow(500, 500, "WIP", nil, nil)
	if err != nil {
		return err
	}
	window.MakeContextCurrent()
	window.SetSizeLimits(400, 300, glfw.DontCare, glfw.DontCare)

	if err := gl.Init(); err != nil {
		return err
	}
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	debugMenu, err := debug.NewDebug(window)
	if err != nil {
		return err
	}
	defer debugMenu.Cleanup()

	gameWorld := world.World{
		Chunks:      make(map[[2]int]*world.Chunk),
		ChunkRadius: 3,
	}
	if err := gameWorld.Init(); err != nil {
		return err
	}
	defer gameWorld.Cleanup()
	initialChunk := world.NewChunk(0, 0)
	initialChunk.UploadMesh()
	gameWorld.Chunks[[2]int{0, 0}] = initialChunk

	// Temporary fixed spawn (remove once GetSurfaceHeight is verified)
	player := player.NewPlayer(mgl32.Vec3{0, 10, 0})
	pony, err := entities.NewPony(mgl32.Vec3{0, 10, 0}, mgl32.Vec3{0, 0, 0})
	if err != nil {
		return err
	}

	width, height := window.GetSize()
	projection := mgl32.Perspective(mgl32.DegToRad(45), float32(width)/float32(height), 0.1, 100.0)

	cursorCaptured := true
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	lastX, lastY := float64(width)/2, float64(height)/2
	firstMouse := true

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
		lastX, lastY = xpos, ypos
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
	for !window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := float32(currentTime - lastTime)
		lastTime = currentTime

		player.Update(window, &gameWorld, deltaTime)
		debugMenu.Update(deltaTime)
		gameWorld.UpdateChunks(player.Camera.Position)

		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		view := player.Camera.GetViewMatrix()
		gameWorld.Render(view, projection, player.Camera.Position)
		pony.Render(view, projection)
		gl.BindVertexArray(0)
		debugMenu.Render(player.Position)

		window.SwapBuffers()
		glfw.PollEvents()
	}
	return nil
}
