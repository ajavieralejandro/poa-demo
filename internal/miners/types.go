import(
		"poa-demo/internal/blockchain"

)
type Address uint32

type SatoshiID uint32


type Miner struct{
	Addr Address
	Satoshis []SatoshiID
	Blocks []Blocks

}