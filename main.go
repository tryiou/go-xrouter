// Copyright (c) 2020 The Blocknet developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tryiou/go-xrouter/blockcfg"
	"github.com/tryiou/go-xrouter/xrouter"
)

func main() {
	log.SetOutput(os.Stdout)

	config := blockcfg.MainnetParams
	// Manually set seed node (via ip or dns)
	//config.DNSSeeds = []chaincfg.DNSSeed{
	//	{"seed1.blocknet.co", false}, // optional direct connect to trusted node
	//}
	// Instantiate the xrouter client
	client, err := xrouter.NewClient(config)
	if err != nil {
		log.Println(err.Error())
		return
	}
	// Start xrouter (this will begin querying the network)
	client.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	defer shutdown(client)

	// Wait for xrouter to be ready
	if ready, err := client.WaitForXRouter(ctx); err != nil || !ready {
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		log.Println("XRouter failed to connect and obtain service nodes", errStr)
		return
	}
	log.Printf("XRouter is ready")

	// List all network services
	//for _, service := range client.ListNetworkServices() {
	//	log.Printf(service)
	//}

	ctx2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel2()
	queryCount := 1
	if err := client.WaitForServices(ctx2, []string{"xrs::CCSinglePrice", "xr::BTC"}, queryCount); err != nil {
		log.Printf("error: %v", err)
		return
	}

	{
		// Query the price oracle to obtain Bitcoin's price in USD
		var params []interface{}
		params = append(params, "BTC", "USD")
		if reply, flag, err := client.CallService("xrs::CCSinglePrice", params, queryCount); err != nil {
			log.Printf("error: %v", err)
			return
		} else {
			if reply == nil {
				log.Printf("No replies found. %v\n", flag)
			} else {
				log.Printf("Result from %v: %v. %v", hex.EncodeToString(reply.Pubkey), string(reply.Reply), flag)

			}
		}
	}

	{
		// Query the BTC oracle to obtain the chain height
		if reply, flag, err := client.GetBlockCount("xr::BTC", queryCount); err != nil {
			log.Printf("error: %v", err)
			return
		} else {
			if reply == nil {
				log.Printf("No replies found. %v\n", flag)
			} else {
				log.Printf("Result from %v: %v. %v", hex.EncodeToString(reply.Pubkey), string(reply.Reply), flag)

			}
		}
	}
}

func shutdown(client *xrouter.Client) {
	if err := client.Stop(); err != nil {
		fmt.Printf("error shutdown: %v", err)
	}
}
