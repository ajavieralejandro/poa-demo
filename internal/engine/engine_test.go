package engine

import (
	"testing"

	"poa-demo/internal/config"
)

func TestMineNextAppendsVerifiedBlock(t *testing.T) {
	eng, err := New(config.Config{Difficulty: 0, InitialSupply: 1000})
	if err != nil {
		t.Fatal(err)
	}

	var height uint64
	for attempt := 0; attempt < 20; attempt++ {
		block, mineErr := eng.MineNext()
		if mineErr != nil {
			err = mineErr
			continue
		}
		height = block.Header.Height
		err = nil
		break
	}
	if err != nil {
		t.Fatal(err)
	}
	if height != 1 {
		t.Fatalf("altura = %d", height)
	}

	blocks, supply, difficulty := eng.Snapshot()
	if len(blocks) != 2 || blocks[0].Header.Height != 0 || blocks[1].Header.Height != 1 {
		t.Fatalf("cadena = %d bloques", len(blocks))
	}
	if supply != 1000 || difficulty != 0 {
		t.Fatalf("supply %d dificultad %d", supply, difficulty)
	}
}

func TestNewRejectsSupplyBelowPool(t *testing.T) {
	_, err := New(config.Config{InitialSupply: 10})
	if err == nil {
		t.Fatal("se esperaba un error de suministro")
	}
}
