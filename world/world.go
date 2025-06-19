package world

import (
	"math"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type World struct {
	Chunks      map[[2]int]*Chunk
	ChunkRadius int
}

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
			gl.DeleteVertexArrays(1, &w.Chunks[key].VAO)
			gl.DeleteBuffers(1, &w.Chunks[key].VBO)
			delete(w.Chunks, key)
		}
	}
}
