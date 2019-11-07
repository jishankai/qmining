package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"strconv"
	"io/ioutil"

	"github.com/gorilla/mux"

	"github.com/sammy007/open-ethereum-pool/storage"
	"github.com/sammy007/open-ethereum-pool/util"
)

type ApiConfig struct {
	Enabled              bool   `json:"enabled"`
	Listen               string `json:"listen"`
	StatsCollectInterval string `json:"statsCollectInterval"`
	HashrateWindow       string `json:"hashrateWindow"`
	HashrateLargeWindow  string `json:"hashrateLargeWindow"`
	LuckWindow           []int  `json:"luckWindow"`
	Payments             int64  `json:"payments"`
	Blocks               int64  `json:"blocks"`
	PurgeOnly            bool   `json:"purgeOnly"`
	PurgeInterval        string `json:"purgeInterval"`
}

type ApiServer struct {
	config              *ApiConfig
	backend             *storage.RedisClient
	hashrateWindow      time.Duration
	hashrateLargeWindow time.Duration
	stats               atomic.Value
	miners              map[string]*Entry
	minersMu            sync.RWMutex
	statsIntv           time.Duration
}

type Entry struct {
	stats     map[string]interface{}
	updatedAt int64
}

func NewApiServer(cfg *ApiConfig, backend *storage.RedisClient) *ApiServer {
	hashrateWindow := util.MustParseDuration(cfg.HashrateWindow)
	hashrateLargeWindow := util.MustParseDuration(cfg.HashrateLargeWindow)
	return &ApiServer{
		config:              cfg,
		backend:             backend,
		hashrateWindow:      hashrateWindow,
		hashrateLargeWindow: hashrateLargeWindow,
		miners:              make(map[string]*Entry),
	}
}

func (s *ApiServer) Start() {
	if s.config.PurgeOnly {
		log.Printf("Starting API in purge-only mode")
	} else {
		log.Printf("Starting API on %v", s.config.Listen)
	}

	s.statsIntv = util.MustParseDuration(s.config.StatsCollectInterval)
	statsTimer := time.NewTimer(s.statsIntv)
	log.Printf("Set stats collect interval to %v", s.statsIntv)

	purgeIntv := util.MustParseDuration(s.config.PurgeInterval)
	purgeTimer := time.NewTimer(purgeIntv)
	log.Printf("Set purge interval to %v", purgeIntv)

	sort.Ints(s.config.LuckWindow)

	if s.config.PurgeOnly {
		s.purgeStale()
	} else {
		s.purgeStale()
		s.collectStats()
	}

	go func() {
		for {
			select {
			case <-statsTimer.C:
				if !s.config.PurgeOnly {
					s.collectStats()
				}
				statsTimer.Reset(s.statsIntv)
			case <-purgeTimer.C:
				s.purgeStale()
				purgeTimer.Reset(purgeIntv)
			}
		}
	}()

	if !s.config.PurgeOnly {
		s.listen()
	}
}

func (s *ApiServer) listen() {
	r := mux.NewRouter()
	r.HandleFunc("/api/stats", s.StatsIndex)
	r.HandleFunc("/api/miners", s.MinersIndex)
	r.HandleFunc("/api/blocks", s.BlocksIndex)
	r.HandleFunc("/api/payments", s.PaymentsIndex)
	r.HandleFunc("/api/accounts", s.AccountIndex)
	r.HandleFunc("/api/QkcPrices", s.GetCoinQKC)
	r.NotFoundHandler = http.HandlerFunc(notFound)
	err := http.ListenAndServe(s.config.Listen, r)
	if err != nil {
		log.Fatalf("Failed to start API: %v", err)
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusNotFound)
}

func (s *ApiServer) GetCoinQKC(w http.ResponseWriter, r *http.Request) {
   w.Header().Set("Content-Type", "application/json; charset=UTF-8")
   w.Header().Set("Access-Control-Allow-Origin", "*")
   w.Header().Set("Cache-Control", "no-cache")
   w.WriteHeader(http.StatusOK)
   resp, err := http.Get("https://api.coingecko.com/api/v3/coins/quark-chain")
   if err != nil {
      log.Println( err)
   }
   defer resp.Body.Close()
   body, err := ioutil.ReadAll(resp.Body)
   if err != nil {
      // handle error
      w.Write([]byte(err.Error()))
   }
   reply := make(map[string]interface{})
   m := make(map[string]interface{})
   err = json.Unmarshal(body, &m)
   if err != nil {
	   log.Println("Error Unmarshal json: ", err)
   }
   reply["market_data"] = (m["market_data"].(map[string]interface{}))["current_price"]
   err = json.NewEncoder(w).Encode(reply)
   if err != nil {
	   log.Println("Error serializing API response: ", err)
   }
}

func (s *ApiServer) purgeStale() {
	start := time.Now()
	total, err := s.backend.FlushStaleStats(s.hashrateWindow, s.hashrateLargeWindow)
	if err != nil {
		log.Println("Failed to purge stale data from backend:", err)
	} else {
		log.Printf("Purged stale stats from backend, %v shares affected, elapsed time %v", total, time.Since(start))
	}
}

