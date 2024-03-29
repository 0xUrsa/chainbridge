package substrate

import (
	// "encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stafiprotocol/chainbridge/config"
	"github.com/stafiprotocol/chainbridge/utils/msg"
	// "github.com/stafiprotocol/go-substrate-rpc-client/types"
	// "github.com/stretchr/testify/assert"
)

var (
	rId      = msg.ResourceIdFromSlice(hexutil.MustDecode("0x000000000000000000000000000000a9e0095b8965c01e6a09c97938f3860901"))
	decimals = map[string]*big.Int{
		"Default": big.NewInt(1000000),
	}
)

// func TestWriter_resolveResourceId(t *testing.T) {
// 	stop := make(chan int)
// 	defer close(stop)
// 	errs := make(chan error)
// 	conn, err := NewConnection(localCfg, AliceTestLogger, make(chan int))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer conn.Close()

// 	w := NewWriter(conn, AliceTestLogger, errs, stop, decimals)
// 	re, err := w.resolveResourceId(rId)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	fmt.Println(re)
// }

func TestStr(t *testing.T) {
	fmt.Println(hexutil.Encode([]byte(config.BridgeSwap + ".transfer_native_back")))
	//0x427269646765537761702e7472616e736665725f6e61746976655f6261636b
	fmt.Println(hexutil.Encode([]byte(config.BridgeSwap + ".transfer_rtoken_back")))
	//0x427269646765537761702e7472616e736665725f72746f6b656e5f6261636b
	fmt.Println(hexutil.Encode([]byte(config.BridgeSwap + ".transfer_xtoken_back")))
	//0x427269646765537761702e7472616e736665725f78746f6b656e5f6261636b
}

// func TestNewProp(t *testing.T) {
// 	src := "0x7b226465706f7369744e6f6e6365223a332c2263616c6c223a7b2243616c6c496e646578223a7b2253656374696f6e496e646578223a33312c224d6574686f64496e646578223a317d2c2241726773223a224d7a59304e445933556a4e55626d394e59324a4d4d6b46316145706f59575633635842426332355854553441414f694a42435048696741414141414141414141227d2c22736f757263654964223a322c227265736f757263654964223a5b302c302c302c302c302c302c302c302c302c302c302c302c302c302c302c3136392c3232342c392c39312c3133372c3130312c3139322c33302c3130362c392c3230312c3132312c35362c3234332c3133342c392c315d2c226d6574686f64223a22427269646765537761702e7472616e736665725f6e61746976655f6261636b227d"
// 	b := hexutil.MustDecode(src)
// 	prop := new(proposal)
// 	json.Unmarshal(b, prop)
// 	fmt.Println(prop)

// 	var voteRes voteState
// 	srcId, err := types.EncodeToBytes(prop.SourceId)
// 	assert.NoError(t, err)
// 	fmt.Println("srcId:", hexutil.Encode(srcId))

// 	propBz, err := prop.encode()
// 	assert.NoError(t, err)

// 	errs := make(chan error)
// 	conn, err := NewConnection(localCfg, AliceTestLogger, make(chan int))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer conn.Close()
// 	stop := make(chan int)
// 	defer close(stop)
// 	w := NewWriter(conn, AliceTestLogger, errs, stop, decimals)

// 	exists, err := w.conn.QueryStorage(config.BridgeCommon, "Votes", srcId, propBz, &voteRes)
// 	assert.NoError(t, err)

// 	fmt.Println("exists:", exists)

// 	fmt.Println(voteRes)

// 	acId := types.NewAccountID(w.conn.key.PublicKey)
// 	fmt.Println(hexutil.Encode(acId[:]))
// 	cmp := containsVote(voteRes.Voted, acId)
// 	fmt.Println(cmp)
// }

func TestEncode(t *testing.T) {
	//a := "msg"
	b := []byte("sdfsdf")
	fmt.Println(hexutil.Encode(b))
}
