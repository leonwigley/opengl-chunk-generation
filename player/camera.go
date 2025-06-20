package player

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

type Camera struct {
	Position      mgl32.Vec3
	Front         mgl32.Vec3
	Up            mgl32.Vec3
	Right         mgl32.Vec3
	WorldUp       mgl32.Vec3
	Yaw           float32
	Pitch         float32
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
	c.Pitch -= float32(-yoffset)
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

func thirdPerson() {

}
