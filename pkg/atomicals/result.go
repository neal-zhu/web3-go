package atomicals

import (
	"log"

	"github.com/btcsuite/btcd/wire"
)

const MAX_SEQUENCE = 0xffffffff

type Result struct {
	FinalCopyData CopiedData `json:"finalCopyData"`
	FinalSequence uint32     `json:"finalSequence"`
}

func PrintMsgTx(msgTx *wire.MsgTx) {
	log.Printf("version: %d", msgTx.Version)
	for _, txIn := range msgTx.TxIn {
		log.Printf("txin: %s %d", txIn.PreviousOutPoint.String(), txIn.Sequence)
	}
	for _, txOut := range msgTx.TxOut {
		log.Printf("txout: %d %x", txOut.Value, txOut.PkScript)
	}
	log.Printf("locktime: %d", msgTx.LockTime)
	log.Printf("hash: %s", msgTx.TxHash().String())
}
