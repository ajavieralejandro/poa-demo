package consensus

import (
	"testing"

	"poa-demo/internal/blockchain"
)

func TestSelectRejectsEmptySupply(t *testing.T) {
	_, err := Select(blockchain.Header{}, 0)
	if err != ErrEmptySupply {
		t.Fatalf("supply vacío: %v", err)
	}
}

func TestMineAndSelectStayPending(t *testing.T) {
	if _, err := Mine(blockchain.Block{}, 2, 8); err != ErrNotImplemented {
		t.Fatalf("Mine: %v", err)
	}
	ids, err := Select(blockchain.Header{}, 4)
	if err != ErrNotImplemented {
		t.Fatalf("Select: %v", err)
	}
	if ids != ([3]blockchain.SatoshiID{}) {
		t.Fatalf("Select devolvió índices: %v", ids)
	}
}
