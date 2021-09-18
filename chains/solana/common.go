package solana

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/stafiprotocol/chainbridge/shared/solana"
	"github.com/stafiprotocol/chainbridge/utils/msg"
	"github.com/stafiprotocol/solana-go-sdk/bridgeprog"
	solClient "github.com/stafiprotocol/solana-go-sdk/client"
	solCommon "github.com/stafiprotocol/solana-go-sdk/common"
	"github.com/stafiprotocol/solana-go-sdk/sysprog"
	solTypes "github.com/stafiprotocol/solana-go-sdk/types"
)

var retryLimit = 50
var waitTime = time.Second * 5
var backCheckLen = 10

type EventTransferOut struct {
	Transfer     solCommon.PublicKey
	Receiver     []byte
	Amount       uint64
	DestChainId  uint8
	ResourceId   [32]byte
	DepositNonce uint64
}


func (w *writer) printContentError(m msg.Message, err error) {
	w.log.Error("msg resolve failed", "source", m.Source, "dest", m.Destination, "type", m.Type, "err", err)
}

// submitMessage inserts the chainId into the msg and sends it to the router
func (w *writer) submitMessage(m msg.Message) bool {
	err := w.router.Send(m)
	if err != nil {
		w.log.Error("failed to process event", "err", err)
		return false
	}

	return true
}

func GetProposalAccountPubkey(baseAccount, programID solCommon.PublicKey, srcChainId uint8, depositCount uint64) (solCommon.PublicKey, string) {
	seed := fmt.Sprintf("stafi mint proposal: %d/%d", srcChainId, depositCount)
	return solCommon.CreateWithSeed(baseAccount, seed, programID), seed
}

func mapToString(accountsMap map[solCommon.PublicKey]solClient.GetStakeActivationResponse) string {
	ret := ""
	for account, active := range accountsMap {
		ret = ret + account.ToBase58() + fmt.Sprintf(" : %+v", active) + "\n"
	}
	return ret
}

func (w *writer) waitingForProposalExe(rpcClient *solClient.Client, proposalAccountAddress, processName string) bool {
	retry := 0
	for {
		if retry >= retryLimit {
			w.log.Error(fmt.Sprintf("[%s] GetMintProposalInfo reach retry limit", processName),
				"proposalAccountAddress ", proposalAccountAddress)
			return false
		}
		proposalAccountInfo, err := rpcClient.GetMintProposalInfo(context.Background(), proposalAccountAddress)
		if err == nil && proposalAccountInfo.DidExecute == 1 {
			break
		} else {
			w.log.Warn(fmt.Sprintf("[%s] proposalAccount not execute yet, waiting...", processName),
				"proposalAccountInfo", proposalAccountAddress)
			time.Sleep(waitTime)
			retry++
			continue
		}
	}
	return true
}

func (w *writer) waitingForProposalAccountCreate(rpcClient *solClient.Client, proposalAccountAddress, processName string) bool {
	retry := 0
	for {
		if retry >= retryLimit {
			w.log.Error(fmt.Sprintf("[%s] GetMintProposalInfo reach retry limit", processName),
				"proposal account address", proposalAccountAddress)
			return false
		}
		_, err := rpcClient.GetMintProposalInfo(context.Background(), proposalAccountAddress)
		if err != nil {
			w.log.Warn(fmt.Sprintf("[%s] GetMintProposalInfo failed, waiting...", processName),
				"proposal account address", proposalAccountAddress,
				"err", err)
			time.Sleep(waitTime)
			retry++
			continue
		} else {
			break
		}
	}
	return true
}

func (w *writer) createProposalAccount(
	rpcClient *solClient.Client,
	poolClient *solana.PoolClient,
	toAccountApubkey,
	proposalAccountPubkey solCommon.PublicKey,
	proposalAccountSeed string,
	resourceId [32]byte,
	amount uint64,
	processName string,
) bool {
	res, err := rpcClient.GetRecentBlockhash(context.Background())
	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] GetRecentBlockhash failed", processName),
			"err", err)
		return false
	}
	miniMumBalanceForRent, err := rpcClient.GetMinimumBalanceForRentExemption(context.Background(), solClient.MintProposalInfoLengthDefault)
	if err != nil || miniMumBalanceForRent == 0 {
		w.log.Error(fmt.Sprintf("[%s] GetMinimumBalanceForRentExemption failed", processName),
			"err", err)
		return false
	}
	//send from one relayers
	//create multisig tx account of this era
	rawTx, err := solTypes.CreateRawTransaction(solTypes.CreateRawTransactionParam{
		Instructions: []solTypes.Instruction{
			sysprog.CreateAccountWithSeed(
				poolClient.FeeAccount.PublicKey,
				proposalAccountPubkey,
				poolClient.ProposalBaseAccount.PublicKey,
				poolClient.BridgeProgramId,
				proposalAccountSeed,
				miniMumBalanceForRent,
				solClient.MintProposalInfoLengthDefault,
			),
			bridgeprog.CreateMintProposal(
				poolClient.BridgeProgramId,
				poolClient.BridgeAccountPubkey,
				proposalAccountPubkey,
				toAccountApubkey,
				poolClient.FeeAccount.PublicKey,
				resourceId,
				amount,
				poolClient.TokenProgramId,
			),
		},
		Signers:         []solTypes.Account{poolClient.FeeAccount, poolClient.ProposalBaseAccount},
		FeePayer:        poolClient.FeeAccount.PublicKey,
		RecentBlockHash: res.Blockhash,
	})

	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] createProposalAccount CreateRawTransaction failed", processName),
			"err", err)
		return false
	}

	txHash, err := rpcClient.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] createProposalAccount SendRawTransaction failed", processName),
			"err", err)
		return false
	}
	w.log.Info(fmt.Sprintf("[%s] create proposal account has send", processName),
		"tx hash", txHash,
		"proposal account", proposalAccountPubkey.ToBase58())
	return true
}

