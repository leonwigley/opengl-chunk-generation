package block

// BlockID represents a block type identifier.
type BlockID byte

const (
	BlockAir BlockID = iota
	BlockGrass
	BlockDirt
	BlockStone
)

// Block defines the interface for all block types.
type Block interface {
	ID() BlockID
	GetUVs(face string) (u0, v0, u1, v1 float32) // Texture coordinates for a face
	IsSolid() bool                               // For collision and rendering
}

// GrassBlock represents a grass block.
type GrassBlock struct{ id BlockID }

func (g GrassBlock) ID() BlockID   { return g.id }
func (g GrassBlock) IsSolid() bool { return true }
func (g GrassBlock) GetUVs(face string) (u0, v0, u1, v1 float32) {
	if face == "top" {
		return 0.0 / 4, 0.0 / 4, 1.0 / 4, 1.0 / 4 // Grass top
	}
	return 0.0 / 4, 1.0 / 4, 1.0 / 4, 2.0 / 4 // Grass side
}

// DirtBlock represents a dirt block.
type DirtBlock struct{ id BlockID }

func (d DirtBlock) ID() BlockID   { return d.id }
func (d DirtBlock) IsSolid() bool { return true }
func (d DirtBlock) GetUVs(face string) (u0, v0, u1, v1 float32) {
	return 1.0 / 4, 0.0 / 4, 2.0 / 4, 1.0 / 4 // Dirt
}

// StoneBlock represents a stone block.
type StoneBlock struct{ id BlockID }

func (s StoneBlock) ID() BlockID   { return s.id }
func (s StoneBlock) IsSolid() bool { return true }
func (s StoneBlock) GetUVs(face string) (u0, v0, u1, v1 float32) {
	return 2.0 / 4, 0.0 / 4, 3.0 / 4, 1.0 / 4 // Stone
}

// AirBlock represents an air block (non-solid).
type AirBlock struct{ id BlockID }

func (a AirBlock) ID() BlockID   { return a.id }
func (a AirBlock) IsSolid() bool { return false }
func (a AirBlock) GetUVs(face string) (u0, v0, u1, v1 float32) {
	return 0, 0, 0, 0 // No rendering
}

// Registry maps BlockIDs to Block instances.
var Blocks = map[BlockID]Block{
	BlockAir:   AirBlock{id: BlockAir},
	BlockGrass: GrassBlock{id: BlockGrass},
	BlockDirt:  DirtBlock{id: BlockDirt},
	BlockStone: StoneBlock{id: BlockStone},
}
