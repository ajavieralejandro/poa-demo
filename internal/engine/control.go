package engine

import (
	"context"
	"time"

	"poa-demo/internal/blockchain"
	"poa-demo/internal/miners"
)

// Run es el único worker de minado. Start no crea otro.
// Cuando la demo está en pausa o llegó al tope, el worker queda bloqueado en la condición.
func (e *Engine) Run(ctx context.Context, emit func(Notice)) {
	e.mu.Lock()
	if e.workerOn {
		e.mu.Unlock()
		return
	}
	e.workerOn = true
	e.mu.Unlock()

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			e.mu.Lock()
			e.cond.Broadcast()
			e.mu.Unlock()
		case <-done:
		}
	}()

	for {
		e.mu.Lock()
		for !e.running || e.reachedLimitLocked() {
			if ctx.Err() != nil {
				e.mu.Unlock()
				return
			}
			e.cond.Wait()
		}
		if ctx.Err() != nil {
			e.mu.Unlock()
			return
		}

		gen := e.generation
		e.roundID++
		token := e.roundID
		roundCtx, cancel := context.WithCancel(ctx)
		e.roundCancel = cancel
		e.busy = true
		node, chain, nodes, bits := e.prepareLocked()
		e.mu.Unlock()

		next, err := mineAndSealCtx(roundCtx, node, chain, nodes, bits)

		var committed *blockchain.Block
		var limited bool
		var status View
		e.mu.Lock()
		e.busy = false
		if e.roundID == token {
			e.roundCancel = nil
		}
		accept := e.generation == gen && e.running && roundCtx.Err() == nil && err == nil
		if accept && next.Header.Prev == e.chain.Tip && e.blockCountLocked() < e.maxBlocks {
			hash := next.Header.Hash()
			e.chain.Blocks[hash] = next
			e.chain.Tip = hash
			committed = &next
			if e.reachedLimitLocked() {
				e.running = false
				e.stopReason = StopBlockLimit
				limited = true
				status = e.viewLocked()
			}
		}
		interval := e.interval
		e.cond.Broadcast()
		e.mu.Unlock()

		if committed != nil && emit != nil {
			emit(Notice{Kind: NoticeBlock, Block: *committed, Generation: gen})
			if limited {
				emit(Notice{Kind: NoticeStatus, Generation: gen, State: status})
			}
		}

		if (committed != nil && !limited) || (err != nil && ctx.Err() == nil && roundCtx.Err() == nil) {
			delay := interval
			if committed == nil {
				delay = time.Second
			}
			e.mu.Lock()
			if e.roundID == token && e.generation == gen {
				e.roundCancel = cancel
			}
			e.mu.Unlock()
			e.sleep(ctx, roundCtx, delay)
		}
		cancel()
		e.mu.Lock()
		if e.roundID == token {
			e.roundCancel = nil
		}
		e.mu.Unlock()
	}
}

// Pause cancela la ronda en curso y no vuelve hasta que ese candidato ya no puede entrar.
func (e *Engine) Pause() View {
	e.controlMu.Lock()
	defer e.controlMu.Unlock()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running = false
	if !e.reachedLimitLocked() {
		e.stopReason = ""
	}
	if e.roundCancel != nil {
		e.roundCancel()
	}
	for e.busy {
		e.cond.Wait()
	}
	e.cond.Broadcast()
	return e.viewLocked()
}

// Start reanuda el minado de la generación actual. Si ya está en curso, no crea otro worker.
func (e *Engine) Start() (View, error) {
	e.controlMu.Lock()
	defer e.controlMu.Unlock()
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.reachedLimitLocked() {
		return e.viewLocked(), ErrBlockLimit
	}
	if !e.running {
		e.running = true
		e.stopReason = ""
		e.cond.Broadcast()
	}
	return e.viewLocked(), nil
}

// Reset descarta la cadena, deja solo el génesis y arranca el minado de la generación nueva.
func (e *Engine) Reset() View {
	e.controlMu.Lock()
	defer e.controlMu.Unlock()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.generation++
	e.running = false
	if e.roundCancel != nil {
		e.roundCancel()
	}
	for e.busy {
		e.cond.Wait()
	}
	e.installGenesisLocked()
	e.cursor = 0
	e.stopReason = ""
	e.running = !e.reachedLimitLocked()
	if !e.running {
		e.stopReason = StopBlockLimit
	}
	e.cond.Broadcast()
	return e.viewLocked()
}

// View devuelve la cadena y el estado de ejecución.
func (e *Engine) View() View {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.viewLocked()
}

func (e *Engine) prepareLocked() (*miners.Node, blockchain.Chain, map[blockchain.Address]*miners.Node, int) {
	addr := e.addresses[e.cursor%len(e.addresses)]
	e.cursor++
	return e.nodes[addr], cloneChain(e.chain), e.nodes, e.difficulty
}

func (e *Engine) installGenesisLocked() {
	genesis := blockchain.NewGenesisBlock(e.supply)
	chain := blockchain.Chain{
		Blocks: map[blockchain.Hash]blockchain.Block{},
		Tip:    genesis.Header.Hash(),
		Owners: map[blockchain.SatoshiID]blockchain.Address{},
	}
	chain.Blocks[chain.Tip] = genesis
	chain, nodes := miners.NewDeterministicPool(chain, e.addresses)
	e.chain = chain
	e.nodes = nodes
}

func (e *Engine) blockCountLocked() int {
	return len(orderedBlocks(e.chain))
}

func (e *Engine) reachedLimitLocked() bool {
	return e.blockCountLocked() >= e.maxBlocks
}

func (e *Engine) viewLocked() View {
	blocks := orderedBlocks(e.chain)
	count := len(blocks)
	var height uint64
	if count > 0 {
		height = blocks[count-1].Header.Height
	}
	limit := count >= e.maxBlocks
	status := StatusPaused
	if e.running && !limit {
		status = StatusRunning
	}
	reason := e.stopReason
	if limit {
		reason = StopBlockLimit
	}
	return View{
		Blocks:       blocks,
		Supply:       e.supply,
		Difficulty:   e.difficulty,
		Interval:     e.interval,
		Status:       status,
		Generation:   e.generation,
		Height:       height,
		MaxBlocks:    e.maxBlocks,
		BlockCount:   count,
		LimitReached: limit,
		StopReason:   reason,
	}
}

func (e *Engine) sleep(ctx, roundCtx context.Context, delay time.Duration) {
	if delay <= 0 {
		return
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-roundCtx.Done():
	case <-timer.C:
	}
}
