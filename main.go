package main

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"

	"github.com/cosmos/cosmos-sdk/types/query"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

type Chain struct {
	ChainID   string `yaml:"chain_id"`
	Validator string `yaml:"validator"`
	Node      string `yaml:"node"`
}

type chains struct {
	Chains []Chain `yaml:"chains"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	chainsYaml, err := os.ReadFile("chains.yml")
	if err != nil {
		log.Fatal(err)
	}

	chains := &chains{}
	err = yaml.Unmarshal(chainsYaml, chains)
	if err != nil {
		log.Fatal(err)
	}

	for _, chain := range chains.Chains {
		//log.Println("checking", chain.ChainID)
		client, err := rpchttp.New(chain.Node, "/websocket")
		if err != nil {
			log.Printf("error connecting to %s: %s\n", chain.ChainID, err)
			continue
		}

		status, err := client.Status(context.Background())
		if err != nil {
			log.Printf("error getting status %s\n", err)
			continue
		}
		if status.NodeInfo.Network != chain.ChainID {
			log.Printf("chain id mismatch %s != %s\n", status.NodeInfo.Network, chain.ChainID)
			continue
		}

		qpr := gov.QueryProposalsRequest{
			ProposalStatus: gov.StatusVotingPeriod,
			//Voter:          "",
			//Depositor:      "",
			Pagination: &query.PageRequest{Limit: 100, Offset: 0},
		}

		q, err := qpr.Marshal()
		if err != nil {
			log.Printf("error marshaling %s\n", err)
			continue
		}

		response := &gov.QueryProposalsResponse{}
		for _, path := range []string{"/cosmos.gov.v1beta1.Query/Proposals", "/cosmos.gov.v1.Query/Proposals"} {
			pResp, err := client.ABCIQuery(context.Background(), path, q)
			if err != nil {
				log.Printf("error getting votes %s\n", err)
				continue
			}
			err = response.Unmarshal(pResp.Response.Value)
			if err != nil {
				log.Printf("error unmarshaling %s\n", err)
				continue
			}
			if len(response.Proposals) != 0 {
				break
			}
		}

		if len(response.Proposals) == 0 {
			fmt.Print("* no proposals on ", chain.ChainID, "\n\n")
			continue
		}
		fmt.Printf("* found %d proposals for %s\n", len(response.Proposals), chain.ChainID)
		for _, proposal := range response.Proposals {
			qr := gov.QueryVoteRequest{
				ProposalId: proposal.ProposalId,
				Voter:      chain.Validator,
			}
			q, err := qr.Marshal()
			if err != nil {
				log.Printf("error marshaling %s\n", err)
				continue
			}
			pResp, err := client.ABCIQuery(context.Background(), "/cosmos.gov.v1beta1.Query/Vote", q)
			if err != nil {
				log.Printf("error getting vote %s\n", err)
				continue
			}
			response := &gov.QueryVoteResponse{}
			err = response.Unmarshal(pResp.Response.Value)
			if err != nil {
				log.Printf("error unmarshaling %s\n", err)
				continue
			}
			voted := "❌"
			if response.Vote.Voter == chain.Validator {
				voted = "✅"
			}
			fmt.Printf("%s proposal %d ends: %s\n", voted, proposal.ProposalId, proposal.VotingEndTime.Local().Format("2006-01-02 15:04:05"))
		}
		fmt.Println("")
	}
}