func (s *ApiServer) collectStats() {
	start := time.Now()
	stats, err := s.backend.CollectStats(s.hashrateWindow, s.config.Blocks, s.config.Payments)
	if err != nil {
		log.Printf("Failed to fetch stats from backend: %v", err)
		return
	}
	if len(s.config.LuckWindow) > 0 {
		stats["luck"], err = s.backend.CollectLuckStats(s.config.LuckWindow)
		if err != nil {
			log.Printf("Failed to fetch luck stats from backend: %v", err)
			return
		}
	}
	s.stats.Store(stats)
	log.Printf("Stats collection finished %s", time.Since(start))
}

func (s *ApiServer) StatsIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	reply := make(map[string]interface{})
	nodes, err := s.backend.GetNodeStates()
	if err != nil {
		log.Printf("Failed to get nodes stats from backend: %v", err)
	}
	reply["nodes"] = nodes

	stats := s.getStats()
	if stats != nil {
		reply["now"] = util.MakeTimestamp()
		reply["stats"] = stats["stats"]
		reply["hashrate"] = stats["hashrate"]
		reply["minersTotal"] = stats["minersTotal"]
		reply["maturedTotal"] = stats["maturedTotal"]
		reply["immatureTotal"] = stats["immatureTotal"]
		reply["candidatesTotal"] = stats["candidatesTotal"]
		reply["hashrateList"] = stats["hashrateList"]
	}

	err = json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}

func (s *ApiServer) MinersIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	reply := make(map[string]interface{})
	stats := s.getStats()
	if stats != nil {
		reply["now"] = util.MakeTimestamp()
		reply["miners"] = stats["miners"]
		reply["hashrate"] = stats["hashrate"]
		reply["minersTotal"] = stats["minersTotal"]
	}

	err := json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}

func (s *ApiServer) BlocksIndex(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	page, err_int := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)

	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)

	if err_int != nil {
		log.Println("Block Error serializing API page: ", err_int)
	}

	pageSize := limit
	reply := make(map[string]interface{})
	reply["pageSize"] = pageSize
	reply["page"] = page
	stats := s.getStats()
	if stats != nil {
		lowerBound := pageSize * (page - 1)
		upperBound := pageSize * page
		totalInt := int64(len(stats["matured"].([]*storage.BlockData)[:]))
		if upperBound > totalInt {
				upperBound = totalInt
		}
		reply["data"] = stats["matured"].([]*storage.BlockData)[lowerBound:upperBound]
		reply["limit"] = int64(len(stats["matured"].([]*storage.BlockData)[lowerBound:upperBound])) 
		reply["numberPages"] = (totalInt + pageSize - 1) / pageSize 
		reply["count"] = totalInt
	}
	reply["code"] = 0
	reply["msg"] = ""

	err := json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}

func (s *ApiServer) PaymentsIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	reply := make(map[string]interface{})
	stats := s.getStats()
	reply["payments"] = []string{}
	if stats != nil {
		reply["payments"] = stats["payments"]
		reply["paymentsTotal"] = stats["paymentsTotal"]
	}

	err := json.NewEncoder(w).Encode(reply)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}

func (s *ApiServer) AccountIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")

	//login := strings.ToLower(mux.Vars(r)["login"])

	login_query := r.URL.Query().Get("login")
	login := strings.ToLower(login_query)
	if len(login) != 42 {
		log.Println("Url Param 'login' is missing")
		return
	}

        page, err_int := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)

        limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)

        if err_int != nil || limit <= 0 {
		return
		log.Println("limit", limit)
                log.Println("Account Error serializing API page: ", err_int)
        }

	s.minersMu.Lock()
	defer s.minersMu.Unlock()

	reply, ok := s.miners[login]
	now := util.MakeTimestamp()
	cacheIntv := int64(s.statsIntv / time.Millisecond)
	// Refresh stats if stale
	if !ok || reply.updatedAt < now-cacheIntv {
		exist, err := s.backend.IsMinerExists(login)
		if !exist {
			w.WriteHeader(http.StatusOK)
			okb := make(map[string]interface{})
			okb["code"] = 0
			okb["msg"] = "404"
			okb["count"] = 0
			okb["data"] = make([]string, 0)
			err := json.NewEncoder(w).Encode(okb)
			if err != nil {
				log.Println("Error serializing API response: ", err)
				log.Println("Url param 'login' is missing")
			}
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Failed to fetch stats from backend: %v", err)
			return
		}

		stats, err := s.backend.GetMinerStats(login, s.config.Payments)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Failed to fetch stats from backend: %v", err)
			return
		}
		workers, err := s.backend.CollectWorkersStats(s.hashrateWindow, s.hashrateLargeWindow, login)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Failed to fetch stats from backend: %v", err)
			return
		}
		for key, value := range workers {
			stats[key] = value
		}
		stats["pageSize"] = s.config.Payments
		lowerBound := limit * (page - 1)
		upperBound := limit * page
		totalPayments := stats["paymentsTotal"].(int64)
		if upperBound > totalPayments {
			upperBound = totalPayments
		}
		stats["payments"] = stats["payments"].([]map[string]interface{})[lowerBound:upperBound]
		stats["code"] = 0
		stats["msg"] = ""
		stats["count"] = totalPayments
		reply = &Entry{stats: stats, updatedAt: now}
		s.miners[login] = reply
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(reply.stats)
	if err != nil {
		log.Println("Error serializing API response: ", err)
	}
}

func (s *ApiServer) getStats() map[string]interface{} {
	stats := s.stats.Load()
	if stats != nil {
		return stats.(map[string]interface{})
	}
	return nil
}
