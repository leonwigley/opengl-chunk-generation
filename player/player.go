package player

import (
	"math"
	"something/block"
	aaa "something/world"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Player struct {
	Camera   *Camera
	Position mgl32.Vec3
	Velocity mgl32.Vec3
	OnGround bool
	Height   float32
	Width    float32
}

func NewPlayer(position mgl32.Vec3) *Player {
	return &Player{
		Camera:   NewCamera(position.Add(mgl32.Vec3{0, 1.5, 0})),
		Position: position,
		Velocity: mgl32.Vec3{0, 0, 0},
		OnGround: false,
		Height:   1.8,
		Width:    0.5,
	}
}

func (p *Player) Update(window *glfw.Window, world *aaa.World, deltaTime float32) {
	speed := float32(10.0)
	if window.GetKey(glfw.KeyW) == glfw.Press {
		p.Velocity = p.Velocity.Add(p.Camera.Front.Mul(speed * deltaTime))
	}
	if window.GetKey(glfw.KeyS) == glfw.Press {
		p.Velocity = p.Velocity.Sub(p.Camera.Front.Mul(speed * deltaTime))
	}
	if window.GetKey(glfw.KeyA) == glfw.Press {
		p.Velocity = p.Velocity.Sub(p.Camera.Right.Mul(speed * deltaTime))
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		p.Velocity = p.Velocity.Add(p.Camera.Right.Mul(speed * deltaTime))
	}
	if window.GetKey(glfw.KeySpace) == glfw.Press && p.OnGround {
		p.Velocity[1] = 8.0
		p.OnGround = false
	}

	gravity := float32(-25.0)
	p.Velocity[1] += gravity * deltaTime
	p.move(world, deltaTime)
	p.Camera.Position = p.Position.Add(mgl32.Vec3{0, p.Height - 0.2, 0})
}

func (p *Player) move(world *aaa.World, deltaTime float32) {
	newPos := p.Position
	steps := []mgl32.Vec3{
		{p.Velocity[0] * deltaTime, 0, 0},
		{0, p.Velocity[1] * deltaTime, 0},
		{0, 0, p.Velocity[2] * deltaTime},
	}
	for _, step := range steps {
		testPos := newPos.Add(step)
		if !p.checkCollision(world, testPos) {
			newPos = testPos
		} else if step[1] < 0 {
			p.Velocity[1] = 0
			p.OnGround = true
		}
	}
	p.Position = newPos
}

func (p *Player) checkCollision(world *aaa.World, pos mgl32.Vec3) bool {
	minX := int(math.Floor(float64(pos.X() - p.Width/2)))
	maxX := int(math.Floor(float64(pos.X() + p.Width/2)))
	minY := int(math.Floor(float64(pos.Y())))
	maxY := int(math.Floor(float64(pos.Y() + p.Height)))
	minZ := int(math.Floor(float64(pos.Z() - p.Width/2)))
	maxZ := int(math.Floor(float64(pos.Z() + p.Width/2)))
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			for z := minZ; z <= maxZ; z++ {
				chunkX, chunkZ := x/int(aaa.ChunkSize), z/int(aaa.ChunkSize)
				localX, localY, localZ := x%int(aaa.ChunkSize), y, z%int(aaa.ChunkSize)
				if localX < 0 {
					localX += int(aaa.ChunkSize)
					chunkX--
				}
				if localZ < 0 {
					localZ += int(aaa.ChunkSize)
					chunkZ--
				}
				if localY < 0 || localY >= int(aaa.ChunkSize) {
					continue
				}
				chunk, exists := world.Chunks[[2]int{chunkX, chunkZ}]
				if exists && chunk.Blocks[localX][localY][localZ] != block.BlockAir {
					return true
				}
			}
		}
	}
	return false
}
