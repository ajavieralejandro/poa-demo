package miners

import (
	// TODO: Importar los paquetes necesarios para hash y conversión numérica 
	// (ej: "crypto/sha256", "encoding/binary")
	"crypto/sha256"
	"encoding/binary"
	"poa-demo/internal/blockchain"

)

// Pool centraliza la gestión de todos los nodos mineros de la red.
type Pool struct {
	Miners []Miner
}

/package miner

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"

	"poa-demo/internal/blockchain"
)

// Estructura auxiliar interna para ordenar los mineros por su hash de forma determinista
type minerHashPair struct {
	addr    Address
	hashVal uint64
}

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
		// TODO 1: Preparar los datos a hashear para esta posición específica.
		// Pista: Necesitamos meter la cabecera 'h' (o algún campo clave de ella) 
		// y la variable 'pos' en un arreglo de bytes para que cada una de las 3 
		// vueltas genere un hash completamente distinto.
		var buf [5]byte
		binary.BigEndian.PutUint32(buf[0:4], h.Prev)
		// Guardamos la posición (pos) en el último byte (el byte en la posición 4)
		buf[4] = pos

		// TODO 2: Aplicar la función de hash criptográfico (ej: sha256.Sum256(...)) 
		// sobre esos bytes combinados.
		hashResult := sha256.Sum256(buf[:])
		// TODO 3: Convertir una porción del resultado del hash a un número entero grande 
		// Pista: Puedes usar binary.BigEndian.Uint64(...) con los primeros 8 bytes del hash.
		//esto no entiendo bien porque 
		numeroGrande := binary.BigEndian.Uint64(hashResult[0:8])
		// TODO 4: Calcular el SatoshiID ganador usando el operador módulo (%) con el supply, 
		// y asignarlo a la posición correspondiente:
		out[pos] = blockchain.SatoshiID(numeroGrande % uint64(supply))
	}

	return out, nil
}