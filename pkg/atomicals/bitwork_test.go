package atomicals_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-atomicals/pkg/atomicals"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

var inputStr = `{"copiedData":{"args":{"bitworkc":"000000","bitworkr":"6238","mint_ticker":"sophon","nonce":9999999,"time":1705918935}},"nonceStart":0,"nonceEnd":9999999,"workerOptions":{"electrumApi":{"baseUrl":"http://ep.atomicalneutron.com/proxy","usePost":true,"isOpenFlag":false},"satsbyte":10,"address":"bc1pq9a5tkcc987mknndz5fgrsj9ateyu046v6majxnzwpkxwy2t87nqygunry","opType":"dmt","dmtOptions":{"mintAmount":100000,"ticker":"sophon"}},"fundingWIF":"L4cjYizvfRVpjLNjZDqTYuKD5fJugNoYkTkYFDjpw21UrL5E4JT1","fundingUtxo":{"txid":"0000008674690288a63dd83588d3a765a45a02aa9b6954b7eba16daf58507006","vout":1,"status":{"confirmed":true,"block_height":826688,"block_hash":"0000000000000000000169fe9dbf52a11829a4c730e38b3d81fe11f399f4ca3c","block_time":1705844320},"value":17437,"index":1},"fees":{"commitAndRevealFee":2920,"commitAndRevealFeePlusOutputs":102920,"revealFeePlusOutputs":101810,"commitFeeOnly":1110,"revealFeeOnly":1810},"performBitworkForCommitTx":true,"workerBitworkInfoCommit":{"input_bitwork":"000000","hex_bitwork":"000000","prefix":"000000"},"scriptP2TR":null,"hashLockP2TR":null,"workerBitworkInfoReveal":{"input_bitwork":"6238","hex_bitwork":"6238","prefix":"6238"},"additionalOutputs":[{"address":"bc1pq9a5tkcc987mknndz5fgrsj9ateyu046v6majxnzwpkxwy2t87nqygunry","value":100000}]}`
var input atomicals.Input

func init() {
	err := json.Unmarshal([]byte(inputStr), &input)
	if err != nil {
		panic(err)
	}
	input.Init()
}

func BenchmarkHash1(b *testing.B) {
	w := []byte("hello")
	for i := 0; i < b.N; i++ {
		chainhash.DoubleHashB(w)
	}

}

func BenchmarkHash2(b *testing.B) {
	w := []byte("hello")
	hasher := sha256.New()
	first := make([]byte, 32)
	for i := 0; i < b.N; i++ {
		hasher.Reset()
		hasher.Write(w)
		hasher.Sum(first[:0])
		hasher.Reset()
		hasher.Write(first)
		hasher.Sum(nil)
	}

}

func TestKeyPairInfo(t *testing.T) {
	keyPairInfo := input.MustBuildKeyPairInfo()
	if hex.EncodeToString(keyPairInfo.Ouput) != "51202c10a9002c3a1825e509cfdcd3fc1aec72b48e2c7a970a3ff6dfad4f933f6c9a" {
		t.Fatalf("output %x", keyPairInfo.Ouput)
	}
	if hex.EncodeToString(keyPairInfo.ChildNode.Serialize()) != "5770855d9e2ea845416d88f681e45a153327cd3f27c7fe57a48f4017a5bc9e0b" {
		t.Fatalf("childNode %x", keyPairInfo.ChildNode.Serialize())
	}
	if hex.EncodeToString(keyPairInfo.TweakedChildNode.Serialize()) != "2baf8ac2324dc17bd3b9bb24e13eda12e46e4bf00f03f813f076b09361615548" {
		t.Fatalf("tweakedChildNode %x", keyPairInfo.TweakedChildNode.Serialize())
	}
	if hex.EncodeToString(keyPairInfo.ChildNodeXOnlyPubkey) != "2b540a6937b561458734e82ec9392a8449e97f9c8093948ae0b9419a0d92ad82" {
		t.Fatalf("childNodeXOnlyPubkey %x", keyPairInfo.ChildNodeXOnlyPubkey)
	}

	addr, err := btcutil.DecodeAddress("bc1p9sg2jqpv8gvztegfelwd8lq6a3etfr3v02ts50lkm7k5lyeldjdqxj4cl2", &chaincfg.MainNetParams)
	if err != nil {
		fmt.Println("Error decoding address:", err)
		return
	}

	t.Logf("Xonly Pubkey: %x\n", addr.ScriptAddress())
	t.Logf("%x", keyPairInfo.ChildNodeXOnlyPubkey)
}

