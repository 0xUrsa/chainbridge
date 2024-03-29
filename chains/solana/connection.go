package solana

import (
	"fmt"
	"strings"

	"github.com/ChainSafe/log15"
	"github.com/stafiprotocol/chainbridge/shared/solana"
	"github.com/stafiprotocol/chainbridge/shared/solana/vault"
	"github.com/stafiprotocol/chainbridge/utils/core"
	solClient "github.com/stafiprotocol/solana-go-sdk/client"
	solCommon "github.com/stafiprotocol/solana-go-sdk/common"
	solTypes "github.com/stafiprotocol/solana-go-sdk/types"
)

type Connection struct {
	endpoint    string
	queryClient *solClient.Client
	poolClient  *solana.PoolClient
	log         log15.Logger
	stop        <-chan int
}

type PoolAccounts struct {
	FeeAccount          string `json:"feeAccount"`
	ProposalBaseAccount string `json:"proposalBaseAccount"`
	BridgeAccountPubkey string `json:"bridgeAccountPubkey"`
	BridgePdaPubkey     string `json:"bridgePdaPubkey"`
	BridgeProgramId     string `json:"bridgeProgramId"`
	TokenProgramId      string `json:"TokenProgramId"`
}

func NewConnection(cfg *core.ChainConfig, log log15.Logger, stop <-chan int) (*Connection, error) {
	log.Info("NewConnection", "name", cfg.Name, "KeystorePath", cfg.KeystorePath, "Endpoint", cfg.Endpoint)
	pAccounts := PoolAccounts{
		FeeAccount:          cfg.Opts["feeAccount"],
		ProposalBaseAccount: cfg.Opts["proposalBaseAccount"],
		BridgeAccountPubkey: cfg.Opts["bridgeAccountPubkey"],
		BridgePdaPubkey:     cfg.Opts["bridgePdaPubkey"],
		BridgeProgramId:     cfg.Opts["bridgeProgramId"],
		TokenProgramId:      cfg.Opts["TokenProgramId"],
	}
	if !strings.HasSuffix(cfg.KeystorePath, "/") {
		cfg.KeystorePath += "/"
	}
	v, err := vault.NewVaultFromWalletFile(cfg.KeystorePath + cfg.From)
	if err != nil {
		return nil, err
	}
	boxer, err := vault.SecretBoxerForType(v.SecretBoxWrap)
	if err != nil {
		return nil, fmt.Errorf("secret boxer: %w", err)
	}

	if err := v.Open(boxer); err != nil {
		return nil, fmt.Errorf("opening: %w", err)
	}

	privKeyMap := make(map[string]vault.PrivateKey)
	for _, privKey := range v.KeyBag {
		privKeyMap[privKey.PublicKey().String()] = privKey
	}

	poolAccounts := solana.PoolAccounts{
		FeeAccount:          solTypes.AccountFromPrivateKeyBytes(privKeyMap[pAccounts.FeeAccount]),
		ProposalBaseAccount: solTypes.AccountFromPrivateKeyBytes(privKeyMap[pAccounts.ProposalBaseAccount]),
		BridgeAccountPubkey: solCommon.PublicKeyFromString(pAccounts.BridgeAccountPubkey),
		BridgePdaPubkey:     solCommon.PublicKeyFromString(pAccounts.BridgePdaPubkey),
		BridgeProgramId:     solCommon.PublicKeyFromString(pAccounts.BridgeProgramId),
		TokenProgramId:      solCommon.PublicKeyFromString(pAccounts.TokenProgramId),
	}
	poolClient := solana.NewPoolClient(log, solClient.NewClient(cfg.EndpointList), poolAccounts)

	return &Connection{
		endpoint:    cfg.Endpoint,
		queryClient: solClient.NewClient(cfg.EndpointList),
		log:         log,
		stop:        stop,
		poolClient:  poolClient,
	}, nil
}

func (c *Connection) GetQueryClient() *solClient.Client {
	return c.queryClient
}
