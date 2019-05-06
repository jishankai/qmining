## Ethash Mining Pool for QuarkChain


### Features

**This pool is being further developed to provide an easy to use pool for QuarkChain Ethash miners. This software is functional however an optimised release of the pool is expected soon. Testing and bug submissions are welcome!**

* Support for HTTP and Stratum mining
* Support failover pool
* Separate stats for workers: can highlight timed-out workers so miners can perform maintenance of rigs
* JSON-API for stats
* Support [Ethminer mining](https://github.com/ethereum-mining/ethminer)
* Support [Claymore mining](https://github.com/nanopool/Claymore-Dual-Miner/releases)
* Support NiceHash Stratum mode

#### Proxies

* [Ether-Proxy](https://github.com/sammy007/ether-proxy) HTTP proxy with web interface
* [Stratum Proxy](https://github.com/Atrides/eth-proxy) for Ethereum

## Building on Linux

Dependencies:
  * VPS with at least 2G of Ram
  * Disk space > 50G
  * go >= 1.9
  * pyquarkchain
  * redis-server >= 2.8.0
  

**I highly recommend to use Ubuntu 18.04 LTS.**

## AWS AMI (Private)
QuarkChain Ethash Stratum Pool (ami-02a5fe1b1c1a9d211) -- Oregon
  
### Update

    $ sudo apt-get update
    $ sudo apt-get dist-upgrade
    $ sudo apt-get install build-essential make
    $ sudo reboot

## Install GoLang

    $ wget https://dl.google.com/go/go1.10.2.linux-amd64.tar.gz
    $ sudo tar -xvf go1.10.2.linux-amd64.tar.gz
    $ sudo mv go /usr/local
    $ sudo vim ~/.profile
    
    
~/.profile

    ...
    # set PATH so it includes user's private bin directories
    PATH="$HOME/bin:$HOME/.local/bin:$PATH"
    export PATH=$PATH:/usr/local/go/bin
    
    
Check the go version

    $ source ~/.profile
    $ go version
    go version go1.10.2 linux/amd64

## Installing Redis and Test

### Build and Install Redis

    $ sudo apt-get update
    $ sudo apt-get install build-essential tcl
    $ curl -O http://download.redis.io/redis-stable.tar.gz
    $ tar xzvf redis-stable.tar.gz
    $ cd redis-stable
    $ make
    $ make test
    $ sudo make install

### Configure Redis

    $ sudo mkdir /etc/redis
    $ sudo cp ~/redis-stable/redis.conf /etc/redis
    $ sudo vim /etc/redis/redis.conf
    
/etc/redis/redis.conf

    . . .

    # Set supervised to systemd
    supervised systemd

    # Set the dir
    dir /var/lib/redis

    . . .
    
    # Warning: since Redis is pretty fast an outside user can try up to
    # 150k passwords per second against a good box. This means that you should
    # use a very strong password otherwise it will be very easy to break.
    #
    requirepass tothem00n
    
### Create a Redis systemed Unit File

    $ sudo vim /etc/systemd/system/redis.service
    
/etc/systemd/system/redis.service

    [Unit]
     Description=Redis In-Memory Data Store
     After=network.target

     [Service]
     User=redis
     Group=redis
     ExecStart=/usr/local/bin/redis-server /etc/redis/redis.conf
     ExecStop=/usr/local/bin/redis-cli shutdown
     Restart=always

    [Install]
    WantedBy=multi-user.target

### Create the Redis User, Group and Directories

    $ sudo adduser --system --group --no-create-home redis
    $ sudo mkdir /var/lib/redis
    $ sudo chown redis:redis /var/lib/redis
    $ sudo chmod 770 /var/lib/redis
    
### Start and Test Redis

    $ sudo systemctl start redis
    $ sudo systemctl status redis
    
### Test the Redis Instance Functionality

    $ redis-cli
    $ ping
    PONG
    $ exit
    
### Enable Redis to Start at Boot

    $ sudo systemctl enable redis
    


## Run a full QuarkChain cluster

First install  [pyquarkchain](https://github.com/QuarkChain/pyquarkchain.git).

    $ git clone https://github.com/QuarkChain/pyquarkchain.git
    $ cd pyquarkchain
    $ python3 quarkchain/cluster/cluster.py --cluster_config /path/to/cluster_config_template.json

It will start sync process depend on speed of your server it may take time to completely sync. 
You can use the quick sync by following the [instructions](https://github.com/QuarkChain/pyquarkchain/wiki/Run-a-Private-Cluster-on-the-QuarkChain-Testnet-2.0).
## Running Ethash Pool
    $ git clone https://github.com/QuarkChain/mining.git
    $ cd mining
    $ make
    $ git branch Ethash_pool
    # Edit the config.json for different shard setting
    $ ./build/bin/open-ethereum-pool config.example_Ethash.json

You can use Ubuntu upstart - check for sample config in <code>upstart.conf</code>.


### Check the mining state
The payout functions and the web UI do not work currently. You can achieve the mining state by reading from the redis database. 

    python3 redis_check_mining_state.py --port 6379
    
### Ethash and Qkchash pool port configuration

Root Guardian Pool IP is:
  * pool0.quarkchain.io [52.11.34.67] (Oregen)
  * 18.138.144.212 (Singapore)
  * 52.197.219.127 (Tokyo)
  
Root chain coinbase address:
  * 0x7DeB90eF2097D8A9e423516e199b9D95EB2b4D97 (POSW)
  * 0xf923ac88fc61837662bace7e94720c7a071997e6
  * 0x2b7acc42b0dc2a1562601e2ed9957eadff7a134

Guardian Signature Server IP:
  * 34.74.159.114
  * 34.83.17.20
  * 35.187.152.61

Shard chaid Pool IP is:
  * pool1.quarkchain.io [54.203.168.137] (Oregen)

Mainnet Pool AMI:

Mainnet_pool_standard_AMI_05_06 (ami-0d838d064757dd87b) (Oregen, Singapore, and Tokyo)


|Chains |Hash Algorithm |Stratum port| Proxy port | API port(web server) | redis port
| ---           | ---          | ---  | ---  | ---  | --- |
| Root Chain    | Ethash       | 8000 | 8888 | 8079 | 6379 | 
| Shard 0       | Ethash       | 8008 | 8888 | 8080 | 6379 | 
| Shard 1       | Ethash       | 8018 | 8881 | 8081 | 6379 | 
| Shard 2       | Ethash       | 8028 | 8882 | 8082 | 6379 | 
| Shard 3       | Ethash       | 8038 | 8883 | 8083 | 6379 | 
| Shard 4       | Ethash       | 8048 | 8884 | 8084 | 6379 | 
| Shard 5       | Ethash       | 8058 | 8885 | 8085 | 6379 | 
| Shard 6       | Qkchash      | 8068 | 8886 | 8086 | 6378 | 
| Shard 7       | Qkchash      | 8078 | 8887 | 8087 | 6378 | 


|Chain |Shard |Hash Algo |Parameter for Ethminer shard ID|
| ---      | ---     |---  | --- |
| 0  | 0      | Ethash               | 1 |
| 1  |  0      | Ethash        | 10001 |
| 2  |  0       | Ethash              | 20001 |
| 3  |  0       | Ethash              | 30001 |
| 4  |  0       | Ethash              | 40001 |
| 5  |  0       | Ethash              | 50001 |
| 6 |  0       | Qkchash              | 60001 |
| 7 |  0       | Qkchash              | 70001 |

### Claymore 
It supports Claymore mining, which is dual Ethereum+Decred mining software. Download the [closed source software]((https://github.com/nanopool/Claymore-Dual-Miner/releases)) and connect the pool using the following command. 

    $ ./ethdcrminer64 -epool pool1.quarkchain.io:8008 $COINBASE_ADDRESS -mode 1 -allcoins 1

### NiceHash

This Ethash pool supports both NiceHash and ETHPROXY. 

    Autodetection process passes all known stratum modes.
    - 1st pass EthStratumClient::ETHEREUMSTRATUM2 (3)  Not supported
    - 2nd pass EthStratumClient::ETHEREUMSTRATUM  (2)  Supported (NiceHash)
    - 3rd pass EthStratumClient::ETHPROXY         (1)  Supported
    - 4th pass EthStratumClient::STRATUM          (0)  Not supported


![](https://i.imgur.com/OwKfnBD.png)


### Configuration

Configuration is actually simple, just read it twice and think twice before changing defaults.

**Don't copy config directly from this manual. Use the config.example_Ethash.json from the package,
otherwise you will get errors on start because of JSON comments.**

```javascript
{
  // Set to the number of CPU cores of your server
  "threads": 2,
  // Prefix for keys in redis store
  "coin": "qkc",
  // Give unique name to each instance
  "name": "main",

  "proxy": {
    "enabled": true,

    // Bind HTTP mining endpoint to this IP:PORT
    "listen": "0.0.0.0:8888",

    // Allow only this header and body size of HTTP request from miners
    "limitHeadersSize": 1024,
    "limitBodySize": 256,

    /* Set to true if you are behind CloudFlare (not recommended) or behind http-reverse
      proxy to enable IP detection from X-Forwarded-For header.
      Advanced users only. It's tricky to make it right and secure.
    */
    "behindReverseProxy": false,

    // Stratum mining endpoint
    "stratum": {
      "enabled": true,
      // Bind stratum mining socket to this IP:PORT
      "listen": "0.0.0.0:8008",
      "timeout": "120s",
      "maxConn": 8192,
      // Fill in the shard Id here
      "shardId": "0x1",
    },

    // Try to get new job from geth in this interval
    "blockRefreshInterval": "120ms",
    "stateUpdateInterval": "3s",
    // Require this share difficulty from miners
    "difficulty": 2000000000,

    /* Reply error to miner instead of job if redis is unavailable.
      Should save electricity to miners if pool is sick and they didn't set up failovers.
    */
    "healthCheck": true,
    // Mark pool sick after this number of redis failures.
    "maxFails": 100,
    // TTL for workers stats, usually should be equal to large hashrate window from API section
    "hashrateExpiration": "3h",

    "policy": {
      "workers": 8,
      "resetInterval": "60m",
      "refreshInterval": "1m",

      "banning": {
        "enabled": false,
        /* Name of ipset for banning.
        Check http://ipset.netfilter.org/ documentation.
        */
        "ipset": "blacklist",
        // Remove ban after this amount of time
        "timeout": 1800,
        // Percent of invalid shares from all shares to ban miner
        "invalidPercent": 30,
        // Check after after miner submitted this number of shares
        "checkThreshold": 30,
        // Bad miner after this number of malformed requests
        "malformedLimit": 5
      },
      // Connection rate limit
      "limits": {
        "enabled": false,
        // Number of initial connections
        "limit": 30,
        "grace": "5m",
        // Increase allowed number of connections on each valid share
        "limitJump": 10
      }
    }
  },

  // Provides JSON data for frontend which is static website
  "api": {
    "enabled": true,
    "listen": "0.0.0.0:8080",
    // Collect miners stats (hashrate, ...) in this interval
    "statsCollectInterval": "5s",
    // Purge stale stats interval
    "purgeInterval": "10m",
    // Fast hashrate estimation window for each miner from it's shares
    "hashrateWindow": "30m",
    // Long and precise hashrate from shares, 3h is cool, keep it
    "hashrateLargeWindow": "3h",
    // Collect stats for shares/diff ratio for this number of blocks
    "luckWindow": [64, 128, 256],
    // Max number of payments to display in frontend
    "payments": 50,
    // Max numbers of blocks to display in frontend
    "blocks": 50,

    /* If you are running API node on a different server where this module
      is reading data from redis writeable slave, you must run an api instance with this option enabled in order to purge hashrate stats from main redis node.
      Only redis writeable slave will work properly if you are distributing using redis slaves.
      Very advanced. Usually all modules should share same redis instance.
    */
    "purgeOnly": false
  },

  // Check health of each geth node in this interval
  "upstreamCheckInterval": "5s",

  /* List of geth nodes to poll for new jobs. Pool will try to get work from
    first alive one and check in background for failed to back up.
    Current block template of the pool is always cached in RAM indeed.
  */
  "upstream": [
    {
      "name": "main",
      "url": "http://127.0.0.1:38391",
      "timeout": "10s"
    },
    {
      "name": "backup",
      "url": "http://127.0.0.2:38391",
      "timeout": "10s"
    }
  ],

  // This is standard redis connection options
  "redis": {
    // Where your redis instance is listening for commands
    "endpoint": "127.0.0.1:6379",
    "poolSize": 10,
    "database": 0,
    "password": ""
  },

  // This module periodically remits ether to miners
  "unlocker": {
    "enabled": false,
    // Pool fee percentage
    "poolFee": 1.0,
    // Pool fees beneficiary address (leave it blank to disable fee withdrawals)
    "poolFeeAddress": "",
    // Donate 10% from pool fees to developers
    "donate": true,
    // Unlock only if this number of blocks mined back
    "depth": 120,
    // Simply don't touch this option
    "immatureDepth": 20,
    // Keep mined transaction fees as pool fees
    "keepTxFees": false,
    // Run unlocker in this interval
    "interval": "10m",
    // QuarkChain instance node rpc endpoint for unlocking blocks
    "daemon": "http://127.0.0.1:38391",
    // Rise error if can't reach geth in this amount of time
    "timeout": "10s"
  },

  // Pay out miners using this module
  "payouts": {
    "enabled": false,
    // Require minimum number of peers on node
    "requirePeers": 25,
    // Run payouts in this interval
    "interval": "12h",
    // QuarkChain instance node rpc endpoint for payouts processing
    "daemon": "http://127.0.0.1:38391",
    // Rise error if can't reach geth in this amount of time
    "timeout": "10s",
    // Address with pool balance
    "address": "0x0",
    // Let geth to determine gas and gasPrice
    "autoGas": true,
    // Gas amount and price for payout tx (advanced users only)
    "gas": "21000",
    "gasPrice": "50000000000",
    // Send payment only if miner's balance is >= 0.5 Ether
    "threshold": 500000000,
    // Perform BGSAVE on Redis after successful payouts session
    "bgsave": false
  }
}
```

    
    