func (w *writer) approveProposal(
	rpcClient *solClient.Client,
	poolClient *solana.PoolClient,
	proposalAccountPubkey,
	mintAccount,
	toAccount solCommon.PublicKey,
	processName string) bool {
	res, err := rpcClient.GetRecentBlockhash(context.Background())
	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] GetRecentBlockhash failed", processName),
			"err", err)
		return false
	}
	rawTx, err := solTypes.CreateRawTransaction(solTypes.CreateRawTransactionParam{
		Instructions: []solTypes.Instruction{
			bridgeprog.ApproveMintProposal(
				poolClient.BridgeProgramId,
				poolClient.BridgeAccountPubkey,
				poolClient.BridgePdaPubkey,
				proposalAccountPubkey,
				poolClient.FeeAccount.PublicKey,
				mintAccount,
				toAccount,
				poolClient.TokenProgramId,
			),
		},
		Signers:         []solTypes.Account{poolClient.FeeAccount},
		FeePayer:        poolClient.FeeAccount.PublicKey,
		RecentBlockHash: res.Blockhash,
	})

	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] approveProposal CreateRawTransaction failed", processName),
			"err", err)
		return false
	}

	txHash, err := rpcClient.SendRawTransaction(context.Background(), rawTx)
	if err != nil {
		w.log.Error(fmt.Sprintf("[%s] approveProposal SendRawTransaction failed", processName),
			"err", err)
		return false
	}

	w.log.Info(fmt.Sprintf("[%s] approveProposal multisig tx account has send", processName),
		"tx hash", txHash,
		"proposal account", proposalAccountPubkey.ToBase58())

	return true
}

func (w *writer) IsProposalExe(proposalAccountPubkey solCommon.PublicKey) bool {
	accountInfo, err := w.conn.GetQueryClient().GetMintProposalInfo(context.Background(), proposalAccountPubkey.ToBase58())
	if err == nil && accountInfo.DidExecute == 1 {
		return true
	}
	return false
}

func (w *writer) CheckProposalAccount(
	proposalAccountPubkey solCommon.PublicKey,
	Mint solCommon.PublicKey,
	To solCommon.PublicKey,
	Amount uint64,
) bool {
	accountInfo, err := w.conn.GetQueryClient().GetMintProposalInfo(context.Background(), proposalAccountPubkey.ToBase58())
	if err == nil {
		BridgeBts, err := solCommon.SerializeData(w.conn.poolClient.BridgeAccountPubkey)
		if err != nil {
			return false
		}
		MintBts, err := solCommon.SerializeData(Mint)
		if err != nil {
			return false
		}
		ToBts, err := solCommon.SerializeData(To)
		if err != nil {
			return false
		}
		AmountBts, err := solCommon.SerializeData(Amount)
		if err != nil {
			return false
		}
		TokenProgramBts, err := solCommon.SerializeData(w.conn.poolClient.TokenProgramId)
		if err != nil {
			return false
		}

		onchainBridgeBts, err := solCommon.SerializeData(accountInfo.Bridge)
		if err != nil {
			return false
		}
		onchainMintBts, err := solCommon.SerializeData(accountInfo.Mint)
		if err != nil {
			return false
		}
		onchainToBts, err := solCommon.SerializeData(accountInfo.To)
		if err != nil {
			return false
		}
		onchainAmountBts, err := solCommon.SerializeData(accountInfo.Amount)
		if err != nil {
			return false
		}
		onchainTokenProgramBts, err := solCommon.SerializeData(accountInfo.TokenProgram)
		if err != nil {
			return false
		}
		if bytes.Equal(onchainBridgeBts, BridgeBts) &&
			bytes.Equal(onchainMintBts, MintBts) &&
			bytes.Equal(onchainToBts, ToBts) &&
			bytes.Equal(onchainAmountBts, AmountBts) &&
			bytes.Equal(onchainTokenProgramBts, TokenProgramBts) {
			return true
		}
		w.log.Error("Check proposal account not equal ", "account", accountInfo)
	}

	return false
}