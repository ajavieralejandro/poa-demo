package blockchain

import (
	"bytes"
	"testing"
)

func TestNewLinearAssignsSatoshis(t *testing.T) {
	ch := NewLinear([]Address{1, 1, 2, 3})
	if ch.Owners[0] != 1 || ch.Owners[1] != 1 || ch.Owners[2] != 2 || ch.Owners[3] != 3 {
		t.Fatalf("reparto = %v", ch.Owners)
	}
	genesis, ok := ch.Blocks[ch.Tip]
	if !ok {
		t.Fatal("el tip no está en el mapa")
	}
	if genesis.Header.Height != 0 {
		t.Fatalf("altura = %d", genesis.Header.Height)
	}
	if got := HashHeader(genesis.Header); got != ch.Tip {
		t.Fatal("el tip no es el hash de la cabecera génesis")
	}
}

func TestMessagesSeparatePositionAndSignature(t *testing.T) {
	h := Header{Miner: 2, Height: 1}
	if bytes.Equal(HeaderMessage(0, h), HeaderMessage(1, h)) {
		t.Fatal("la posición tiene que cambiar el mensaje de cabecera")
	}
	withSig := Tx{Satoshis: []SatoshiID{1}, Sender: 1, Recipient: 3, Auth: Signature{Value: []byte{9}}}
	plain := Tx{Satoshis: []SatoshiID{1}, Sender: 1, Recipient: 3}
	if !bytes.Equal(TxMessage(withSig), TxMessage(plain)) {
		t.Fatal("la autorización no entra en el mensaje de la transacción")
	}
}

func TestAppendDoesNotCommit(t *testing.T) {
	ch := NewLinear([]Address{1, 1, 2, 3})
	tip := ch.Tip
	owners := ch.Owners[1]

	wrong := Block{Header: Header{Prev: Hash{1}, Height: 1}}
	if err := ch.Append(wrong, nil, nil); err != ErrNotLinearTip {
		t.Fatalf("padre ajeno: %v", err)
	}

	next := Block{Header: Header{Prev: tip, Height: 1}}
	if err := ch.Append(next, nil, nil); err != ErrNotImplemented {
		t.Fatalf("alta pendiente: %v", err)
	}
	if ch.Tip != tip || len(ch.Blocks) != 1 || ch.Owners[1] != owners {
		t.Fatal("Append modificó la cadena")
	}
}
