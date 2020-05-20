package blockchain

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

//Block ...
type Block struct {
	Index        int
	Timestamp    int64
	Transactions []Transaction
	Proof        int
	PreviousHash string
}

//Transaction ...
type Transaction struct {
	Sender    string
	Recipient string
	Amount    int
}

//BlockChain ...
type BlockChain struct {
	Transactions []Transaction
	Chain        []Block
	Nodes        map[string]bool
}

// NewBlock ....
func (b *BlockChain) NewBlock(proof int, previousHash string) *Block {

	hash := previousHash

	if len(previousHash) <= 0 {
		hash = b.HashBlock(b.Chain[len(b.Chain)-1])
	}

	block := &Block{
		Index:        len(b.Chain),
		Timestamp:    time.Now().Unix(),
		Transactions: b.Transactions,
		Proof:        proof,
		PreviousHash: hash,
	}

	b.Transactions = nil

	b.Chain = append(b.Chain, *block)

	return block
}

// NewTransaction ...
func (b *BlockChain) NewTransaction(sender string, recipient string, amount int) int {
	transaction := &Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	}

	b.Transactions = append(b.Transactions, *transaction)

	return b.Chain[len(b.Chain)-1].Index
}

// ValidateProof checks if the proof matches the algorithm
func (b *BlockChain) validateProof(lastProof int, newProof int) bool {
	proof := fmt.Sprintf("%d%d", lastProof, newProof)

	hasher := sha256.New()
	hasher.Write([]byte(proof))

	// get the last 4 characters
	guess := base64.URLEncoding.EncodeToString(hasher.Sum(nil))[:4]

	// check if the end in 0000
	return guess == "0000"
}

func (b *BlockChain) ProofOfWork(lastProof int) int {
	newProof := 1
	for {
		if b.validateProof(lastProof, newProof) {
			return newProof
		}
		newProof++
	}
}

// HashBlock return a base64 encoded of sha256 hash
func (b *BlockChain) HashBlock(block Block) string {
	var marshalledBlock, err = json.Marshal(block)
	if err != nil {
		panic(err)
	}

	hasher := sha256.New()

	hasher.Write([]byte(marshalledBlock))

	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func (b *BlockChain) CreateFirstBlock() {
	b.NewBlock(100, "1")
}

func (b *BlockChain) RegisterNode(adress string) {
	b.Nodes[adress] = true
}

func (b *BlockChain) RegisterNodes(adresses []string) {
	for _, address := range adresses {
		b.RegisterNode(address)
	}
}

func (b *BlockChain) ResolveConflicts() bool {
	neighbours := b.Nodes
	var newChain []Block
	maxLength := len(b.Chain)

	for adress := range neighbours {
		nodeChain := GetNodeChain(adress)

		if len(nodeChain) > maxLength && validateChain(nodeChain) {
			newChain = nodeChain
			maxLength = len(nodeChain)
		}
	}

	if newChain != nil {
		b.Chain = newChain
		return true
	}

	return false
}

func validateChain(chain []Block) bool {
	return true
}

func GetNodeChain(address string) []Block {
	resp, _ := http.Get(address)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	chain := []Block{}
	json.Unmarshal(body, chain)
	return chain
}
