package proxy

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"hash"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sammy007/open-ethereum-pool/rpc"
	//"github.com/sammy007/open-ethereum-pool/util"
	"golang.org/x/crypto/sha3"
)

type hasher func(dest []byte, data []byte)

const (
	epochLength = 30000 // Blocks per epoch
)

var pow256 = math.BigPow(2, 256)

const maxBacklog = 3

type heightDiffPair struct {
	diff   *big.Int
	height uint64
}

type BlockTemplate struct {
	sync.RWMutex
	Header               string
	Seed                 string
	Target               string
	Difficulty           *big.Int
	Height               uint64
	GetPendingBlockCache *rpc.GetBlockReplyPart
	nonces               map[string]bool
	headers              map[string]heightDiffPair
}

type Block struct {
	difficulty  *big.Int
	hashNoNonce common.Hash
	nonce       uint64
	mixDigest   common.Hash
	number      uint64
}

func (b Block) Difficulty() *big.Int     { return b.difficulty }
func (b Block) HashNoNonce() common.Hash { return b.hashNoNonce }
func (b Block) Nonce() uint64            { return b.nonce }
func (b Block) MixDigest() common.Hash   { return b.mixDigest }
func (b Block) NumberU64() uint64        { return b.number }

func (s *ProxyServer) fetchBlockTemplate() {
	rpc := s.rpc()
	count := 0
	//GetWork for all miners seperately
	for m, _ := range s.sessions {
		go func(cs *Session) {
			s.sessionsMu.Lock()
			s.updateMap[cs.login] = false
			s.sessionsMu.Unlock()
			reply, err := rpc.GetWorkWithID(s.config.Proxy.Stratum.ShardId, cs.login)
			if err != nil {
				log.Printf("Error while refreshing block template on %s: %s", rpc.Name, err)
				return
			}
			// No need to update, we have fresh job
			t := s.currentBlockTemplateWithId(cs.login)
			if t == nil {
				var inital_atomic atomic.Value
				s.minerBlockTemplateMap[cs.login] = inital_atomic
			}
			if t != nil && t.Header == reply[0] {
				return
			}
			diff_template_seperate := DiffHexToDiff(reply[2])
			height_temp_seperate := HexToInt64(reply[1])
			seed_seperate := seedHash(height_temp_seperate)
			guardian_diff_seperate := diff_template_seperate
			if len(reply) == 4 {
				guardian_diff_seperate = new(big.Int).Div(diff_template_seperate, new(big.Int).SetInt64(10000))
			}
			// Seed equals to hex string Height
			nTemplate := BlockTemplate{
				Header:     reply[0],
				Seed:       fmt.Sprintf("0x%x", seed_seperate),
				Target:     GetTargetHexFromDiff(guardian_diff_seperate),
				Height:     height_temp_seperate,
				Difficulty: guardian_diff_seperate,
				//Difficulty:           big.NewInt(diff),
				GetPendingBlockCache: nil,
				headers:              make(map[string]heightDiffPair),
			}
			// Copy job backlog and add current one
			nTemplate.headers[reply[0]] = heightDiffPair{
				diff:   guardian_diff_seperate,
				height: HexToInt64(reply[1]),
			}
			if t != nil {
				for k, v := range t.headers {
					//if v.height > height-maxBacklog {
					nTemplate.headers[k] = v
					//}
				}
			}
			s.sessionsMu.Lock()
			atomic_temp := s.minerBlockTemplateMap[cs.login]
			atomic_temp.Store(&nTemplate)
			s.minerBlockTemplateMap[cs.login] = atomic_temp
			s.updateMap[cs.login] = true
			s.sessionsMu.Unlock()
			count++
			if t != nil && t.Height > s.Height {
				s.Height = t.Height
				s.Difficulty = t.Difficulty
			}
			if s.config.Proxy.Stratum.Enabled {
				go s.broadcastNewJobs()
			}

		}(m)
	}
}

func (s *ProxyServer) fetchPendingBlock() (*rpc.GetBlockReplyPart, uint64, int64, error) {
	rpc := s.rpc()
	reply, err := rpc.GetPendingBlock(s.config.Proxy.Stratum.ShardId)
	if err != nil {
		log.Printf("Error while refreshing pending block on %s: %s", rpc.Name, err)
		return nil, 0, 0, err
	}
	blockNumber, err := strconv.ParseUint(strings.Replace(reply.Number, "0x", "", -1), 16, 64)
	if err != nil {
		log.Println("Can't parse pending block number")
		return nil, 0, 0, err
	}
	blockDiff, err := strconv.ParseInt(strings.Replace(reply.Difficulty, "0x", "", -1), 16, 64)
	if err != nil {
		log.Println("Can't parse pending block difficulty")
		return nil, 0, 0, err
	}
	return reply, blockNumber, blockDiff, nil
}

// makeHasher creates a repetitive hasher, allowing the same hash data structures to
// be reused between hash runs instead of requiring new ones to be created. The returned
// function is not thread safe!
func makeHasher(h hash.Hash) hasher {
	// sha3.state supports Read to get the sum, use it to avoid the overhead of Sum.
	// Read alters the state but we reset the hash before every operation.
	type readerHash interface {
		hash.Hash
		Read([]byte) (int, error)
	}
	rh, ok := h.(readerHash)
	if !ok {
		panic("can't find Read method on hash")
	}
	outputLen := rh.Size()
	return func(dest []byte, data []byte) {
		rh.Reset()
		rh.Write(data)
		rh.Read(dest[:outputLen])
	}
}

// seedHash is the seed to use for generating a verification cache and the mining
// dataset.
func seedHash(block uint64) []byte {
	seed := make([]byte, 32)
	if block < epochLength {
		return seed
	}
	keccak256 := makeHasher(sha3.NewLegacyKeccak256())
	for i := 0; i < int(block/epochLength); i++ {
		keccak256(seed, seed)
	}
	return seed
}

func HexToInt64(height string) uint64 {
	value, _ := strconv.ParseUint(strings.Replace(height, "0x", "", -1), 16, 64)
	return value
}

// QuarkChain new function to adjust RPC
func DiffHexToDiff(diffHex string) *big.Int {
	diffBytes := common.FromHex(diffHex)
	return new(big.Int).SetBytes(diffBytes)
}

func GetTargetHexFromDiff(diff *big.Int) string {
	diff1 := new(big.Int).Div(pow256, diff)
	return string(common.ToHex(diff1.Bytes()))
}