func TestCBOR(t *testing.T) {
	data := `{
		"args": {
		  "bitworkc": "000000",
		  "bitworkr": "6238",
		  "mint_ticker": "sophon",
		  "nonce": 9999999,
		  "time": 1705493802
		}
	}`
	var copiedData atomicals.CopiedData
	if err := json.Unmarshal([]byte(data), &copiedData); err != nil {
		t.Fatalf("json.Unmarshal([]byte(data), &copiedData) failed: %v", err)
	}

	encoded := copiedData.MustEncodeCbor()
	if hex.EncodeToString(encoded) != "a16461726773a56474696d651a65a7c52a656e6f6e63651a0098967f68626974776f726b636630303030303068626974776f726b7264363233386b6d696e745f7469636b657266736f70686f6e" {
		t.Fatalf("encoded %x", encoded)
	}
}

func TestBuildScript(t *testing.T) {
	input := input
	expectedPayload := "a16461726773a46474696d651a65899a24656e6f6e63651a0098967f68626974776f726b63666161626263636b6d696e745f7469636b657265717561726b"
	cborData := input.MustEncodeCbor()
	if hex.EncodeToString(cborData) != expectedPayload {
		t.Fatalf("cborData %x expect %s", cborData, expectedPayload)
	}

	script := input.MustBuildScript(cborData)
	if hex.EncodeToString(script) != "202b540a6937b561458734e82ec9392a8449e97f9c8093948ae0b9419a0d92ad82ac00630461746f6d03646d743ea16461726773a46474696d651a65899a24656e6f6e63651a0098967f68626974776f726b63666161626263636b6d696e745f7469636b657265717561726b68" {
		t.Fatalf("script %x", script)
	}

	for i := 0; i < 100; i++ {
		input.CopiedData.Args.Nonce += 1
		cborData := input.MustEncodeCbor()
		bytes.Equal(cborData, input.UpdateCborCache())
		bytes.Equal(input.MustBuildScript(cborData), input.UpdateScript())
	}
}

func TestScriptP2TR(t *testing.T) {
	scriptP2TR := input.ScriptP2TR(input.MustBuildScript(input.MustEncodeCbor()))
	if hex.EncodeToString(scriptP2TR.Output) != "51200520b9ace3bfb80bb7997527c16595df01d9417dec84c02a9abf502a211b0f63" {
		t.Fatalf("scriptP2TR %x", scriptP2TR.Output)
	}
}

func BenchmarkBitworkInfo(b *testing.B) {
	bw := atomicals.BitworkInfo{
		Prefix: "aabbccde",
	}
	bw.ParsePreifx()
	hash, _ := chainhash.NewHashFromStr("aabbccde")
	for i := 0; i < b.N; i++ {
		bw.HasValidBitwork(hash)
	}
}

func TestBitworkInfo(t *testing.T) {
	type Test struct {
		hash   string
		prefix string
		ext    int
		expect bool
	}
	data := `{
		"prefix": "2b540a",
		"ext": 2
	}`
	var bw atomicals.BitworkInfo
	if err := json.Unmarshal([]byte(data), &bw); err != nil {
		t.Fatalf("json.Unmarshal([]byte(data), &bw) failed: %v", err)
	}
	bw.ParsePreifx()
	t.Logf("%+v", bw)

	for _, test := range []Test{
		{
			hash:   "2b540a6937b561458734e82ec9392a8449e97f9c8093948ae0b9419a0d92ad82",
			prefix: "2b540a",
			expect: true,
		},
		{
			hash:   "2b540a6937b561458734e82ec9392a8449e97f9c8093948ae0b9419a0d92ad82",
			prefix: "2b540b",
			expect: false,
		},
		{
			hash:   "2b541a6937b561458734e82ec9392a8449e97f9c8093948ae0b9419a0d92ad82",
			prefix: "2b541",
			expect: true,
		},
		{
			hash:   "2b541a6937b561458734e82ec9392a8449e97f9c8093948ae0b9419a0d92ad82",
			prefix: "2b543",
			expect: false,
		},
		{
			hash:   "2b543a6937b561458734e82ec9392a8449e97f9c8093948ae0b9419a0d92ad82",
			prefix: "2b543",
			ext:    10,
			expect: true,
		},
		{
			hash:   "2b543b6937b561458734e82ec9392a8449e97f9c8093948ae0b9419a0d92ad82",
			prefix: "2b543",
			ext:    10,
			expect: true,
		},
		{
			hash:   "2b543b6937b561458734e82ec9392a8449e97f9c8093948ae0b9419a0d92ad82",
			prefix: "2b543b",
			ext:    3,
			expect: true,
		},
		{
			hash:   "2b543b6937b561458734e82ec9392a8449e97f9c8093948ae0b9419a0d92ad82",
			prefix: "2b543b",
			ext:    7,
			expect: false,
		},
		{
			hash:   "2b543a6937b561458734e82ec9392a8449e97f9c8093948ae0b9419a0d92ad82",
			prefix: "2b543",
			ext:    13,
			expect: false,
		},
	} {
		bw := atomicals.BitworkInfo{Prefix: test.prefix}
		if test.ext != 0 {
			bw.Ext = byte(test.ext)
		}
		bw.ParsePreifx()
		hash, err := chainhash.NewHashFromStr(test.hash)
		if err != nil {
			t.Fatalf("chainhash.NewHashFromStr(test.hash) failed: %v", err)
		}
		if bw.HasValidBitwork(hash) != test.expect {
			t.Fatalf("bw.HasValidBitwork(hash) != test.expect %v", test)
		}
	}

}

