// Package blockchain define el modelo de una cadena lineal académica.
// Los identificadores son UUID públicos. No son claves privadas.
package blockchain

// Signature es una firma desacoplada, independiente del algoritmo.
// SignerID es el identificador público del firmante (UUID), no una clave privada.
// Value guardará los bytes de la firma cuando se elija un algoritmo.
// hash(UUID || contenido) no debe usarse como firma digital.
type Signature struct {
	SignerID string
	Value    []byte
}

// Transaction es una transferencia entre dos identificadores públicos.
type Transaction struct {
	ID          string
	SenderID    string
	RecipientID string
	Amount      uint64
	Signature   Signature
}

// Validator es un participante que puede ser elegido para firmar un bloque.
// Quiénes entran en la selección y qué umbral basta siguen sin definir.
type Validator struct {
	ID string
}

// Block es un eslabón de una cadena lineal.
// ChildBlock guarda el hash del único sucesor. No representa bifurcaciones
// ni ramas en competencia.
type Block struct {
	Transactions    []Transaction
	PreviousHash    string
	BlockHash       string
	BlockSignatures []Signature
	ChildBlock      string
}

// Chain indexa los bloques por su hash. HeadHash es la cabeza actual.
type Chain struct {
	Blocks   map[string]*Block
	HeadHash string
}
