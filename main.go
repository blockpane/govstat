package main

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/types/query"
	gov1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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

		// this seems to marshal the same on v1 and v1beta1
		qpr := gov1b1.QueryProposalsRequest{
			ProposalStatus: gov1b1.StatusVotingPeriod,
			Pagination:     &query.PageRequest{Limit: 100, Offset: 0},
		}

		q, err := qpr.Marshal()
		if err != nil {
			log.Printf("error marshaling %s\n", err)
			continue
		}

		response1b1 := &gov1b1.QueryProposalsResponse{}
		response1 := &gov1.QueryProposalsResponse{}
		var path string
		var props int
		for _, path = range []string{"v1", "v1beta1"} {
			pResp, err := client.ABCIQuery(context.Background(), fmt.Sprintf("/cosmos.gov.%s.Query/Proposals", path), q)
			if err != nil {
				log.Printf("error getting votes %s\n", err)
				continue
			}
			switch path {
			case "v1beta1":
				err = response1b1.Unmarshal(pResp.Response.Value)
			case "v1":
				err = response1.Unmarshal(pResp.Response.Value)
			}
			if err != nil {
				log.Printf("error unmarshaling %s\n", err)
				continue
			}
			switch path {
			case "v1beta1":
				props = len(response1b1.Proposals)
			case "v1":
				props = len(response1.Proposals)
			}
			if props != 0 {
				break
			}
		}
		if props == 0 {
			fmt.Print("* no proposals on ", chain.ChainID, "\n\n")
			continue
		}
		fmt.Printf("* found %d proposals for %s\n", props, chain.ChainID)
		switch path {
		case "v1beta1":
			v1b1Response(response1b1, client, chain)
		case "v1":
			v1Response(response1, client, chain)
		}
	}
}

func v1Response(response *gov1.QueryProposalsResponse, client *rpchttp.HTTP, chain Chain) {
	for _, proposal := range response.Proposals {
		qr := gov1.QueryVoteRequest{
			ProposalId: proposal.Id,
			Voter:      chain.Validator,
		}
		q, err := qr.Marshal()
		if err != nil {
			log.Printf("error marshaling %s\n", err)
			continue
		}
		pResp, err := client.ABCIQuery(context.Background(), "/cosmos.gov.v1.Query/Vote", q)
		if err != nil {
			log.Printf("error getting vote %s\n", err)
			continue
		}
		// fmt.Println(len(pResp.Response.Value))
		response := &gov1.QueryVoteResponse{}
		err = response.Unmarshal(pResp.Response.Value)
		if err != nil {
			log.Printf("error unmarshaling %s\n", err)
			continue
		}
		voted := "❌"
		if response.Vote != nil && response.Vote.Voter == chain.Validator {
			voted = "✅"
		}
		fmt.Printf("%s proposal %d ends: %s\n", voted, proposal.Id, proposal.VotingEndTime.Local().Format("2006-01-02 15:04:05"))
	}
	fmt.Println("")
}

func v1b1Response(response *gov1b1.QueryProposalsResponse, client *rpchttp.HTTP, chain Chain) {
	for _, proposal := range response.Proposals {
		qr := gov1b1.QueryVoteRequest{
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
		// fmt.Println(len(pResp.Response.Value))
		response := &gov1b1.QueryVoteResponse{}
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
