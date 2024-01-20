package atomicals

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

const MAX_SEQUENCE = 0xffffffff

type Result struct {
	FinalCopyData CopiedData      `json:"finalCopyData"`
	FinalSequence uint32          `json:"finalSequence"`
	Embed         string          `json:"embed"`
	CommitTxHash  *chainhash.Hash `json:"commitTxHash"`
	RevealTxhash  *chainhash.Hash `json:"revealTxhash"`
}

func PrintMsgTx(msgTx *wire.MsgTx) {
	log.Printf("version: %d", msgTx.Version)
	for _, txIn := range msgTx.TxIn {
		log.Printf("txin: %s %d sig: %x", txIn.PreviousOutPoint.String(), txIn.Sequence, txIn.SignatureScript)
	}
	for _, txOut := range msgTx.TxOut {
		log.Printf("txout: %d %x", txOut.Value, txOut.PkScript)
	}
	log.Printf("locktime: %d", msgTx.LockTime)
	log.Printf("hash: %s", msgTx.TxHash().String())
}

func DecAndPrintTx(hexTx string) {
	var msgTx wire.MsgTx
	data, _ := hex.DecodeString(hexTx)
	if err := msgTx.Deserialize(bytes.NewReader((data))); err != nil {
		log.Fatalf("msgTx.Deserialize failed: %v", err)
	}
	PrintMsgTx(&msgTx)
}
