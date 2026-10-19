// Package engine ejecuta una ronda de minado, firma y verificación
// sobre la cadena lineal que ya usa la demo de consola.
package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"poa-demo/internal/blockchain"
	"poa-demo/internal/config"
	"poa-demo/internal/consensus"
	"poa-demo/internal/miners"
)

// ErrBlockLimit es el rechazo de Start cuando la generación ya tiene MAX_BLOCKS bloques.
var ErrBlockLimit = errors.New("block_limit_reached")

const (
	StatusPaused   = "paused"
	StatusRunning  = "running"
	StopBlockLimit = "block_limit"
)

// validatorNodes es el reparto fijo de la demo: el suministro tiene que alcanzar
// para dar al menos un satoshi a cada nodo y evitar el desbordamiento del pool.
const validatorNodes = 100

// Engine guarda la cadena confirmada y el único coordinador de la demo.
// MineNext sigue disponible para la prueba de estrés. El servidor usa Run.
type Engine struct {
	mu         sync.RWMutex
	controlMu  sync.Mutex
	cond       *sync.Cond
	chain      blockchain.Chain
	nodes      map[blockchain.Address]*miners.Node
	addresses  []blockchain.Address
	cursor     int
	difficulty int
	supply     uint32
	maxBlocks  int
	interval   time.Duration

	generation  uint64
	roundID     uint64
	running     bool
	stopReason  string
	busy        bool
	workerOn    bool
	roundCancel context.CancelFunc
}

// View es la foto que comparten HTTP y el stream.
type View struct {
	Blocks       []blockchain.Block
	Supply       uint32
	Difficulty   int
	Interval     time.Duration
	Status       string
	Generation   uint64
	Height       uint64
	MaxBlocks    int
	BlockCount   int
	LimitReached bool
	StopReason   string
}

// Notice avisa un bloque confirmado o un cambio de estado.
type Notice struct {
	Kind       string
	Block      blockchain.Block
	Generation uint64
	State      View
}

const (
	NoticeBlock  = "block"
	NoticeStatus = "status"
)

// New arma la génesis y el pool determinista a partir de la configuración.
func New(cfg config.Config) (*Engine, error) {
	if cfg.InitialSupply < validatorNodes {
		return nil, fmt.Errorf("engine: INITIAL_SUPPLY %d es menor que los %d nodos del pool", cfg.InitialSupply, validatorNodes)
	}

	genesis := blockchain.NewGenesisBlock(cfg.InitialSupply)
	chain := blockchain.Chain{
		Blocks: map[blockchain.Hash]blockchain.Block{},
		Tip:    genesis.Header.Hash(),
		Owners: map[blockchain.SatoshiID]blockchain.Address{},
	}
	chain.Blocks[chain.Tip] = genesis

	addresses := make([]blockchain.Address, validatorNodes)
	for i := range addresses {
		addresses[i] = blockchain.Address(i + 1)
	}
	chain, nodes := miners.NewDeterministicPool(chain, addresses)

	maxBlocks := cfg.MaxBlocks
	if maxBlocks < 3 {
		maxBlocks = 1000
	}

	eng := &Engine{
		chain:      chain,
		nodes:      nodes,
		addresses:  addresses,
		difficulty: cfg.Difficulty,
		supply:     cfg.InitialSupply,
		maxBlocks:  maxBlocks,
		interval:   cfg.BlockInterval,
		generation: 1,
		running:    true,
	}
	eng.cond = sync.NewCond(&eng.mu)
	return eng, nil
}

// Snapshot devuelve los bloques desde la génesis hasta el tip, más los parámetros públicos.
func (e *Engine) Snapshot() (blocks []blockchain.Block, supply uint32, difficulty int) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return orderedBlocks(e.chain), e.supply, e.difficulty
}

// MineNext mina, firma y verifica el hijo del tip. Solo entonces avanza la cadena.
func (e *Engine) MineNext() (blockchain.Block, error) {
	e.mu.Lock()
	addr := e.addresses[e.cursor%len(e.addresses)]
	e.cursor++
	node := e.nodes[addr]
	chain := cloneChain(e.chain)
	nodes := e.nodes
	bits := e.difficulty
	e.mu.Unlock()

	next, err := mineAndSeal(node, chain, nodes, bits)
	if err != nil {
		return blockchain.Block{}, err
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if next.Header.Prev != e.chain.Tip {
		return blockchain.Block{}, fmt.Errorf("engine: el bloque no extiende el tip actual")
	}
	hash := next.Header.Hash()
	e.chain.Blocks[hash] = next
	e.chain.Tip = hash
	return next, nil
}

func mineAndSeal(node *miners.Node, chain blockchain.Chain, nodes map[blockchain.Address]*miners.Node, bits int) (blockchain.Block, error) {
	return mineAndSealCtx(context.Background(), node, chain, nodes, bits)
}

func mineAndSealCtx(ctx context.Context, node *miners.Node, chain blockchain.Chain, nodes map[blockchain.Address]*miners.Node, bits int) (blockchain.Block, error) {
	if err := ctx.Err(); err != nil {
		return blockchain.Block{}, err
	}
	nextBlock, err := node.MineWithContext(ctx, chain, bits)
	if err != nil {
		return blockchain.Block{}, fmt.Errorf("minar: %w", err)
	}

	lucky, err := consensus.Select(nextBlock.Header, chain)
	if err != nil {
		return blockchain.Block{}, fmt.Errorf("seleccionar validadores: %w", err)
	}

	for i, satID := range lucky {
		owner := miners.FindOwner(nodes, chain, satID)
		if owner == nil {
			return blockchain.Block{}, fmt.Errorf("el satoshi %d no tiene dueño", satID)
		}
		if i == 2 {
			nextBlock, err = owner.ProposeBlock(nextBlock, chain, nil, bits)
		} else {
			nextBlock, err = owner.SignBlock(nextBlock, chain)
		}
		if err != nil {
			return blockchain.Block{}, fmt.Errorf("firmar bloque: %w", err)
		}
	}

	if err = consensus.Verify(nextBlock, chain, bits); err != nil {
		return blockchain.Block{}, fmt.Errorf("verificar bloque: %w", err)
	}
	return nextBlock, nil
}

func cloneChain(chain blockchain.Chain) blockchain.Chain {
	blocks := make(map[blockchain.Hash]blockchain.Block, len(chain.Blocks))
	for hash, block := range chain.Blocks {
		blocks[hash] = block
	}
	owners := make(map[blockchain.SatoshiID]blockchain.Address, len(chain.Owners))
	for id, owner := range chain.Owners {
		owners[id] = owner
	}
	return blockchain.Chain{
		Blocks: blocks,
		Tip:    chain.Tip,
		Owners: owners,
	}
}

func orderedBlocks(chain blockchain.Chain) []blockchain.Block {
	reversed := make([]blockchain.Block, 0, len(chain.Blocks))
	seen := make(map[blockchain.Hash]struct{}, len(chain.Blocks))
	hash := chain.Tip
	for {
		if _, duplicated := seen[hash]; duplicated {
			break
		}
		block, ok := chain.Blocks[hash]
		if !ok {
			break
		}
		seen[hash] = struct{}{}
		reversed = append(reversed, block)
		if block.Header.Height == 0 {
			break
		}
		hash = block.Header.Prev
	}

	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return reversed
}
