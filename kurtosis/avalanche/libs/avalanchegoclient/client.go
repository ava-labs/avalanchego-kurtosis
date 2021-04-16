package avalanchegoclient

import (
	"fmt"
	"time"

	"github.com/otherview/avalanchego-kurtosis/kurtosis/avalanche/libs/constants"
	"github.com/ava-labs/avalanchego/api/admin"
	"github.com/ava-labs/avalanchego/api/health"
	"github.com/ava-labs/avalanchego/api/info"
	"github.com/ava-labs/avalanchego/api/ipcs"
	"github.com/ava-labs/avalanchego/api/keystore"
	"github.com/ava-labs/avalanchego/vms/avm"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/coreth/ethclient"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/sirupsen/logrus"
)

// Chain names
const (
	XChain = "X"
	CChain = "C"
)

// Client is a general client for avalanche
type Client struct {
	admin              *admin.Client
	xChain             *avm.Client
	health             *health.Client
	info               *info.Client
	ipcs               *ipcs.Client
	keystore           *keystore.Client
	platform           *platformvm.Client
	cChain             *evm.Client
	cChainEth          *ethclient.Client
	cChaiConcurrentEth *ConcurrentEthClient
	ipAddr             string
	port               int
}

// NewClient returns a Client for interacting with the Chain endpoints
func NewClient(ipAddr string, port int, requestTimeout time.Duration) *Client {
	uri := fmt.Sprintf("http://%s:%d", ipAddr, port)
	cClient, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d/ext/bc/C/ws", ipAddr, port))
	if err != nil {
		// retry loop on next call
		cClient = nil
	}
	return &Client{
		ipAddr:             ipAddr,
		port:               port,
		admin:              admin.NewClient(uri, requestTimeout),
		xChain:             avm.NewClient(uri, XChain, requestTimeout),
		health:             health.NewClient(uri, requestTimeout),
		info:               info.NewClient(uri, requestTimeout),
		ipcs:               ipcs.NewClient(uri, requestTimeout),
		keystore:           keystore.NewClient(uri, requestTimeout),
		platform:           platformvm.NewClient(uri, requestTimeout),
		cChain:             evm.NewCChainClient(uri, requestTimeout),
		cChainEth:          cClient,
		cChaiConcurrentEth: NewConcurrentEthClient(cClient),
	}
}

// PChainAPI ...
func (c *Client) PChainAPI() *platformvm.Client {
	return c.platform
}

// XChainAPI ...
func (c *Client) XChainAPI() *avm.Client {
	return c.xChain
}

// CChainAPI ...
func (c *Client) CChainAPI() *evm.Client {
	return c.cChain
}

// CChainEthAPI
func (c *Client) CChainEthAPI() *ethclient.Client {
	var err error
	var cClient *ethclient.Client
	if c.cChainEth == nil {
		for startTime := time.Now(); time.Since(startTime) < constants.TimeoutDuration; time.Sleep(time.Second) {
			cClient, err = ethclient.Dial(fmt.Sprintf("ws://%s:%d/ext/bc/C/ws", c.ipAddr, c.port))
			if err == nil {
				c.cChainEth = cClient
				return c.cChainEth
			}
		}

		logrus.Infof("About to panic, the avalanchegoclient is unable to contact the CChain at : %s because of %s",
			fmt.Sprintf("ws://%s:%d/ext/bc/C/ws", c.ipAddr, c.port),
			err.Error())
		panic(err)
	}

	return c.cChainEth
}

// CChaiConcurrentEth wraps the ethclient.Client in a concurrency-safe implementation
func (c *Client) CChaiConcurrentEth() *ConcurrentEthClient {
	if c.cChainEth == nil || c.cChaiConcurrentEth.client == nil {
		c.cChaiConcurrentEth = NewConcurrentEthClient(c.CChainEthAPI())
	}

	return c.cChaiConcurrentEth
}

// InfoAPI ...
func (c *Client) InfoAPI() *info.Client {
	return c.info
}

// HealthAPI ...
func (c *Client) HealthAPI() *health.Client {
	return c.health
}

// IpcsAPI ...
func (c *Client) IpcsAPI() *ipcs.Client {
	return c.ipcs
}

// KeystoreAPI ...
func (c *Client) KeystoreAPI() *keystore.Client {
	return c.keystore
}

// AdminAPI ...
func (c *Client) AdminAPI() *admin.Client {
	return c.admin
}

func (c *Client) Reconnect() *Client {
	c.cChainEth = nil
	c.CChaiConcurrentEth()
	return c
}
