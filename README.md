# Spartan NC Ethereum Node Installation User Guide

## Introduction

A Non-Cryptocurrency Public Chain is a transformed public chain framework based on an existing public chain. Gas Credit transfers are not permitted between standard wallets. There will be no cryptocurrency incentives for mining or participating in consensus.

## 1. About Spartan-I Chain (Powered by NC Ethereum)

This document is a guide to install, configure and run a full node in the Non-Cryptocurrency (NC ETH) public blockchain. The Non-Cryptocurrency Ethereum blockchain is a blockchain compatible with Ethereum but runs independently from the public Ethereum blockchain. Full Nodes can freely join and exit the Spartan Network, synchronize block information of the entire chain and submit transaction requests to the network.

A full node of Non-Cryptocurrency Ethereum runs an EVM (Ethereum Virtual Machine) that allows developers to create smart contracts by solidity coding language in the blockchain. Also, different tools and wallets available for Ethereum (such as Truffle, HardHat, Metamask, etc…) can be used in the Non-Crypto Ethereum public blockchain.

Each Ethereum network has two identifiers, a network ID and a chain ID. Although they are often set to the same value, they are used for different purposes. The peer-to-peer communication between nodes uses the network ID, while the transaction signature process uses the chain ID.

NC ETH Network ID = Chain ID  = 9090

Below is the instruction for Linux.


## 2. Geth Installation

#### 2.1 Install by Commands

To build the node by commands, Go 1.15 or above should be installed into your server first:

Install `go` by the following steps:

Download and untar the installation file

```
wget https://go.dev/dl/go1.18.5.linux-amd64.tar.gz

tar -C /usr/local -zxvf go1.18.5.linux-amd64.tar.gz
```

Modify environment variables, for example in bash

```shell
vim /etc/profile

# insert at the bottom of the file
export PATH=$PATH:/usr/local/go/bin

source /etc/profile
```

Check the installation result

```
go version
```
Before compiling the source code, make sure that `gcc` has been successfully installed. If not, please install `gcc` first.

```
gcc -v
```

Download the source code of Spartan NC Ethereum from github:

```
git clone https://github.com/BSN-Spartan/NC-Ethereum.git
```

Compile the source code in `NC-Ethereum` directory:

```
cd NC-Ethereum
make all
cp -r build/bin/* /usr/bin/
```
#### 2.2 Install by Docker images

If you build the node by Docker, follow the commands below to install geth:

```
docker pull bsnspartan/nc-eth:1.10.17
```

## 3. Create a Node

#### 3.1 Node Initialization

