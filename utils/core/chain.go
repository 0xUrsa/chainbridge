// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package core

import (
	"github.com/stafiprotocol/chainbridge/utils/msg"
)

type Chain interface {
	Start() error // Start chain
	SetRouter(*Router)
	Id() msg.ChainId
	Name() string
	Stop()
}

type ChainConfig struct {
	Name           string            // Human-readable chain name
	Id             msg.ChainId       // ChainID
	Endpoint       string            // url for rpc endpoint
	EndpointList   []string          // url list for rpc endpoint
	From           string            // address of key to use
	KeystorePath   string            // Location of key files
	Insecure       bool              // Indicated whether the test keyring should be used
	BlockstorePath string            // Location of blockstore
	FreshStart     bool              // If true, blockstore is ignored at start.
	LatestBlock    bool              // If true, overrides blockstore or latest block in config and starts from current block
	Opts           map[string]string // Per chain options
	Symbols        []interface{}     // map info for symbols and resourceIds
}
