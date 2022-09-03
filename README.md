# Spartan NC Ethereum Node Installation User Guide

## Introduction

A Non-Cryptocurrency Public Chain is a transformed public chain framework based on an existing public chain. Gas Credit transfers are not permitted between standard wallets. There will be no cryptocurrency incentives for mining or participating in consensus.

## 1. About Spartan-I Chain (Powered by NC Ethereum)

This document is a guide to install, configure and run an accounting node in the Non-Cryptocurrency (NC ETH) public blockchain. The Non-Cryptocurrency Ethereum blockchain is a blockchain compatible with Ethereum that run independently from the public Ethereum blockchain. Accounting Nodes, can freely join and exit the Spartan Network, synchronize block information of the entire chain, and submit transaction requests to the network.

A Non-Cryptocurrency Ethereum node runs an EVM (Ethereum Virtual Machine) that allows developers to use solidity coding language to create smart contracts that are compatible with the Ethereum network. Also, all the different tools and wallets available for Ethereum (such as Truffle, HardHat, Metamask, etc…) can be used in a Non-Crypto Ethereum public blockchain.

Ethereum networks have two identifiers, a network ID and a chain ID. Although they often have the same value, they have different uses. Peer-to-peer communication between nodes uses the network ID, while the transaction signature process uses the chain ID.

NC ETH   Network Id = Chain Id  = 9090


## 2. Geth Installation

#### 2.1 Download the Source Code

Download the source code of Spartan NC Ethereum from github:
```
git clone https://github.com/BSN-Spartan/NC-Ethereum.git
```
#### 2.2 Install by Commands:

If you want to build the node by commands, Go 1.15 or above should be installed into your server first:

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

Compile the source code in `NC-Ethereum` directory:

```
cd NC-Ethereum
make all
cp -r build/bin/* /usr/local/bin/
```
#### 2.3 Install by Docker images:

If you build the node by Docker, follow the commands below to install geth:

```
docker pull ethereum/client-go:v1.10.8
docker pull statusteam/bootnode:v0.64.3
```

## 3. Create a Node

#### 3.1 Create an Account for the Node


Create the data directory node1 and create an account:

```shell
mkdir node1
geth --datadir node1 account new
#Input the password
Password: 123456
#Repeat the password
Repeat password: 123456
#Generate the information
Public address of the key: 0x381b085ED9f5674DB928E74d7b4f0347d4b512c6
Path of the secret key file: /root/keystore/UTC--2021-05-19T06-30-54.972732175Z--381b085ed9f5674db928e74d7b4f0347d4b512c6
```
Then, the keystore directory and its subfile will be generated in node1 directory.

You can save the input password and the created address into node1 directory as text files.

```
echo '0x381b085ED9f5674DB928E74d7b4f0347d4b512c6' > node1/accounts.txt
echo '123456' >node1/password.txt
```

#### 3.3 Node Initialization


