package miners

import (
	"fmt"
	"poa-demo/internal/blockchain"
	"poa-demo/internal/consensus"
)

var siguienteSatoshiID uint32 = 0

// NewNode inicializa y retorna un nuevo nodo minero con su dirección y sus satoshis asignados.
func NewNode(addr blockchain.Address, satoshis map[blockchain.SatoshiID]struct{}) *Node {
	return &Node{
		addr,
		satoshis,
	}
}

func (m *Node) Mine(chain blockchain.Chain, bits int) (blockchain.Block, error) {
	tip := chain.Blocks[chain.Tip]
	return consensus.Mine(tip, chain, m.Addr, bits)
}

// HasSatoshi verifica si este minero es el propietario de un SatoshiID específico
// que resultó ganador en el sorteo Follow the Satoshi.
func (m *Node) HasSatoshi(satID blockchain.SatoshiID) bool {
	_, ok := m.Satoshis[satID]
	return ok
}

// SignBlock genera una firma digital válida utilizando la identidad y clave del minero
// para autorizar un bloque propuesto.
func (m *Node) SignBlock(block blockchain.Block, chain blockchain.Chain) (blockchain.Block, error) {
	lucky, err := consensus.Select(block.Header, chain)
	if err != nil {
		return block, err
	}
	for _, satId := range lucky {
		if m.HasSatoshi(satId) {
			fmt.Printf("[DEBUG Firma] Bloque firmado por %d\n", m.Addr)
			sign := blockchain.Signature{
				Signer: m.Addr,
				Value:  satId,
			}
			block.Signatures[block.TotalSignatures()] = sign
			return block, nil
		}
	}
	return block, blockchain.ErrNotOwner
}

// ProposeBlock coordina el turno de minería del nodo: empaqueta transacciones,
// Si es el ultimo en firmar el bloque, agrega las transacciones y devuelve el bloque firmado
func (m *Node) ProposeBlock(newBlock blockchain.Block, chain blockchain.Chain, txs []blockchain.Tx, bits int) (blockchain.Block, error) {
	if newBlock.TotalSignatures() != 2 {
		return blockchain.Block{}, blockchain.ErrMissingSignatures
	}
	newBlock, err := m.SignBlock(newBlock, chain)
	if err != nil {
		return blockchain.Block{}, err
	}
	newBlock.Txs = txs
	if err = consensus.Verify(newBlock, chain, bits); err != nil {
		return newBlock, err
	}
	fmt.Printf("[DEBUG Firma] Bloque propuesto por %d\n", m.Addr)
	return newBlock, nil
}
