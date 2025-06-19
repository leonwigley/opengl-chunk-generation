package inventory

import "something/block"

type Inventory struct {
	Slots []block.Block
}

func NewInventory() *Inventory {
	return &Inventory{Slots: make([]block.Block, 36)}
}
