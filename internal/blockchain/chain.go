package blockchain

import "errors"

// ErrNotImplemented lo devuelve toda operación de cadena que todavía
// es un contrato vacío. El resto del resultado no es un valor válido.
var ErrNotImplemented = errors.New("blockchain: not implemented")

// Hasher calcula el hash de un bloque.
//
// Un hash final puede comprometer el contenido y el conjunto de firmas.
// Ese compromiso no sustituye la verificación de cada firma.
// Si más adelante un prototipo usa hash(UUID || contenido), ese valor
// es solo una simulación y no aporta autenticidad criptográfica.
type Hasher interface {
	HashBlock(block *Block) (string, error)
}

// Rules valida transacciones y bloques. Los criterios concretos están pendientes.
type Rules interface {
	ValidateTransaction(tx Transaction) error
	ValidateBlock(block *Block) error
}

// Appender incorpora a la cadena un bloque ya finalizado.
type Appender interface {
	AppendBlock(chain *Chain, block *Block) error
}

// Pending reserva el lugar del hash, la validación y la incorporación.
// Ningún método modifica la cadena ni devuelve un bloque o un hash válido.
type Pending struct{}

var (
	_ Hasher   = Pending{}
	_ Rules    = Pending{}
	_ Appender = Pending{}
)

func (Pending) HashBlock(*Block) (string, error) {
	return "", ErrNotImplemented
}

func (Pending) ValidateTransaction(Transaction) error {
	return ErrNotImplemented
}

func (Pending) ValidateBlock(*Block) error {
	return ErrNotImplemented
}

func (Pending) AppendBlock(*Chain, *Block) error {
	return ErrNotImplemented
}
