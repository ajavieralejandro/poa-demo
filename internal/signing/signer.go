package signing

import (
	"errors"

	"poa-demo/internal/blockchain"
)

var (
	ErrNotImplemented = errors.New("signing: not implemented")
	ErrNoLocalKey     = errors.New("signing: falta la clave privada local")
	ErrNoMaterial     = errors.New("signing: falta la clave pública o la firma")
	ErrorNoIds        = errors.New("signing: falta id de las transacciones a enviar")
	ErrorIdsDuplicados = errors.New("signing: ids de satoshis duplicados")
	ErrorMeEmpty      = errors.New("signing: me es cero o invalida")
	ErrorSameAddress  = errors.New("signing: origen y destino son iguales")
)

type Local struct {
	Address blockchain.Address
	Public  []byte
	private []byte
}

func (l Local) Sign(message []byte) ([]byte, error) {
	if len(l.private) == 0 {
		return nil, ErrNoLocalKey
	}
	return nil, ErrNotImplemented
}

func Verify(public, message, sig []byte) error {
	if len(public) == 0 || len(sig) == 0 {
		return ErrNoMaterial
	}
	return ErrNotImplemented
}

func CreateTx(transaccion blockchain.Tx) error {
    err := ValidateTxParams(transaccion.Sender, transaccion.Recipient, transaccion.Satoshis)
    if err != nil {
        return err
    }
    
    // Aquí continuaría el flujo si la validación es exitosa...
    return nil
}

// ValidateTxParams verifica la validez de los parámetros básicos de una transacción.
func ValidateTxParams(me, to blockchain.Address, ids []blockchain.SatoshiID) error {
	// 1) Que las direcciones no sean cero (Address es uint32)
	if me == 0 || to == 0 {
		return ErrorMeEmpty
	}
	if me == to {
		return ErrorSameAddress
	}

	// 2) Que la lista de ids no sea vacía
	if len(ids) == 0 {
		return ErrorNoIds
	}

	// 3) Que no haya satoshis duplicados
	seen := make(map[blockchain.SatoshiID]bool)
	for _, id := range ids {
		_, existe := seen[id]
		if existe {
			return ErrorIdsDuplicados
		}
		seen[id] = true
	}

	return nil
}