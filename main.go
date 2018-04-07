package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/davecgh/go-spew/spew"

	"github.com/gorilla/mux"
)

// lets make objects

// Block , contains data thats gonna
// get written to the blockchain
type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

// Blockchain , Blockchain ay Lets make a blockchain
var Blockchain []Block

// calculateHash takes a block of data
// and spits out its hash
func calculateHash(block Block) string {
	// a record is a concatted string of the block's info (dna)
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	// then hash that muh' and turn it back to a string. tada
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// generateBlock creates a new Block
// (once we've calculated the hash of course)
func generateBlock(oldBlock Block, BPM int) (Block, error) {
	// create that new block, take note of the time too.
	var newBlock Block
	t := time.Now()

	// aight set that new block up sum'n proper
	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	// calculate that yung hash now
	newBlock.Hash = calculateHash(newBlock)

	// send it up
	return newBlock, nil

}

// isBlockValid checks a block to see if
// its been tampered wit
func isBlockValid(newBlock, oldBlock Block) bool {
	// if the index is off, you iffy
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}
	// if the hash is off, you suspect
	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}
	// if the new hash dont equal up you fradulent
	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	// otherwise you goodie
	return true
}

// updateChain updates our Blockchain if we find
// one with more blocks
func updateChain(potentialBlocks []Block) {
	if len(potentialBlocks) > len(Blockchain) {
		Blockchain = potentialBlocks
	}
}

func handleGetBlockhain(w http.ResponseWriter, r *http.Request) {
	// turn the Blockchain to json and send it back
	bytes, err := json.MarshalIndent(Blockchain, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

// Message is for out POST message
type Message struct {
	BPM int
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message

	decoder := json.NewDecoder(r.Body)
	// if you didnt send us json, we cant do shit 4 u
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	// either way we gotta close this later, so set it now
	defer r.Body.Close()

	// create the new block from the most recent block and our piece of data.
	// in this someones heart BPMs
	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}
	// then check if the new block is valid
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		updateChain(newBlockchain)
		spew.Dump(Blockchain)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)

}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

// run() , runs a server to view the blocks
func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listenin on ", httpAddr)

	// serve
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

// makeMuxRouter(), are functions to handle r/w interfacing
// with the chain itself
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockhain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// run the whole the whole thing async & start the server
	go func() {
		t := time.Now()
		// Create the genesis
		genesisBlock := Block{0, t.String(), 0, "", ""}
		// pprint it
		spew.Dump(genesisBlock)
		// add it as the first block
		Blockchain = append(Blockchain, genesisBlock)
	}()

	log.Fatal(run())
}