Download [genesis.json](https://github.com/BSN-Spartan/NC-Ethereum/blob/main/spartan/genesis.json) , 
[static-nodes.json](https://github.com/BSN-Spartan/NC-Ethereum/blob/main/spartan/static-nodes.json) , 
[trusted-nodes.json](https://github.com/BSN-Spartan/NC-Ethereum/blob/main/spartan/trusted-nodes.json) to local folder.

Copy genesis.json to node1 directory and initialize it:

```shell
cp genesis.json node1/
geth --datadir node1 init node1/genesis.json
```

#### 3.4 Check enode Data

Run the command below with bootnode:

```shell
bootnode --nodekey=node1/geth/nodekey  --writeaddress
#Response data:
a1cca894f28d39c3d6367693da384a5e457f87ca9058947eb23ac4bf89e6fa176ef63432e7b0180397a97b4855d406bd45e30829f43ed4bd7291b99fe39b5264
```

Or, you can get the data by this command if using Docker:

```shell
docker run --rm -v $PWD/node1:/root statusteam/bootnode:v0.64.3 --nodekey=/root/geth/nodekey  --writeaddress
#Response data:
970b32110f1d2cf251b8f65dfa55029bb52ddfbb5c3e7632bf9cca8c2ea29aa7c88dd1787aafa32297c41f6cb787994dd687a5003146a2d2001e72812011789c
```

#### 3.5 Configure Node Files

copy `static-nodes.json` and `trusted-nodes.json` to `node1/geth/`:
```
cp static-nodes.json trusted-nodes.json node1/geth/
```
For detailed information, please refer to
https://geth.ethereum.org/docs/interface/peer-to-peer

#### 3.6 Start the Node

VDC have two types of nodes. One is a node that interacts with the Data Center system, which is the management node; the other is a business node. The two types of nodes launch in different ways. The details are as follows.



### Start the Management Node by Commands:

```
geth --networkid 9090 --datadir node1/ --syncmode 'full' --nodiscover --maxpeers 300 --verbosity 6 --ipcdisable --port 30004 --http --http.addr 0.0.0.0 --http.port 20004 --http.api 'eth,net,web3' --ws --ws.port 8544 --ws.addr 0.0.0.0 --ws.api 'eth,net,web3,txpool' --ws.origins '*' --allow-insecure-unlock --censorship.admin.address 0x94109ebFB3d4153a266e7AC08E8C6F868360DEE6
```
### Start the Business Nodes by Commands:

```
geth --networkid 9090 --datadir node1/ --syncmode 'full' --nodiscover --maxpeers 300 --verbosity 6 --ipcdisable --port 30004 --http --http.addr 0.0.0.0 --http.port 20004 --http.api 'eth,net,web3' --ws --ws.port 8544 --ws.addr 0.0.0.0 --ws.api 'eth,net,web3' --ws.origins '*' --allow-insecure-unlock --censorship.admin.address 0x94109ebFB3d4153a266e7AC08E8C6F868360DEE6
```



**Important parameters:**
`networkid` -- network ID of NC Ethereum
`datadir` -- the diretory to store data after the node is started
`port` -- local port
`http.port` -- rpc port, should be different from local port
`http.api` -- pluggable API interface, can be queried by different modules or achieve different functions

Please keep all other parameters unchanged.

Start the node by Docker:

```shell
node1 docker-compose.yaml
version: "3"
services:
  nce:
    image: ubuntu:22.04
    container_name: spartan1
    restart: always
    working_dir: /opt
    entrypoint: sh /opt/start.sh
    volumes:
      - ./geth:/usr/bin/geth
      - ./node1:/opt
      - /etc/localtime:/etc/localtime
    ports:
      - 30001:30001
      - 20001:20001
      - 8541:8541

# Directory structure
.
├── docker-compose.yaml
├── geth
└── node1
    ├── accounts.txt
    ├── genesis.json
    ├── geth
    ├── keystore
    ├── password.txt
    └── start.sh

```

## 4. Add the Node

New nodes are generated and configured in the same way as the first node, including generating a node wallet, initializing the new node, viewing the node enode information, writing the information to the static-nodes.json and trusted-nodes.json files, and finally starting the node.

## 5. Generate the Node Signature

When joining the Spartan Network as a VDC, the VDC Owner will be rewarded a certain amount of NTT Incentives based on the quantity of the registered node. To achieve this, the VDC Owner should firstly provide the signature of the VDC node to verify the node's ownership.

### Node installed by Commands:

Execute the following command after the node is started:

```
geth validate --datadir node1/
```

datadir is the data directory of the node,you should specify this directory to store the data file of the node.

### **Node Installed by Docker**

Execute below command:

`docker exec spartan-nc-eth geth validate --datadir node/`

### **Node Signature**

After executing the above commands，you will get the following information.  You can fill it in the Data Center Management System when registering the node .

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


