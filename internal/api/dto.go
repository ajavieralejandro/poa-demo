package api

import (
	"encoding/hex"
	"strconv"

	"poa-demo/internal/blockchain"
)

// BlockView es el contrato JSON que consume la cuadrícula de React.
type BlockView struct {
	Header     HeaderView      `json:"header"`
	Signatures []SignatureView `json:"signatures"`
	Hash       string          `json:"hash"`
	Verified   bool            `json:"verified"`
}

// HeaderView omite el suministro: el cliente lo lee en la foto de la cadena.
type HeaderView struct {
	Prev   string `json:"prev"`
	Miner  string `json:"miner"`
	Height uint64 `json:"height"`
	Nonce  uint64 `json:"nonce"`
}

// SignatureView identifica al validador sorteado y el satoshi que le tocó.
type SignatureView struct {
	Signer    string `json:"signer"`
	SatoshiID uint32 `json:"satoshiId"`
}

// ChainView es la respuesta de GET /api/v1/chain.
type ChainView struct {
	Blocks     []BlockView `json:"blocks"`
	Supply     uint32      `json:"supply"`
	Difficulty int         `json:"difficulty"`
}

// NewBlockView proyecta un bloque confirmado al contrato del frontend.
func NewBlockView(block blockchain.Block) BlockView {
	hash := block.Header.Hash()
	signatures := make([]SignatureView, 0, len(block.Signatures))
	for _, signature := range block.Signatures {
		if !signature.Valid() {
			continue
		}
		signatures = append(signatures, SignatureView{
			Signer:    strconv.FormatUint(uint64(signature.Signer), 10),
			SatoshiID: uint32(signature.Value),
		})
	}

	return BlockView{
		Header: HeaderView{
			Prev:   hex.EncodeToString(block.Header.Prev[:]),
			Miner:  strconv.FormatUint(uint64(block.Header.Miner), 10),
			Height: block.Header.Height,
			Nonce:  block.Header.Nonce,
		},
		Signatures: signatures,
		Hash:       hex.EncodeToString(hash[:]),
		Verified:   true,
	}
}

// NewChainView arma la foto completa que el cliente carga antes de abrir el stream.
func NewChainView(blocks []blockchain.Block, supply uint32, difficulty int) ChainView {
	views := make([]BlockView, len(blocks))
	for i, block := range blocks {
		views[i] = NewBlockView(block)
	}
	return ChainView{
		Blocks:     views,
		Supply:     supply,
		Difficulty: difficulty,
	}
}
