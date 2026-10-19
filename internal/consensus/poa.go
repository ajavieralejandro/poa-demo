// Package consensus reserva la minería y la elección de tres satoshis.
//
// La primera versión corre en un proceso y extiende una cadena lineal.
// Elegir un satoshi no depende de una lista de candidatos ni de quién esté
// conectado: el dueño se consulta después en el estado confirmado.
// Bifurcaciones, reorganización y red quedan fuera de este paquete.
package consensus

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"poa-demo/internal/blockchain"
	"slices"
)

var (
	// ErrNotImplemented lo devuelven minería y selección.
	// Ninguna de las dos acepta un bloque ni devuelve un candidato válido.
	ErrNotImplemented = errors.New("consensus: not implemented")
	ErrIncorrectBlock = errors.New("block : not dificulty values")
	ErrSignatureFail  = errors.New("block  : different signatures error")
	ErrEmptySupply    = errors.New("consensus: no hay satoshis para elegir")
	ErrMultySign      = errors.New("consensus: Header proposed can be signed twice by the same validator")
)

func Mine(tip blockchain.Block, chain blockchain.Chain, miner blockchain.Address, bits int) (blockchain.Block, error) {
	return MineWithContext(context.Background(), tip, chain, miner, bits)
}

// MineWithContext busca el nonce igual que Mine y abandona la búsqueda si ctx se cancela.
// La regla de dificultad, el sorteo y las firmas no cambian.
func MineWithContext(ctx context.Context, tip blockchain.Block, chain blockchain.Chain, miner blockchain.Address, bits int) (blockchain.Block, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	prevHash := tip.Header.Hash()

	h := blockchain.Header{
		Prev:   prevHash,
		Miner:  miner,
		Height: tip.Header.Height + 1,
		Nonce:  0,
		Supply: tip.Header.Supply,
	}

	for {
		if err := ctx.Err(); err != nil {
			return blockchain.Block{}, err
		}
		currentHash := h.Hash()
		if cumpleDificultad(currentHash, bits) {
			// ¡PoW encontrado! Ahora ejecutamos el sorteo usando el padre (tip.Header)
			_, err := Select(tip.Header, chain)
			if errors.Is(err, ErrMultySign) {
				h.Nonce++
				continue
			}
			if err != nil && !errors.Is(err, ErrMultySign) {
				return blockchain.Block{}, err
			}

			// Convertimos los SatoshiID ganadores en un arreglo de Firmas (Signatures)
			var sigs [3]blockchain.Signature

			// Retornamos el bloque completamente armado y firmado
			return blockchain.Block{
				Header:     h,
				Signatures: sigs,
				Txs:        []blockchain.Tx{},
			}, nil
		}

		h.Nonce++
	}
}
func cumpleDificultad(hash blockchain.Hash, bits int) bool {
	// Definimos cuántos bytes completos deben ser 0 (asumiendo que 'bits' es múltiplo de 8, o ajustándolo)
	bytesRequeridos := bits / 8
	if bytesRequeridos > len(hash) {
		return false
	}

	for i := 0; i < bytesRequeridos; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	return true
}

// Select deriva tres índices a partir de la cabecera, su padre y la posición.
// supply es la cantidad de satoshis existentes, con identificadores 0..supply-1.
// El módulo tiene que descartar el resto del rango cuando supply no es potencia de dos,
// para no favorecer índices bajos (Rejection Sampling).
// esValidoSinSesgo verifica si el valor aleatorio está por debajo del límite que genera sesgo.
func esValidoSinSesgo(valorAleatorio uint64, supply uint32) bool {
	maxVal := (^uint64(0)) - ((^uint64(0)) % uint64(supply))
	return valorAleatorio < maxVal
}

func Select(h blockchain.Header, chain blockchain.Chain) ([3]blockchain.SatoshiID, error) {
	if h.Supply == 0 {
		return [3]blockchain.SatoshiID{}, ErrEmptySupply
	}

	var out [3]blockchain.SatoshiID
	owners := make(map[blockchain.Address]struct{})

	for pos := uint8(0); pos < 3; pos++ {
		// 1. Obtener el hash base de la cabecera
		hashBase := h.Hash()

		// 2. Combinar el hash base con la posición (pos) y volver a hashear
		dataToHash := append(hashBase[:], pos)
		nuevoHash := sha256.Sum256(dataToHash)

		// 3. Convertir los primeros 8 bytes del hash en uint64
		valorAleatorio := binary.BigEndian.Uint64(nuevoHash[0:8])

		// 4. Rejection sampling: si cae en la zona prohibida, repetimos la iteración (pos--)
		if !esValidoSinSesgo(valorAleatorio, h.Supply) {
			pos--
			continue
		}

		// 5. Aplicar módulo y asignar el SatoshiID convertido explícitamente
		indice := valorAleatorio % uint64(h.Supply)
		out[pos] = blockchain.SatoshiID(indice)
		if _, ok := owners[chain.Owners[out[pos]]]; ok {
			return out, ErrMultySign
		}
		owners[chain.Owners[out[pos]]] = struct{}{}
		fmt.Printf("[DEBUG Select] Posición %d -> SatoshiID elegido: %d\n", pos, indice)

	}

	return out, nil
}

// Verify valida si un bloque recibido por la red es totalmente correcto.
// Comprueba el Proof of Work y que los validadores firmantes coincidan con el sorteo (independientemente del orden).
func Verify(block blockchain.Block, chain blockchain.Chain, bits int) error {
	//Agregamos esta condición para diferenciar entre el bloque genesis
	if block.Header.Height == 0 {
		return nil
	}

	if !slices.Equal(block.Header.Prev[:], chain.Tip[:]) {
		return ErrIncorrectBlock
	}

	// 1. Verificar el Proof of Work de la cabecera del bloque actual.
	hashActual := block.Header.Hash()
	if !cumpleDificultad(hashActual, bits) {
		return errors.New("consensus: el bloque no cumple con la dificultad de PoW requerida")
	}

	// 2. Volver a ejecutar Select utilizando la cabecera del bloque PADRE (parent)
	validadoresEsperados, err := Select(block.Header, chain)
	if err != nil {
		return err
	}

	fmt.Printf("[DEBUG Verify] Validadores esperados por el padre: %v\n", validadoresEsperados)
	fmt.Printf("[DEBUG Verify] Firmantes presentes en el bloque: %v\n", block.Signatures)

	// 3. Robustez: Verificar que cada firmante del bloque esté presente en los validadores esperados,
	// sin importar el orden en el que vengan declarados.
	for _, firma := range block.Signatures {
		encontrado := false
		for _, esperado := range validadoresEsperados {
			//Firma.Valid() para protejer si SatoshiID == 0 y justo no tiene firmante
			if firma.Valid() && firma.Value == esperado {
				encontrado = true
				break
			}
		}

		// Si algún firmante no forma parte del sorteo oficial, el bloque es inválido.
		if !encontrado {
			return errors.New("consensus: el bloque incluye un firmante que no fue seleccionado en el sorteo")
		}
	}

	return nil
}