Download [genesis.json](https://github.com/BSN-Spartan/NC-Ethereum/blob/main/spartan/genesis.json),
[static-nodes.json](https://github.com/BSN-Spartan/NC-Ethereum/blob/main/spartan/static-nodes.json),
[trusted-nodes.json](https://github.com/BSN-Spartan/NC-Ethereum/blob/main/spartan/trusted-nodes.json) to the current folder.

Copy `genesis.json` to node1 directory and initialize it:

```shell
mkdir node1
cp genesis.json node1/
geth --datadir node1 init node1/genesis.json
```

#### 3.2 Configure Node Files

copy `static-nodes.json` and `trusted-nodes.json` to `node1/geth/`:
```
cp static-nodes.json trusted-nodes.json node1/geth/
```
For detailed explanation of the two files, please refer to
https://geth.ethereum.org/docs/interface/peer-to-peer

#### 3.3 Start the Node

The data center has two types of nodes. One is called management node that interacts with the Data Center Management System; the other one is called business node. The two types of nodes are launched in different ways:

#### 3.3.1 Start the Node by Commands

##### Start the Management Node by Commands:

```
geth --networkid 9090 --datadir node1/ --syncmode 'full' --nodiscover --maxpeers 300 --verbosity 6 --ipcdisable --port 30004 --http --http.addr 0.0.0.0 --http.port 20004 --http.api 'eth,net,web3,txpool' --ws --ws.port 8544 --ws.addr 0.0.0.0 --ws.api 'eth,net,web3' --ws.origins '*' --allow-insecure-unlock --censorship.admin.address 0x94109ebFB3d4153a266e7AC08E8C6F868360DEE6
```
##### Start the Business Node by Commands:

```
geth --networkid 9090 --datadir node1/ --syncmode 'full' --nodiscover --maxpeers 300 --verbosity 6 --ipcdisable --port 30004 --http --http.addr 0.0.0.0 --http.port 20004 --http.api 'eth,net,web3' --ws --ws.port 8544 --ws.addr 0.0.0.0 --ws.api 'eth,net,web3' --ws.origins '*' --allow-insecure-unlock --censorship.admin.address 0x94109ebFB3d4153a266e7AC08E8C6F868360DEE6
```

Or you can execute in the background via `nohup`:

To stop the node in `nohup` mode, please refer to the below command:
```
pkill -INT geth
```

**Important parameters:**

- `networkid` -- network ID of NC Ethereum
- `datadir` -- the diretory to store data after the node is started
- `port` -- local port
- `http.port` -- rpc port, should be different from local port
- `http.api` -- pluggable API interface, can be queried by different modules or achieve different functions

Please keep all other parameters unchanged.

#### 3.3.2 Start the Node by Docker

##### Start the Management Node by Docker:

```
docker run -d -p 35004:35004 -p 25004:25004 -p 8554:8554 -v $PWD/node1:/node1 --restart=always --name spartan-nc-eth bsnspartan/nc-eth:1.10.17 --networkid 9090 --datadir /node1/ --syncmode 'full' --nodiscover --maxpeers 300 --verbosity 6 --ipcdisable --port 35004 --http --http.addr 0.0.0.0 --http.port 25004 --http.api 'eth,net,web3,txpool' --ws --ws.port 8554 --ws.addr 0.0.0.0 --ws.api 'eth,net,web3' --ws.origins '*' --allow-insecure-unlock --censorship.admin.address 0x94109ebFB3d4153a266e7AC08E8C6F868360DEE6
```

##### Start the Business Node by Docker:

```
docker run -d -p 35004:35004 -p 25004:25004 -p 8554:8554 -v $PWD/node1:/node1 --restart=always --name spartan-nc-eth bsnspartan/nc-eth:1.10.17 --networkid 9090 --datadir /node1/ --syncmode 'full' --nodiscover --maxpeers 300 --verbosity 6 --ipcdisable --port 35004 --http --http.addr 0.0.0.0 --http.port 25004 --http.api 'eth,net,web3' --ws --ws.port 8554 --ws.addr 0.0.0.0 --ws.api 'eth,net,web3' --ws.origins '*' --allow-insecure-unlock --censorship.admin.address 0x94109ebFB3d4153a266e7AC08E8C6F868360DEE6
```

You can change the port to your own one and remember to run this command where node1 directory is located.


## 4. Add a New Node (Optional)

The process of creating and configuration new nodes is the same as the one above, including initializing the node, configuring the node files, and finally starting the node.

## 5. Generate the Node Signature

When joining the Spartan Network as a VDC, the VDC will be rewarded a certain amount of NTT Incentives based on the quantity of the registered nodes and their health status. To achieve this, the VDC Owner should firstly provide the signature of the VDC node to verify the node's ownership.

### Node installed by Commands:

Execute the following command after the node is started:

```
geth validate --datadir node1/
```

datadir is the data directory of the node,you should specify this directory to store the data file of the node.

### **Node Installed by Docker**

Execute the following command after the node is started:

`docker exec spartan-nc-eth geth validate --datadir node1/`

### **Node Signature**

After executing the above commands, you will get the following information. You can fill it in the Data Center Management System when registering the node.

```shell
{
  "nodeId": "9ddd61e4f29d286228b0e4ea2fa0ab44bea60909f7633ad419a14a80ee7a5aa2",
  "address": "enode://5409333437067eea683b5671c7e846af1e7406e4d1fe18b4a3c9bc24c8fecdb729e1a47c6159dc4d4d99f18ea34250f3071c42d5c28599125a1f8ad758d4f0aa",
  "signature": "0xb5b6911b86cc3dfe8b3564bd6cdd978c80b24aff4487030e32c8678893ab598477286fc1f2c0b29822b3e060f2a8e37a44d95cdb32e52c70cce9b1a877a7cd6f01"
}

```

## 6.  Ethereum and Geth Documentation

Below is a list of useful online documentation about Ethereum and geth:

How to set up Geth and execute some basic tasks using the command line tools:

https://geth.ethereum.org/docs/getting-started


JSON-RPC API methods
Interacting with Geth requires sending requests to specific JSON-RPC API methods. Geth supports all standard JSON-RPC API endpoints. You can send RPC requests on the port 8545

https://geth.ethereum.org/docs/rpc/server



Developer Documentation

This documentation is designed to help you build with Ethereum. It covers Ethereum as a concept, explains the Ethereum tech stack, and documents advanced topics for more complex applications and use cases.

https://ethereum.org/en/developers/docs/


Smart Contract tutorials

A list of curated Ethereum tutorials to learn about coding smart contracts and DApps.

https://ethereum.org/en/developers/tutorials/


