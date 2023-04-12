# govstat

A very simple tool to query multiple chains for governance proposals.

The config is loaded from a file called `chains.yml` in the current working directory.
The format for this file is simple:

```yaml
---

# update the following with your own chain IDs, RPC endpoint, and validator addresses
# repeat as needed for each chain you want to monitor

chains:
- chain_id: "akashnet-2"
  validator: "akash1zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
  node: "http://somehost:26657"
- chain_id: "injective-1"
  validator: "inj1zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
  node: "http://someotherhost:26657"
```

It will query the RPC endpoint for governance proposals in the "voting_period" state showing
the status of the validator's vote on each proposal.
A red ❌ indicates the validator has not voted on the proposal, a green ✅ indicates the validator has.

## Example output:

```text
$ govstat

* found 2 proposals for akashnet-2
✅ proposal 199 ends: 2023-04-17 17:51:38
✅ proposal 200 ends: 2023-04-17 17:56:50

* found 2 proposals for injective-1
✅ proposal 218 ends: 2023-04-14 06:16:45
✅ proposal 219 ends: 2023-04-14 07:38:26

* found 2 proposals for osmosis-1
✅ proposal 481 ends: 2023-04-15 09:37:29
❌ proposal 482 ends: 2023-04-17 09:07:18

* no proposals on jackal-1

* found 1 proposals for juno-1
✅ proposal 282 ends: 2023-04-15 13:55:15

* found 1 proposals for kava_2222-10
✅ proposal 135 ends: 2023-04-13 16:26:39

* no proposals on mars-1

* no proposals on migaloo-1

* found 2 proposals for secret-4
✅ proposal 233 ends: 2023-04-18 06:59:47
✅ proposal 235 ends: 2023-04-19 04:28:05

* found 4 proposals for stargaze-1
✅ proposal 155 ends: 2023-04-12 18:34:49
✅ proposal 156 ends: 2023-04-14 07:54:31
✅ proposal 157 ends: 2023-04-14 09:19:52
✅ proposal 158 ends: 2023-04-14 09:22:00

* no proposals on nois-1

```

## Installation

```bash
$ git clone https://github.com/blockpane/govstat
$ cd govstat
$ go install ./...
```