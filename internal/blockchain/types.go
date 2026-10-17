// Package blockchain define la primera versión local:
package blockchain

// un proceso, tres posiciones, satoshis individuales y una cadena lineal.
// Las claves privadas no pertenecen a este paquete.
import "fmt"

// SatoshiID identifica una unidad. En la génesis los identificadores son 0..n-1.
type SatoshiID uint32

// Address es un identificador entero. No prueba identidad por sí sola:
// la firma se verifica con la clave pública que el proceso asocia a esa dirección.
type Address uint32

// Hash es el SHA-256 de una cabecera canónica.
type Hash [32]byte

// Header es la cabecera de trabajo. No incluye transacciones.
type Header struct {
	Prev   Hash
	Miner  Address
	Height uint64
	Nonce  uint64
}

// Signature autentica un mensaje.
// Value son los bytes de la firma, no una dirección cifrada.
// Position es el cupo 0, 1 o 2. En la autorización de una transacción se ignora.
type Signature struct {
	Signer   Address
	Position uint8
	Value    []byte
}

// Tx transfiere satoshis concretos. La lista sola no autoriza el movimiento:
// hacen falta emisor, receptor y Auth.
type Tx struct {
	Satoshis  []SatoshiID
	Sender    Address
	Recipient Address
	Auth      Signature
}

// Block reúne la cabecera, las tres firmas y las transacciones.
// La clave del mapa es HashHeader, no un campo guardado aparte.
type Block struct {
	Header     Header
	Signatures [3]Signature
	Txs        []Tx
}

// Chain es lineal. Owners es la única propiedad confirmada.
// Tip es el único extremo aceptado.
type Chain struct {
	Blocks map[Hash]Block
	Tip    Hash
	Owners map[SatoshiID]Address
}
// Summary retorna una representación formateada del bloque.
// Summary retorna una representación formateada y segura del bloque.
func (b Block) Summary() string {
    n := b.Header.Height
    nonce := b.Header.Nonce
    
    // Obtienes el hash de la cabecera usando la función correspondiente de tu paquete blockchain
    // (Asegúrate de usar el nombre de la función que calcula el hash de la cabecera en tu proyecto, ej: HashHeader(b.Header))
    hashArray := HashHeader(b.Header) 
    
    // Tomamos los primeros 4 bytes y los formateamos como texto hexadecimal
    hashCorto := fmt.Sprintf("%x", hashArray[:4])

    return fmt.Sprintf("Height: %d | Nonce: %d | Hash: %s...", n, nonce, hashCorto)
}