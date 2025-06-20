package world

import (
	"something/block"

	"github.com/aquilax/go-perlin"
	"github.com/go-gl/gl/v4.6-core/gl"
)

type Chunk struct {
	Blocks      [ChunkSize][ChunkSize][ChunkSize]block.BlockID
	VAO         uint32
	VBO         uint32
	VertexCount int32
}

func NewChunk(x, z int32) *Chunk {
	var c Chunk
	p := perlin.NewPerlin(2, 2, 3, 42)
	for i := 0; i < ChunkSize; i++ {
		for k := 0; k < ChunkSize; k++ {
			worldX := float64(x*ChunkSize+int32(i)) / 50.0
			worldZ := float64(z*ChunkSize+int32(k)) / 50.0
			height := int(p.Noise2D(worldX, worldZ)*10 + 8)
			if height < 0 {
				height = 0
			} else if height > ChunkSize-1 {
				height = ChunkSize - 1
			}
			for j := 0; j < ChunkSize; j++ {
				if j < height-2 {
					c.Blocks[i][j][k] = block.BlockStone
				} else if j < height {
					c.Blocks[i][j][k] = block.BlockDirt
				} else if j == height {
					c.Blocks[i][j][k] = block.BlockGrass
				} else {
					c.Blocks[i][j][k] = block.BlockAir
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
				blockID := c.Blocks[x][y][z]
				b := block.Blocks[blockID]
				if !b.IsSolid() {
					continue
				}
				faces := []string{"right", "left", "top", "bottom", "front", "back"}
				offsets := [][3]int{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}}
				for i, face := range faces {
					nx, ny, nz := x+offsets[i][0], y+offsets[i][1], z+offsets[i][2]
					if nx < 0 || nx >= ChunkSize || ny < 0 || ny >= ChunkSize || nz < 0 || nz >= ChunkSize {
						mesh = append(mesh, c.createFace(float32(x), float32(y), float32(z), face, blockID)...)
						continue
					}
					if !block.Blocks[c.Blocks[nx][ny][nz]].IsSolid() {
						mesh = append(mesh, c.createFace(float32(x), float32(y), float32(z), face, blockID)...)
					}
				}
			}
		}
	}
	return mesh
}

func (c *Chunk) UploadMesh() {
	mesh := c.GenerateMesh()
	c.VAO, c.VBO = uploadMesh(mesh)
	c.VertexCount = int32(len(mesh) / 8)
}

func (c *Chunk) Cleanup() {
	gl.DeleteVertexArrays(1, &c.VAO)
	gl.DeleteBuffers(1, &c.VBO)
}

func (c *Chunk) createFace(x, y, z float32, face string, blockID block.BlockID) []float32 {
	b := block.Blocks[blockID]
	u0, v0, u1, v1 := b.GetUVs(face)
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
			x + 1, y, z, nx, ny, nz, u1, v0,
			x + 1, y + 1, z, nx, ny, nz, u1, v1,
			x + 1, y + 1, z + 1, nx, ny, nz, u0, v1,
			x + 1, y, z, nx, ny, nz, u1, v0,
			x + 1, y + 1, z + 1, nx, ny, nz, u0, v1,
			x + 1, y, z + 1, nx, ny, nz, u0, v0,
		}
	case "left":
		return []float32{
			x, y, z, nx, ny, nz, u0, v0,
			x, y + 1, z + 1, nx, ny, nz, u1, v1,
			x, y + 1, z, nx, ny, nz, u0, v1,
			x, y, z, nx, ny, nz, u0, v0,
			x, y, z + 1, nx, ny, nz, u1, v0,
			x, y + 1, z + 1, nx, ny, nz, u1, v1,
		}
	case "top":
		return []float32{
			x, y + 1, z, nx, ny, nz, u0, v0,
			x + 1, y + 1, z, nx, ny, nz, u1, v0,
			x + 1, y + 1, z + 1, nx, ny, nz, u1, v1,
			x, y + 1, z, nx, ny, nz, u0, v0,
			x + 1, y + 1, z + 1, nx, ny, nz, u1, v1,
			x, y + 1, z + 1, nx, ny, nz, u0, v1,
		}
	case "bottom":
		return []float32{
			x, y, z, nx, ny, nz, u0, v0,
			x + 1, y, z + 1, nx, ny, nz, u1, v1,
			x + 1, y, z, nx, ny, nz, u1, v0,
			x, y, z, nx, ny, nz, u0, v0,
			x, y, z + 1, nx, ny, nz, u0, v1,
			x + 1, y, z + 1, nx, ny, nz, u1, v1,
		}
	case "front":
		return []float32{
			x, y, z + 1, nx, ny, nz, u0, v0,
			x + 1, y + 1, z + 1, nx, ny, nz, u1, v1,
			x + 1, y, z + 1, nx, ny, nz, u1, v0,
			x, y, z + 1, nx, ny, nz, u0, v0,
			x, y + 1, z + 1, nx, ny, nz, u0, v1,
			x + 1, y + 1, z + 1, nx, ny, nz, u1, v1,
		}
	case "back":
		return []float32{
			x, y, z, nx, ny, nz, u0, v0,
			x + 1, y, z, nx, ny, nz, u1, v0,
			x + 1, y + 1, z, nx, ny, nz, u1, v1,
			x, y, z, nx, ny, nz, u0, v0,
			x + 1, y + 1, z, nx, ny, nz, u1, v1,
			x, y + 1, z, nx, ny, nz, u0, v1,
		}
	}
	return nil
}

func uploadMesh(vertices []float32) (vao, vbo uint32) {
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.BindVertexArray(0)
	return
}
