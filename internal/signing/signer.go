// Package signing firma y verifica mensajes con una clave que no sale del proceso.
// El algoritmo concreto sigue sin elegir. La firma resultante autentica el mensaje;
// no es una dirección cifrada.
package signing

import (
	"errors"

	"poa-demo/internal/blockchain"
)

var (
	// ErrNotImplemented lo devuelve la criptografía pendiente.
	// No acompaña una firma utilizable.
	ErrNotImplemented = errors.New("signing: not implemented")

	ErrNoLocalKey = errors.New("signing: falta la clave privada local")
	ErrNoMaterial = errors.New("signing: falta la clave pública o la firma")
)

// Local guarda la clave del proceso. private no se exporta y no debe
// copiarse a bloques, transacciones, mensajes ni registros.
type Local struct {
	Address blockchain.Address
	Public  []byte
	private []byte
}

// Sign firma message con la clave local.
func (l Local) Sign(message []byte) ([]byte, error) {
	if len(l.private) == 0 {
		return nil, ErrNoLocalKey
	}
	return nil, ErrNotImplemented
}

// Verify comprueba sig sobre message con la clave pública asociada a la dirección.
// La dirección entera no sustituye a public.
func Verify(public, message, sig []byte) error {
	if len(public) == 0 || len(sig) == 0 {
		return ErrNoMaterial
	}
	return ErrNotImplemented
}

// CreateTx arma una transferencia de satoshis propios.
// No modifica owners. claimed solo cambia después de obtener una firma.
func CreateTx(owners map[blockchain.SatoshiID]blockchain.Address, claimed map[blockchain.SatoshiID]struct{}, me, to blockchain.Address, ids []blockchain.SatoshiID, sign func([]byte) ([]byte, error)) (blockchain.Tx, error) {
	if len(ids) == 0 {
		return blockchain.Tx{}, blockchain.ErrEmptyTx
	}
	seen := make(map[blockchain.SatoshiID]struct{}, len(ids))
	for _, id := range ids {
		if _, dup := seen[id]; dup {
			return blockchain.Tx{}, blockchain.ErrDuplicateSatoshi
		}
		seen[id] = struct{}{}
		owner, ok := owners[id]
		if !ok {
			return blockchain.Tx{}, blockchain.ErrUnknownSatoshi
		}
		if owner != me {
			return blockchain.Tx{}, blockchain.ErrNotOwner
		}
		if _, ok := claimed[id]; ok {
			return blockchain.Tx{}, blockchain.ErrAlreadyPending
		}
	}
	tx := blockchain.Tx{Satoshis: append([]blockchain.SatoshiID{}, ids...), Sender: me, Recipient: to}
	raw, err := sign(blockchain.TxMessage(tx))
	if err != nil {
		return blockchain.Tx{}, err
	}
	tx.Auth = blockchain.Signature{Signer: me, Value: raw}
	for _, id := range ids {
		claimed[id] = struct{}{}
	}
	return tx, nil
}
