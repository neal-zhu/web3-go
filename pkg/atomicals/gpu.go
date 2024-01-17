//go:build cuda
// +build cuda

package atomicals

// #include <stdint.h>
//uint32_t scanhash_sha256d(int thr_id, unsigned char* in, unsigned int inlen, unsigned char *target, unsigned int target_len, char pp, char ext, unsigned int threads, unsigned int start_seq, unsigned int *hashes_done);
//#cgo LDFLAGS: -L. -L../../cuda -lhash
import "C"
import (
	"bytes"
	"go-atomicals/pkg/hashrate"
	"log"
	"os"

	"github.com/btcsuite/btcd/wire"
)

func Mine(input Input, result chan<- Result, reporter *hashrate.HashRateReporter) {
	deviceNum := 1
	devcieNumStr := os.Getenv("CUDA_DEVICE_NUM")
	if devcieNumStr != "" {
		deviceNum = int(devcieNumStr[0] - '0')
	}
	for i := 0; i < deviceNum; i++ {
		go mine(i, input, result, reporter)
	}
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

	scriptP2TR := input.MustBuildScriptP2TR()
	txOut := wire.NewTxOut(int64(input.Fees.RevealFeePlusOutputs), scriptP2TR.Output)
	msgTx.AddTxOut(txOut)
	// add change utxo
	if change := input.GetCommitChange(); change != 0 {
		msgTx.AddTxOut(wire.NewTxOut(change, input.KeyPairInfo.Ouput))
	}

	buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSizeStripped()))
	msgTx.SerializeNoWitness(buf)
	serializedTx := buf.Bytes()

	hashesDone := C.uint(0)
	var (
		pp  = -1
		ext = -1
	)
	if input.WorkerBitworkInfoCommit.PrefixPartial != nil {
		pp = int(*input.WorkerBitworkInfoCommit.PrefixPartial)
	}
	if input.WorkerBitworkInfoCommit.Ext != 0 {
		ext = int(input.WorkerBitworkInfoCommit.Ext)
	}
	for {
		seq := C.scanhash_sha256d(
			C.int(i), // device id
			(*C.uchar)(&serializedTx[0]),
			C.uint(len(serializedTx)),
			(*C.uchar)(&input.WorkerBitworkInfoCommit.PrefixBytes[0]),
			C.uint(len(input.WorkerBitworkInfoCommit.PrefixBytes)),
			C.char(pp),
			C.char(ext),
			C.uint(1<<25),
			C.uint(txIn.Sequence),
			&hashesDone,
		)
		log.Printf("device: %d, seq: %d, hashesDone: %d", i, seq, hashesDone)
		if uint32(seq) != MAX_SEQUENCE {
			txIn.Sequence = uint32(seq)
			break
		}

		input.CopiedData.Args.Nonce++
		scriptP2TR := input.MustBuildScriptP2TR()
		txOut.PkScript = scriptP2TR.Output
		txIn.Sequence = 0
		buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSizeStripped()))
		msgTx.SerializeNoWitness(buf)
		serializedTx = buf.Bytes()
	}

	PrintMsgTx(msgTx)
	result <- Result{
		FinalCopyData: input.CopiedData,
		FinalSequence: txIn.Sequence,
	}
}
