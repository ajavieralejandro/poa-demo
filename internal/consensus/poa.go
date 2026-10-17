// Package consensus reserva la minería y la elección de tres satoshis.
//
// La primera versión corre en un proceso y extiende una cadena lineal.
// Elegir un satoshi no depende de una lista de candidatos ni de quién esté
// conectado: el dueño se consulta después en el estado confirmado.
// Bifurcaciones, reorganización y red quedan fuera de este paquete.
package consensus

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"poa-demo/internal/blockchain"
)

var (
	// ErrNotImplemented lo devuelven minería y selección.
	// Ninguna de las dos acepta un bloque ni devuelve un candidato válido.
	ErrNotImplemented = errors.New("consensus: not implemented")
	ErrIncorrectBlock = errors.New("block : not dificulty values")
	ErrSignatureFail = errors.New("block  : different signatures error")
	ErrEmptySupply = errors.New("consensus: no hay satoshis para elegir")
)

// Mine busca un nonce para la cabecera hija de tip.
// Cuando exista, devolverá solo la cabecera. No escribe en la cadena.
func Mine(tip blockchain.Block, miner blockchain.Address, bits int) (blockchain.Header, error) {
	// 1. Preparamos la cabecera base conectándola con el bloque anterior (tip)
	// Pista: El Height debe ser tip.Header.Height + 1, y el Prev debe ser el Hash del tip.
	prevHash := tip.Header.Hash() // Asumiendo que Hash() es el método de la cabecera

    h := blockchain.Header{
        Prev:   prevHash,
        Miner:  miner,
        Height: tip.Header.Height + 1,
        Nonce:  0,
    }

	// 2. Bucle de búsqueda (Proof of Work)
	for {
		// TODO: Implementar o llamar a la función que calcula el hash de la cabecera actual (h)
		currentHash := h.Hash()
		// TODO: Implementar o llamar a la función que verifica la dificultad (ej: contar ceros iniciales)
		 if cumpleDificultad(currentHash, bits) {
		     return h, nil
		 }

		// Si no se cumple, incrementamos el nonce para probar con el siguiente
		h.Nonce++}
	return blockchain.Header{}, ErrNotImplemented
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

func Select(h blockchain.Header, supply uint32) ([3]blockchain.SatoshiID, error) {
	if supply == 0 {
		return [3]blockchain.SatoshiID{}, ErrEmptySupply
	}

	var out [3]blockchain.SatoshiID

	for pos := uint8(0); pos < 3; pos++ {
		// 1. Obtener el hash base de la cabecera
		hashBase := h.Hash()

		// 2. Combinar el hash base con la posición (pos) y volver a hashear
		dataToHash := append(hashBase[:], pos)
		nuevoHash := sha256.Sum256(dataToHash)

		// 3. Convertir los primeros 8 bytes del hash en uint64
		valorAleatorio := binary.BigEndian.Uint64(nuevoHash[0:8])

		// 4. Rejection sampling: si cae en la zona prohibida, repetimos la iteración (pos--)
		if !esValidoSinSesgo(valorAleatorio, supply) {
			pos--
			continue
		}

		// 5. Aplicar módulo y asignar el SatoshiID convertido explícitamente
		indice := valorAleatorio % uint64(supply)
		out[pos] = blockchain.SatoshiID(indice)
	}

	return out, nil
}

// Verify valida si un bloque recibido por la red es totalmente correcto.
// Comprueba el Proof of Work y que los validadores firmantes coincidan con el sorteo (independientemente del orden).
func Verify(block blockchain.Block, parent blockchain.Block, bits int, supply uint32) (bool, error) {
	
	// 1. Verificar el Proof of Work de la cabecera del bloque actual.
	hashActual := block.Header.Hash()
	if !cumpleDificultad(hashActual, bits) {
		return false, errors.New("consensus: el bloque no cumple con la dificultad de PoW requerida")
	}

	// 2. Volver a ejecutar Select utilizando la cabecera del bloque PADRE (parent)
	validadoresEsperados, err := Select(parent.Header, supply)
	if err != nil {
		return false, err
	}

	// 3. Robustez: Verificar que cada firmante del bloque esté presente en los validadores esperados,
	// sin importar el orden en el que vengan declarados.
	for _, firmante := range block.Signers {
		encontrado := false
		for _, esperado := range validadoresEsperados {
			if firmante == esperado {
				encontrado = true
				break
			}
		}
		
		// Si algún firmante no forma parte del sorteo oficial, el bloque es inválido.
		if !encontrado {
			return false, errors.New("consensus: el bloque incluye un firmante que no fue seleccionado en el sorteo")
		}
	}

	// Opcional: También asegurarnos de que la cantidad de firmantes coincida exactamente (ej: los 3 cupos completos)
	if len(block.Signers) != len(validadoresEsperados) {
		return false, errors.New("consensus: la cantidad de firmas no coincide con los cupos requeridos")
	}

	return true, nil
}