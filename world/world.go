package world

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"

	"something/block"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// ChunkSize defines the dimensions of a chunk (16x16x16).
const ChunkSize = 16

// World manages chunks and rendering resources.
type World struct {
	Chunks      map[[2]int]*Chunk
	ChunkRadius int
	Program     uint32 // Chunk shader
	Texture     uint32 // Grass texture
}

// Init initializes the world's shader and texture.
func (w *World) Init() error {
	var err error
	w.Program, err = createShaderProgram(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		return err
	}
	w.Texture, err = loadTexture("textures/grass.png")
	if err != nil {
		return err
	}
	return nil
}

// UpdateChunks loads/unloads chunks based on player position.
func (w *World) UpdateChunks(playerPos mgl32.Vec3) {
	playerChunkX := int(math.Floor(float64(playerPos.X() / float32(ChunkSize))))
	playerChunkZ := int(math.Floor(float64(playerPos.Z() / float32(ChunkSize))))
	for x := playerChunkX - w.ChunkRadius; x <= playerChunkX+w.ChunkRadius; x++ {
		for z := playerChunkZ - w.ChunkRadius; z <= playerChunkZ+w.ChunkRadius; z++ {
			key := [2]int{x, z}
			if _, exists := w.Chunks[key]; !exists {
				chunk := NewChunk(int32(x), int32(z))
				chunk.UploadMesh()
				w.Chunks[key] = chunk
			}
		}
	}
	for key := range w.Chunks {
		x, z := key[0], key[1]
		if x < playerChunkX-w.ChunkRadius || x > playerChunkX+w.ChunkRadius ||
			z < playerChunkZ-w.ChunkRadius || z > playerChunkZ+w.ChunkRadius {
			w.Chunks[key].Cleanup()
			delete(w.Chunks, key)
		}
	}
}

// Render draws all chunks using the world's shader and texture.
func (w *World) Render(view, projection mgl32.Mat4, viewPos mgl32.Vec3) {
	gl.UseProgram(w.Program)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, w.Texture)
	gl.UniformMatrix4fv(gl.GetUniformLocation(w.Program, gl.Str("view\x00")), 1, false, &view[0])
	gl.UniformMatrix4fv(gl.GetUniformLocation(w.Program, gl.Str("projection\x00")), 1, false, &projection[0])
	gl.Uniform3f(gl.GetUniformLocation(w.Program, gl.Str("lightDir\x00")), 0.5, -1.0, 0.3)
	gl.Uniform3f(gl.GetUniformLocation(w.Program, gl.Str("viewPos\x00")), viewPos.X(), viewPos.Y(), viewPos.Z())
	for pos, chunk := range w.Chunks {
		model := mgl32.Translate3D(float32(pos[0]*ChunkSize), 0, float32(pos[1]*ChunkSize))
		gl.UniformMatrix4fv(gl.GetUniformLocation(w.Program, gl.Str("model\x00")), 1, false, &model[0])
		gl.BindVertexArray(chunk.VAO)
		gl.DrawArrays(gl.TRIANGLES, 0, chunk.VertexCount)
	}
	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// GetSurfaceHeight returns the y-coordinate of the topmost solid block at (x, z).
func (w *World) GetSurfaceHeight(x, z float32) float32 {
	chunkX := int(math.Floor(float64(x / float32(ChunkSize))))
	chunkZ := int(math.Floor(float64(z / float32(ChunkSize))))
	key := [2]int{chunkX, chunkZ}
	chunk, exists := w.Chunks[key]
	if !exists {
		return 0 // Default if chunk not loaded
	}
	localX := int(x) % ChunkSize
	localZ := int(z) % ChunkSize
	if localX < 0 {
		localX += ChunkSize
	}
	if localZ < 0 {
		localZ += ChunkSize
	}
	for y := ChunkSize - 1; y >= 0; y-- {
		blockID := chunk.Blocks[localX][y][localZ]
		if block.Blocks[blockID].IsSolid() {
			return float32(y + 1) // Top of solid block
		}
	}
	return 0 // No solid block found
}

// Cleanup releases the world's resources.
func (w *World) Cleanup() {
	for _, chunk := range w.Chunks {
		chunk.Cleanup()
	}
	gl.DeleteProgram(w.Program)
	gl.DeleteTextures(1, &w.Texture)
}

// Helper functions (shader and texture loading)
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
		return 0, fmt.Errorf("failed to open texture at %s: %v", path, err)
	}
	defer imgFile.Close()
	img, err := png.Decode(imgFile) // Use png.Decode for PNGs
	if err != nil {
		return 0, fmt.Errorf("failed to decode PNG at %s: %v", path, err)
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
