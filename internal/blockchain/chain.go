package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

var (
	// ErrNotImplemented lo devuelve la alta mientras minería, selección
	// y firmas sigan pendientes. Append no confirma un bloque en ese caso.
	ErrNotImplemented = errors.New("blockchain: not implemented")

	ErrNotLinearTip     = errors.New("blockchain: el bloque no extiende el tip")
	ErrEmptyTx          = errors.New("blockchain: la transacción no indica satoshis")
	ErrUnknownSatoshi   = errors.New("blockchain: el satoshi no existe")
	ErrNotOwner         = errors.New("blockchain: el satoshi no pertenece al emisor")
	ErrDuplicateSatoshi = errors.New("blockchain: satoshi repetido")
	ErrAlreadyPending   = errors.New("blockchain: el satoshi ya está en una transacción pendiente")
)

// HeaderBytes es la codificación canónica: Prev, Miner, Height, Nonce.
// Enteros en big-endian.
func HeaderBytes(h Header) []byte {
	buf := make([]byte, 0, 32+4+8+8)
	buf = append(buf, h.Prev[:]...)
	buf = appendU32(buf, uint32(h.Miner))
	buf = appendU64(buf, h.Height)
	buf = appendU64(buf, h.Nonce)
	return buf
}

// HeaderMessage es el mensaje de las posiciones 0 y 1.
// La posición entra en el mensaje para que el mismo dueño firme dos cupos distintos.
func HeaderMessage(position uint8, h Header) []byte {
	return append([]byte{position}, HeaderBytes(h)...)
}

// TxMessage cubre los satoshis, el emisor y el receptor.
// No incluye Auth.Value: esos bytes son la firma, no el mensaje.
func TxMessage(tx Tx) []byte {
	buf := make([]byte, 0, 4+len(tx.Satoshis)*4+8)
	buf = appendU32(buf, uint32(len(tx.Satoshis)))
	for _, id := range tx.Satoshis {
		buf = appendU32(buf, uint32(id))
	}
	buf = appendU32(buf, uint32(tx.Sender))
	buf = appendU32(buf, uint32(tx.Recipient))
	return buf
}

// BlockMessage es el mensaje de la posición 2: la cabecera y las transacciones.
func BlockMessage(h Header, txs []Tx) []byte {
	buf := HeaderBytes(h)
	buf = appendU32(buf, uint32(len(txs)))
	for _, tx := range txs {
		buf = append(buf, TxMessage(tx)...)
	}
	return buf
}

// HashHeader identifica un bloque. Comprometer la cabecera no reemplaza
// la verificación de cada firma.
func HashHeader(h Header) Hash {
	return sha256.Sum256(HeaderBytes(h))
}

// NewLinear crea la génesis y el reparto inicial.
// El satoshi i pertenece a owners[i]. Este bloque no exige las tres firmas.
// No recibe claves privadas. No emite unidades nuevas ni fija comisiones.
func NewLinear(owners []Address) Chain {
	ch := Chain{
		Blocks: map[Hash]Block{},
		Owners: make(map[SatoshiID]Address, len(owners)),
	}
	for i, owner := range owners {
		ch.Owners[SatoshiID(i)] = owner
	}
	genesis := Block{Header: Header{Height: 0}}
	ch.Tip = HashHeader(genesis.Header)
	ch.Blocks[ch.Tip] = genesis
	return ch
}

// Append incorpora un bloque solo si extiende el tip.
//
// El camino pendiente, que todavía responde ErrNotImplemented, tiene que:
// comprobar el trabajo, elegir tres satoshis, verificar cada firma con pub
// y verify, aplicar las transacciones sobre una copia de Owners y recién
// entonces reemplazar Owners, Blocks y Tip juntos.
// pub son claves públicas. verify no recibe claves privadas.
//
// Un padre distinto del tip se rechaza. Esta versión no guarda ramas,
// no reorganiza y no habla por la red.
func (c *Chain) Append(b Block, pub map[Address][]byte, verify func(public, message, sig []byte) error) error {
	if b.Header.Prev != c.Tip {
		return ErrNotLinearTip
	}
	return ErrNotImplemented
}

func appendU32(dst []byte, v uint32) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], v)
	return append(dst, buf[:]...)
}

func NewGenesisBlock() Block {
	return Block{
		Header: Header{
			Height: 0,
			Nonce:  0,
			Prev:   [32]byte{}, // Un arreglo de 32 ceros
		},
		Signatures: [3]Signature{}, // Arreglo vacío de firmas para el génesis
		Txs:        []Tx{},         // Sin transacciones iniciales
	}
}

func appendU64(dst []byte, v uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], v)
	return append(dst, buf[:]...)
}
