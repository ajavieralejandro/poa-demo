// Package consensus reserva las fases de Proof of Activity.
//
// Proof of Activity no es Proof of Authority, ni consiste en reunir la
// firma de todos los usuarios. Los participantes y el umbral quedan
// pendientes de definir.
package consensus

import (
	"errors"

	"poa-demo/internal/blockchain"
)

// ErrNotImplemented lo devuelve toda operación de consenso pendiente.
var ErrNotImplemented = errors.New("consensus: not implemented")

// Selector elige a los validadores que deben firmar un bloque candidato.
// El conjunto de candidatos y el umbral de aceptación no están definidos.
type Selector interface {
	SelectValidators(candidates []blockchain.Validator) ([]blockchain.Validator, error)
}

// Activity ejecuta las cuatro fases de Proof of Activity, en este orden:
//  1. Minar un encabezado candidato con prueba de trabajo.
//  2. Seleccionar validadores según su participación.
//  3. Recoger las firmas de los validadores seleccionados.
//  4. Finalizar el bloque e incorporarlo a la cadena.
type Activity interface {
	MineCandidateHeader(block *blockchain.Block) error
	SelectValidatorsByParticipation(candidates []blockchain.Validator) ([]blockchain.Validator, error)
	CollectSelectedSignatures(block *blockchain.Block, selected []blockchain.Validator) error
	FinalizeAndAppend(chain *blockchain.Chain, block *blockchain.Block) error
}

// Pending reserva la selección y las cuatro fases.
// No elige validadores, no firma y no modifica la cadena.
type Pending struct{}

var (
	_ Selector = Pending{}
	_ Activity = Pending{}
)

func (Pending) SelectValidators([]blockchain.Validator) ([]blockchain.Validator, error) {
	return nil, ErrNotImplemented
}

func (Pending) MineCandidateHeader(*blockchain.Block) error {
	return ErrNotImplemented
}

func (Pending) SelectValidatorsByParticipation([]blockchain.Validator) ([]blockchain.Validator, error) {
	return nil, ErrNotImplemented
}

func (Pending) CollectSelectedSignatures(*blockchain.Block, []blockchain.Validator) error {
	return ErrNotImplemented
}

func (Pending) FinalizeAndAppend(*blockchain.Chain, *blockchain.Block) error {
	return ErrNotImplemented
}
