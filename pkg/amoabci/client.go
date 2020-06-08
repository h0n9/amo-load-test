package amoabci

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"strings"

	"github.com/interchainio/tm-load-test/pkg/loadtest"

	"github.com/amolabs/amo-client-go/lib/keys"
	"github.com/amolabs/amo-client-go/lib/rpc"
)

type AMOABCIClientFactory struct {
	Clients map[string]*AMOABCIClient
}

var _ loadtest.ClientFactory = (*AMOABCIClientFactory)(nil)

func NewAMOABCIClientFactory() *AMOABCIClientFactory {
	return &AMOABCIClientFactory{
		Clients: map[string]*AMOABCIClient{},
	}
}

func (f *AMOABCIClientFactory) ValidateConfig(cfg loadtest.Config) error {
	return nil
}

type AMOABCIClient struct {
	Nickname string
	Key      keys.KeyEntry

	Peers *map[string]*AMOABCIClient

	Endpoint string
}

var _ loadtest.Client = (*AMOABCIClient)(nil)

func (f *AMOABCIClientFactory) NewClient(cfg loadtest.Config) (loadtest.Client, error) {
	nickname := strconv.FormatInt(rand.Int63(), 10)
	keyEntry, err := keys.GenerateKey(nickname, nil, false)
	if err != nil {
		return nil, err
	}

	client := AMOABCIClient{
		Nickname: nickname,
		Key:      *keyEntry,
		Peers:    &f.Clients,
		Endpoint: cfg.Endpoints[0],
	}

	if _, exists := f.Clients[nickname]; exists {
		return nil, fmt.Errorf("%s already exists in clients", nickname)
	}

	fmt.Println("feed:", keyEntry.Address)

	f.Clients[nickname] = &client

	return &client, nil
}

func (c *AMOABCIClient) getLastHeight() (string, error) {
	lastHeight := ""

	rpcRemote := "http://" + strings.Split(c.Endpoint, "/")[2]
	rpc.RpcRemote = rpcRemote
	rawMsg, err := rpc.NodeStatus()
	if err != nil {
		return lastHeight, err
	}

	jsonMsg, err := json.Marshal(rawMsg.SyncInfo)
	if err != nil {
		return lastHeight, err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(jsonMsg, &data)
	if err != nil {
		return lastHeight, err
	}

	lastHeight = data["latest_block_height"].(string)

	return lastHeight, nil
}

func (c *AMOABCIClient) getRandPeer() string {
	target := "E580331EE30FFAAE2E6911D7B6F669444FC26520"

	ks := []string{}
	for k := range *c.Peers {
		keyEntry, err := keys.GenerateKey(k, nil, false)
		if err != nil {
			continue
		}
		if keyEntry.Address == c.Key.Address {
			continue
		}
		ks = append(ks, k)
	}

	if len(ks) > 0 {
		nickname := ks[rand.Intn(len(ks))]
		peers := *c.Peers
		target = peers[nickname].Key.Address
	}

	return target
}

func (c *AMOABCIClient) GenerateTx() ([]byte, error) {
	fee := "0"
	lastHeight, err := c.getLastHeight()
	if err != nil {
		return nil, err
	}

	tmp := rand.Int63()
	randNum, err := crand.Int(crand.Reader, big.NewInt(tmp))
	if err != nil {
		return nil, err
	}
	randByte := make([]byte, 4)
	binary.BigEndian.PutUint32(randByte, uint32(randNum.Int64()))

	target := fmt.Sprintf("00000001%x", randByte)
	custody := "11ffeeff"

	signedTx, err := SignTx("register", struct {
		Target  string `json:"target"`
		Custody string `json:"custody"`
	}{target, custody}, c.Key, fee, lastHeight)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func (c *AMOABCIClient) GenerateTxTransfer() ([]byte, error) {
	to := c.getRandPeer()
	amount := "1000"
	fee := "0"
	lastHeight, err := c.getLastHeight()
	if err != nil {
		return nil, err
	}

	signedTx, err := SignTx("transfer", struct {
		To     string `json:"to"`
		Amount string `json:"amount"`
	}{to, amount}, c.Key, fee, lastHeight)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}