func TestTx(t *testing.T) {
	/*
					2024/01/17 23:42:39 txin: aabbccc8daf7540db4841104e5ccd7b899691f3d6a166e5b00e1a66da0087ad5:1 2519223
				2024/01/17 23:42:39 txout: 111476 5120a843da35d393bf35929c6a42a0d954e7f9e824166d2bc709ae54ef3a53fcd35b
				2024/01/17 23:42:39 txout: 1549461 5120e1dd03f0270a3a76559dae82aac149b2fa7f13f7d3bf120780eaf8bb6ac6a950
				2024/01/17 23:42:39 locktime: 0
				2024/01/17 23:42:39 hash: 2644bf5d8c1fc9e6546a6285106f346cbefd3c1057e6c1718dff35db5cf138ad

				2024/01/17 23:39:07 version: 1
		2024/01/17 23:39:07 txin: aabbccc8daf7540db4841104e5ccd7b899691f3d6a166e5b00e1a66da0087ad5:1 2519223

		2024/01/17 23:39:07 txout: 111476 512051067e0bc5a220a495403def5e98b06a58e185fa51982f107e4d2f7196702be6
		2024/01/17 23:39:07 txout: 1549461 5120e1dd03f0270a3a76559dae82aac149b2fa7f13f7d3bf120780eaf8bb6ac6a950
		2024/01/17 23:39:07 locktime: 0
		2024/01/17 23:39:07 hash: 0000003e6360d990c31bfe15693ce36db55043a73cc6069a15eb914a3fefad3d
	*/
	atomicals.DecAndPrintTx("010000000001010f8e35c03cdfcb47f89f39c354eb77e464e4f1934ae2392cf9954c1afc0000000000000000ffffffff02a086010000000000225120017b45db1829fdbb4e6d151281c245eaf24e3eba66b7d91a62706c67114b3fa60000000000000000156a13313730353735373031393a303031363631393103404115feca81c150ef80e0a9c2e718b7c6e5240ceadb2ba3e7a35c6255bf9468857af1d441dad1968c6a6cc2388826fd67cfa9ec4108851902445c57de0b7072a97d20c85213ca3db22c66947ad20a7a89ba9c4b89917f76a309bee98dd77dcaad3957ac00630461746f6d03646d744c4da16461726773a56474696d651a65abc951656e6f6e63651a0001000068626974776f726b636630303030303068626974776f726b7264363233386b6d696e745f7469636b657266736f70686f6e6821c1c85213ca3db22c66947ad20a7a89ba9c4b89917f76a309bee98dd77dcaad395700000000")

	var msgTx wire.MsgTx
	data, _ := hex.DecodeString("010000000001015e6e17f985d76d36bad983d95931417a5fd217a472972f8eb0e8908447ccbbaa0100000000051c1d0002d5a4000000000000225120ef239ee9203123505c8051fe14ce53c3245d459b393b398f9449b7c6c0e7c7639b32ab00000000002251200e8162c8e0d8e413c2ab9d2bd1e16857804c9a8dcbd432b1c479f3c7e16925c901401ef516cfcb873fafb7082d126a2968180dbea3363d29119199a28397ee90b9f9d9d237aa25e0a541d2fe91a2f143f26e1c01b6a27f6f3b2053fb622ca8e64ec300000000")
	if err := msgTx.Deserialize(bytes.NewReader(data)); err != nil {
		t.Fatalf("msgTx.Deserialize failed: %v", err)
	}
	atomicals.PrintMsgTx(&msgTx)

	w := bytes.NewBuffer(nil)

	if err := msgTx.SerializeNoWitness(w); err != nil {
		t.Fatalf("msgTx.SerializeNoWitness failed: %v", err)
	}
	cpy := w.Bytes()

	maxSeq := uint32(5)
	for i := uint32(0); i != maxSeq; i++ {
		binary.LittleEndian.PutUint32(cpy[42:], msgTx.TxIn[0].Sequence)
		w := bytes.NewBuffer(nil)
		msgTx.SerializeNoWitness(w)
		if !bytes.Equal(w.Bytes(), cpy) {
			t.Logf("%x", w.Bytes())
			t.Logf("%x", cpy)
			t.Fatal("bytes not equal")
		}
		hash := msgTx.TxHash()
		hash2 := chainhash.DoubleHashB(cpy)
		if !bytes.Equal(hash2, hash[:]) {
			t.Logf("%x", hash2)
			t.Logf("%x", hash.CloneBytes())
			t.Fatal("hash not equal")
		}

		msgTx.TxIn[0].Sequence++
	}
}
