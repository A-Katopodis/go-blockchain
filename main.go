package main

import (
	"fmt"
	"net/http"

	b "app/blockchain"
	"encoding/json"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"

	"github.com/go-redis/redis/v7"
)

func main() {
	StartBlockChain()
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/chain", chain)
	r.Get("/mine", mine)
	r.Post("/transaction", Transaction)
	r.Post("/nodes", RegisterNodes)
	r.Post("/consencus", Consencus)
	http.ListenAndServe(":3000", r)
}

func RegisterNodes(w http.ResponseWriter, r *http.Request) {
	nodes := &[]string{}
	json.NewDecoder(r.Body).Decode(nodes)

	blockchain := GetBlockchain()
	blockchain.RegisterNodes(*nodes)

}

func Consencus(w http.ResponseWriter, r *http.Request) {
	GetBlockchain().ResolveConflicts()
}

func StartBlockChain() {
	blockChain := b.BlockChain{}
	blockChain.Chain = []b.Block{}
	blockChain.Transactions = []b.Transaction{}
	blockChain.Nodes = make(map[string]bool)
	blockChain.CreateFirstBlock()
	redis := ConnectToRedis()

	value, _ := json.Marshal(blockChain)

	err := redis.SetNX("blockchain", value, 0).Err()
	if err != nil {
		fmt.Printf(err.Error())
	}

	redis.Set("ID", uuid.New().String(), 0)
}

func Transaction(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	transaction := &tranasctionPostRequest{}

	json.NewDecoder(r.Body).Decode(&transaction)

	blockChain := GetBlockchain()

	index := blockChain.NewTransaction(transaction.Sender, transaction.Recipient, transaction.Amount)

	content, _ := json.Marshal(&NewTransactionResponse{Index: index})
	SaveBlockChain(*blockChain)
	w.Write(content)
}

func mine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	blockchain := GetBlockchain()
	previousBlock := blockchain.Chain[len(blockchain.Chain)-1]
	lastProof := previousBlock.Proof
	newProof := blockchain.ProofOfWork(lastProof)

	nodeID := GetNodeID()

	blockchain.NewTransaction("0", nodeID, 1)

	previousHash := blockchain.HashBlock(previousBlock)

	block := blockchain.NewBlock(newProof, previousHash)

	content, _ := json.Marshal(block)
	SaveBlockChain(*blockchain)
	w.Write(content)
}

func chain(w http.ResponseWriter, r *http.Request) {

	blockChain := GetBlockchain()

	var content, err = json.Marshal(blockChain.Chain)

	if err != nil {
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write(content)
}

type tranasctionPostRequest struct {
	Sender    string
	Recipient string
	Amount    int
}

type NewTransactionResponse struct {
	Index int
}

func ConnectToRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func GetBlockchain() *b.BlockChain {
	redis := ConnectToRedis()
	val := redis.Get("blockchain").Val()
	blockchain := &b.BlockChain{}

	json.Unmarshal([]byte(val), blockchain)

	return blockchain
}

func GetNodeID() string {
	redis := ConnectToRedis()

	return redis.Get("ID").Val()
}

func SaveBlockChain(b b.BlockChain) {
	redis := ConnectToRedis()
	value, _ := json.Marshal(b)

	redis.Set("blockchain", value, 0)
}
