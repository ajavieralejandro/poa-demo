package miners

import (
	// TODO: Importar los paquetes necesarios para hash y conversión numérica
	// (ej: "crypto/sha256", "encoding/binary")
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"poa-demo/internal/blockchain"
	"sort"
)

var (
	ErrEmptySupply = errors.New("Error : empty suplly, no tokens. ")
)

// NewDeterministicPool inicializa el pool y reparte los satoshis del supply
// entre los mineros de forma determinista usando hashing de sus direcciones.
func NewDeterministicPool(addresses []Address, supply uint32) *Pool {
	numMiners := uint32(len(addresses))
	tokensPorMinero := supply / numMiners

	// 1. Calcular el hash criptográfico para cada dirección y pasarlo a uint64
	pairs := make([]minerHashPair, len(addresses))
	for i, addr := range addresses {
		// Convertir la dirección (uint32) a bytes para poder hacerle el hash
		addrBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(addrBytes, uint32(addr))

		// a) Aplicar función SHA-256
		hashSum := sha256.Sum256(addrBytes)

		// b) Convertir los primeros 8 bytes del resultado a un entero uint64
		hashVal := binary.BigEndian.Uint64(hashSum[0:8])

		pairs[i] = minerHashPair{
			addr:    addr,
			hashVal: hashVal,
		}
	}

	// 2. Ordenar los mineros según su valor de hash numérico.
	// Esto garantiza que todos los nodos de la red ordenen a los mineros exactamente igual.
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].hashVal < pairs[j].hashVal
	})

	// 3. Asignar los rangos de SatoshiIDs de manera contigua sin huecos
	miners := make([]Miner, numMiners)
	for i := uint32(0); i < numMiners; i++ {
		inicio := i * tokensPorMinero

		// El último minero absorbe cualquier residuo matemático por redondeo del supply
		var fin uint32
		if i == numMiners-1 {
			fin = supply - 1
		} else {
			fin = (i+1)*tokensPorMinero - 1
		}

		// Rellenar la lista de SatoshiIDs correspondientes a este minero
		var satoshis []SatoshiID
		for s := inicio; s <= fin; s++ {
			satoshis = append(satoshis, SatoshiID(s))
		}

		miners[i] = Miner{
			Addr:     pairs[i].addr,
			Satoshis: satoshis,
			Blocks:   []blockchain.Block{},
		}
	}

	return &Pool{
		Miners: miners,
	}
}

// FindOwner busca en el pool qué minero es el propietario de un SatoshiID dado.
func (p *Pool) FindOwner(satID SatoshiID) (Miner, error) {
	for _, miner := range p.Miners {
		for _, id := range miner.Satoshis {
			if id == satID {
				return miner, nil
			}
		}
	}
	return Miner{}, nil
}

// Select deriva tres índices de satoshis a partir de la cabecera minada (Fase 2: Follow the Satoshi).
// supply es la cantidad de satoshis existentes (0 .. supply-1).
func Select(h blockchain.Header, supply uint32) ([3]blockchain.SatoshiID, error) {
	if supply == 0 {
		return [3]blockchain.SatoshiID{}, ErrEmptySupply
	}

	var out [3]blockchain.SatoshiID

	// Iteramos 3 veces para obtener los 3 cupos de validadores (posiciones 0, 1 y 2)
	for pos := uint8(0); pos < 3; pos++ {
		// 1. Preparamos el búfer de 33 bytes:
		// Los primeros 32 bytes con el hash anterior (h.Prev)
		// El último byte (la posición 32) con el valor de 'pos'
		var buf [33]byte
		copy(buf[0:32], h.Prev[:])
		buf[32] = pos

		// 2. Aplicamos la función de hash criptográfico sobre esos 33 bytes combinados
		hashResult := sha256.Sum256(buf[:])

		// 3. Tomamos los primeros 8 bytes del hash resultante y los convertimos a uint64
		numeroGrande := binary.BigEndian.Uint64(hashResult[0:8])

		// 4. Calculamos el SatoshiID ganador usando el operador módulo (%) con el supply
		out[pos] = blockchain.SatoshiID(numeroGrande % uint64(supply))
	}

	return out, nil
}
