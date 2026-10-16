package miner

import (
	// TODO: Importar los paquetes necesarios para hash y conversión numérica 
	// (ej: "crypto/sha256", "encoding/binary")
)

// Pool centraliza la gestión de todos los nodos mineros de la red.
type Pool struct {
	Miners []Miner
}

// NewDeterministicPool inicializa el pool y reparte los satoshis del supply 
// entre los mineros de forma determinista usando hashing de sus direcciones.
func NewDeterministicPool(addresses []Address, supply uint32) *Pool {
	// TODO 1: Calcular cuántos tokens base le corresponden a cada minero 
	// (por ejemplo: supply / número total de direcciones).

	// TODO 2: Iterar sobre cada Address:
	// a) Aplicar la función de hash (ej: sha256.Sum256([]byte(address...)))
	// b) Convertir los primeros bytes del hash resultante a un número entero uint64 
	//    usando binary.BigEndian.Uint64(...)
	// c) Aplicar un módulo (%) con el supply para asegurarse de que caiga 
	//    dentro del rango válido de IDs de satoshis.

	// TODO 3: Crear las instancias de Miner con sus respectivos SatoshiIDs asignados 
	// y agregarlos a la estructura Pool.

	return &Pool{
		Miners: []Miner{}, // Reemplazar con los mineros inicializados
	}
}

// FindOwner busca en el pool qué minero es el propietario de un SatoshiID dado.
func (p *Pool) FindOwner(satID SatoshiID) (Miner, error) {
	// TODO: Recorrer los mineros del pool y revisar sus listas o rangos de satoshis
	// para retornar el minero dueño del satID, o un error si no se encuentra.
	return Miner{}, nil
}