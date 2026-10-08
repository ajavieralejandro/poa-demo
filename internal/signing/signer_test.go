package signing

import (
	"bytes"
	"testing"

	"poa-demo/internal/blockchain"
)

func TestSignDoesNotInventASignature(t *testing.T) {
	if _, err := (Local{}).Sign([]byte("msg")); err != ErrNoLocalKey {
		t.Fatalf("sin clave: %v", err)
	}
	sig, err := (Local{private: []byte{1}}).Sign([]byte("msg"))
	if err != ErrNotImplemented || len(sig) != 0 {
		t.Fatalf("firma pendiente = %v, %v", sig, err)
	}
	if err := Verify(nil, []byte("msg"), []byte{1}); err != ErrNoMaterial {
		t.Fatalf("sin clave pública: %v", err)
	}
	if err := Verify([]byte{1}, []byte("msg"), []byte{2}); err != ErrNotImplemented {
		t.Fatalf("verificación pendiente: %v", err)
	}
}

func TestCreateTxDoesNotConfirmOrClaimOnFailure(t *testing.T) {
	owners := map[blockchain.SatoshiID]blockchain.Address{0: 1, 1: 1, 2: 2}
	claimed := map[blockchain.SatoshiID]struct{}{}
	sign := func([]byte) ([]byte, error) { return nil, ErrNotImplemented }

	_, err := CreateTx(owners, claimed, 1, 3, nil, sign)
	if err != blockchain.ErrEmptyTx {
		t.Fatalf("vacía: %v", err)
	}
	_, err = CreateTx(owners, claimed, 1, 3, []blockchain.SatoshiID{2}, sign)
	if err != blockchain.ErrNotOwner {
		t.Fatalf("ajeno: %v", err)
	}
	_, err = CreateTx(owners, claimed, 1, 3, []blockchain.SatoshiID{1, 1}, sign)
	if err != blockchain.ErrDuplicateSatoshi {
		t.Fatalf("duplicado: %v", err)
	}
	_, err = CreateTx(owners, claimed, 1, 3, []blockchain.SatoshiID{1}, sign)
	if err != ErrNotImplemented {
		t.Fatalf("firma pendiente: %v", err)
	}
	if owners[1] != 1 || len(claimed) != 0 {
		t.Fatal("CreateTx cambió propiedad confirmada o pendientes")
	}

	claimed[1] = struct{}{}
	_, err = CreateTx(owners, claimed, 1, 3, []blockchain.SatoshiID{1}, sign)
	if err != blockchain.ErrAlreadyPending {
		t.Fatalf("pendiente: %v", err)
	}
}

func TestCreateTxSignsTheCanonicalMessage(t *testing.T) {
	owners := map[blockchain.SatoshiID]blockchain.Address{1: 1}
	claimed := map[blockchain.SatoshiID]struct{}{}
	var got []byte
	sign := func(message []byte) ([]byte, error) {
		got = append([]byte{}, message...)
		return []byte{7}, nil
	}
	tx, err := CreateTx(owners, claimed, 1, 3, []blockchain.SatoshiID{1}, sign)
	if err != nil {
		t.Fatal(err)
	}
	want := blockchain.Tx{Satoshis: []blockchain.SatoshiID{1}, Sender: 1, Recipient: 3}
	if !bytes.Equal(got, blockchain.TxMessage(want)) {
		t.Fatal("se firmó un mensaje distinto del canónico")
	}
	if _, ok := claimed[1]; !ok || owners[1] != 1 {
		t.Fatal("el pendiente y la propiedad confirmada quedaron mal")
	}
	if tx.Auth.Signer != 1 || !bytes.Equal(tx.Auth.Value, []byte{7}) {
		t.Fatalf("autorización = %+v", tx.Auth)
	}
}
