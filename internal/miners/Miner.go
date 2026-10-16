package miners
import(
	"errors"
)

var siguienteSatoshiID uint32 = 0
// NewMiner inicializa y retorna un nuevo nodo minero con su dirección y sus satoshis asignados.
func NewMiner(addr Address, satoshis []SatoshiID) Miner {
	// TODO: Inicializar y retornar la estructura Miner
	return Miner{}
}

// HasSatoshi verifica si este minero es el propietario de un SatoshiID específico 
// que resultó ganador en el sorteo Follow the Satoshi.
func (m *Miner) HasSatoshi(satID SatoshiID) bool {
	// TODO: Recorrer m.Satoshis para comprobar si el ID pertenece a este nodo
	return false
}

// SignBlock genera una firma digital válida utilizando la identidad y clave del minero 
// para autorizar un bloque propuesto.
func (m *Miner) SignBlock(blockHeader blockchain.Header) (blockchain.Signature, error) {
	// TODO: Firmar el header del bloque y retornar la estructura de firma
	return blockchain.Signature{}, nil
}

// ProposeBlock coordina el turno de minería del nodo: empaqueta transacciones, 
// ejecuta la minería (PoW) y firma el bloque resultante si tiene el derecho de hacerlo.
func (m *Miner) ProposeBlock(parent blockchain.Block, txs []blockchain.Tx, bits int) (blockchain.Block, error) {
	// TODO: Llamar al motor de consenso (Mine) y adjuntar la firma del minero
	return blockchain.Block{}, nil
}