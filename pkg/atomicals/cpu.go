//go:build cpu
// +build cpu

package atomicals

import (
	"bytes"
	"encoding/binary"
	"go-atomicals/pkg/hashrate"
	"log"
	"runtime"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

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

func mine(i int, input Input, result chan<- Result, reporter *hashrate.HashRateReporter) {
	// set different time for each goroutine
	input.CopiedData.Args.Time += uint32(i)
	// use uint32 so we can avoid cbor encoding at runtime
	input.CopiedData.Args.Nonce = uint32(^uint16(0)) + 1
	input.Init()

	msgTx := wire.NewMsgTx(wire.TxVersion)
	output := wire.NewOutPoint(input.FundingUtxo.Txid, input.FundingUtxo.Index)
	txIn := wire.NewTxIn(output, nil, nil)
	txIn.Sequence = 0
	msgTx.AddTxIn(txIn)

	scriptP2TR := input.ScriptP2TR(input.UpdateScript())
	txOut := wire.NewTxOut(int64(input.Fees.RevealFeePlusOutputs), scriptP2TR.Output)
	msgTx.AddTxOut(txOut)
	// add change utxo
	if change := input.GetCommitChange(); change != 0 {
		msgTx.AddTxOut(wire.NewTxOut(change, input.KeyPairInfo.Ouput))
	}

	buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSizeStripped()))
	msgTx.SerializeNoWitness(buf)
	serializedTx := buf.Bytes()
	localCounter := 0
	for {
		hash := chainhash.DoubleHashH(serializedTx)
		if input.WorkerBitworkInfoCommit.HasValidBitwork(&hash) {
			log.Printf("worker %d found args %+v", i, input.CopiedData.Args)
			PrintMsgTx(msgTx)
			break
		}
		localCounter++
		txIn.Sequence++
		binary.LittleEndian.PutUint32(serializedTx[42:], txIn.Sequence)
		if txIn.Sequence == MAX_SEQUENCE {
			input.CopiedData.Args.Nonce++
			scriptP2TR := input.ScriptP2TR(input.UpdateScript())
			txOut.PkScript = scriptP2TR.Output
			txIn.Sequence = 0
		}
		if localCounter == 102400 {
			reporter.Report(uint64(localCounter))
			localCounter = 0
		}
	}
	result <- Result{
		FinalCopyData: input.CopiedData,
		FinalSequence: txIn.Sequence,
	}
}

func Mine(input Input, result chan<- Result, reporter *hashrate.HashRateReporter) {
	for i := 0; i < runtime.NumCPU(); i++ {
		go mine(i, input, result, reporter)
	}
}
