package api

import (
	"encoding/hex"
	"strconv"
	"time"

	"poa-demo/internal/blockchain"
	"poa-demo/internal/engine"
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

// ChainView es la respuesta de GET /api/v1/chain y de los POST de control.
type ChainView struct {
	Blocks       []BlockView `json:"blocks"`
	Supply       uint32      `json:"supply"`
	Difficulty   int         `json:"difficulty"`
	IntervalMs   int         `json:"intervalMs"`
	Status       string      `json:"status"`
	Generation   uint64      `json:"generation"`
	Height       uint64      `json:"height"`
	MaxBlocks    int         `json:"maxBlocks"`
	BlockCount   int         `json:"blockCount"`
	LimitReached bool        `json:"limitReached"`
	StopReason   string      `json:"stopReason"`
}

// StatusView es el evento SSE status. No repite la cadena.
type StatusView struct {
	Status       string `json:"status"`
	Generation   uint64 `json:"generation"`
	Height       uint64 `json:"height"`
	Supply       uint32 `json:"supply"`
	Difficulty   int    `json:"difficulty"`
	IntervalMs   int    `json:"intervalMs"`
	MaxBlocks    int    `json:"maxBlocks"`
	BlockCount   int    `json:"blockCount"`
	LimitReached bool   `json:"limitReached"`
	StopReason   string `json:"stopReason"`
}

// BlockEvent es el evento SSE block, con la generación que lo confirmó.
type BlockEvent struct {
	BlockView
	Generation uint64 `json:"generation"`
}

// LimitError es el cuerpo 409 de start cuando la generación ya está llena.
type LimitError struct {
	Error      string `json:"error"`
	MaxBlocks  int    `json:"maxBlocks"`
	BlockCount int    `json:"blockCount"`
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
	return newChainView(engine.View{
		Blocks:     blocks,
		Supply:     supply,
		Difficulty: difficulty,
		Status:     engine.StatusPaused,
	})
}

// ChainFromView proyecta el estado del coordinador.
func ChainFromView(view engine.View) ChainView {
	return newChainView(view)
}

// StatusFromView proyecta el evento status.
func StatusFromView(view engine.View) StatusView {
	chain := newChainView(view)
	return StatusView{
		Status:       chain.Status,
		Generation:   chain.Generation,
		Height:       chain.Height,
		Supply:       chain.Supply,
		Difficulty:   chain.Difficulty,
		IntervalMs:   chain.IntervalMs,
		MaxBlocks:    chain.MaxBlocks,
		BlockCount:   chain.BlockCount,
		LimitReached: chain.LimitReached,
		StopReason:   chain.StopReason,
	}
}

func newChainView(view engine.View) ChainView {
	blocks := view.Blocks
	views := make([]BlockView, len(blocks))
	for i, block := range blocks {
		views[i] = NewBlockView(block)
	}
	if views == nil {
		views = []BlockView{}
	}
	interval := int(view.Interval / time.Millisecond)
	return ChainView{
		Blocks:       views,
		Supply:       view.Supply,
		Difficulty:   view.Difficulty,
		IntervalMs:   interval,
		Status:       view.Status,
		Generation:   view.Generation,
		Height:       view.Height,
		MaxBlocks:    view.MaxBlocks,
		BlockCount:   view.BlockCount,
		LimitReached: view.LimitReached,
		StopReason:   view.StopReason,
	}
}
