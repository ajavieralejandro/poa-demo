// Package signing define la firma de mensajes sin elegir algoritmo.
//
// Una implementación posterior puede usar Ed25519, ECDSA u otro esquema.
// Esta interfaz no impone RSA para la firma definitiva.
//
// Un hash de bloque puede comprometer el contenido junto con el conjunto
// de firmas. Ese hash no sustituye la verificación de cada firma.
// hash(UUID || contenido), si se usa más adelante, es solo una simulación
// y no proporciona autenticidad criptográfica.
//
// Los identificadores de firmante son UUID públicos. Este paquete no los
// trata como claves privadas y no genera claves, UUID, hashes ni firmas.
package signing

import (
	"errors"

	"poa-demo/internal/blockchain"
)

// ErrNotImplemented lo devuelve la firma y la verificación hasta que
// exista un algoritmo.
var ErrNotImplemented = errors.New("signing: not implemented")

// Signer firma y verifica un mensaje opaco.
// El algoritmo concreto lo elige la implementación, no esta interfaz.
type Signer interface {
	Sign(signerID string, message []byte) (blockchain.Signature, error)
	Verify(message []byte, signature blockchain.Signature) error
}

// UnimplementedSigner no produce ni acepta una firma.
// La Signature cero que acompaña al error no es una firma.
type UnimplementedSigner struct{}

var _ Signer = UnimplementedSigner{}

func (UnimplementedSigner) Sign(string, []byte) (blockchain.Signature, error) {
	return blockchain.Signature{}, ErrNotImplemented
}

func (UnimplementedSigner) Verify([]byte, blockchain.Signature) error {
	return ErrNotImplemented
}
