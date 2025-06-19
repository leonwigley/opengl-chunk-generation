package block

import "something/world"

type Block struct {
	ID    world.BlockID
	Name  string
	Solid bool
}

var Blocks = map[world.BlockID]Block{
	world.BlockAir:   {ID: world.BlockAir, Name: "Air", Solid: false},
	world.BlockGrass: {ID: world.BlockGrass, Name: "Grass", Solid: true},
	world.BlockDirt:  {ID: world.BlockDirt, Name: "Dirt", Solid: true},
	world.BlockStone: {ID: world.BlockStone, Name: "Stone", Solid: true},
}
