package world

import (
	"github.com/aquilax/go-perlin"
	"github.com/go-gl/gl/v4.6-core/gl"
)

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
					c.Blocks[i][j][k] = BlockStone
				} else if j < height {
					c.Blocks[i][j][k] = BlockDirt
				} else if j == height {
					c.Blocks[i][j][k] = BlockGrass
				} else {
					c.Blocks[i][j][k] = BlockAir
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
				if x == ChunkSize-1 || c.Blocks[x+1][y][z] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "right", block)...)
				}
				if x == 0 || c.Blocks[x-1][y][z] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "left", block)...)
				}
				if y == ChunkSize-1 || c.Blocks[x][y+1][z] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "top", block)...)
				}
				if y == 0 || c.Blocks[x][y-1][z] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "bottom", block)...)
				}
				if z == ChunkSize-1 || c.Blocks[x][y][z+1] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "front", block)...)
				}
				if z == 0 || c.Blocks[x][y][z-1] == BlockAir {
					mesh = append(mesh, createFace(float32(x), float32(y), float32(z), "back", block)...)
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
	c.VertexCount = int32(len(mesh) / 8)
}

func UploadMesh(vertices []float32) (vao, vbo uint32) {
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(1)
	return
}

func createFace(x, y, z float32, face string, block BlockID) []float32 {
	var nx, ny, nz float32
	var u0, v0, u1, v1 float32
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

	switch block {
	case BlockGrass:
		if face == "top" {
			u0, v0, u1, v1 = 0.0/4, 0.0/4, 1.0/4, 1.0/4
		} else {
			u0, v0, u1, v1 = 0.0/4, 1.0/4, 1.0/4, 2.0/4
		}
	case BlockDirt:
		u0, v0, u1, v1 = 1.0/4, 0.0/4, 2.0/4, 1.0/4
	case BlockStone:
		u0, v0, u1, v1 = 2.0/4, 0.0/4, 3.0/4, 1.0/4
	default:
		u0, v0, u1, v1 = 3.0/4, 0.0/4, 4.0/4, 1.0/4
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
						mesh = append(mesh, createFace(float32(x), float32(y), float32(z), face, BlockAir)...)
					}
				}
			}
		}
	}
	return mesh
}
