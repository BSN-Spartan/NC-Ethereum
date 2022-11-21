# Spartan NC Ethereum Node Installation User Guide

## Introduction

A Non-Cryptocurrency Public Chain is a transformed public chain framework based on an existing public chain. Gas Credit transfers are not permitted between standard wallets. There are no cryptocurrency incentives for mining or participating in consensus. On Spartan Network, there are three Non-Cryptocurrency Public Chains at launch. We except to add more in the foreseeable future.

> As a clear demonstration, all commands in this document are run with root permission. These commands can also be run under normal user permissions, please set the file storage and configure the parameters properly.

## 1. About Spartan-I Chain (Powered by NC Ethereum)

This document is a guide to install, configure and run an full node in the Spartan-I Chain, which is powered by NC Ethereum. The Spartan-I Chain is a blockchain compatible with Ethereum that run independently from the public Ethereum blockchain. Full Nodes, which can freely join and exit the Spartan Network, synchronize block information of the entire chain and submit transaction requests to the network.

A Spartan-I full node runs an EVM (Ethereum Virtual Machine) that allows developers to use Solidity programming language to create smart contracts that are compatible with the Ethereum network. Also, all the different tools and wallets available for Ethereum (such as Truffle, HardHat, Metamask, etc…) can be directly used with Spartan-I Chain.

Ethereum-based networks have two identifiers, a network ID and a chain ID. Although they often have the same value, they have different uses. Peer-to-peer communication between nodes uses the network ID, while the transaction signature process uses the chain ID.

Spartan-I Chain Network ID = Chain ID  = 9090

Below is the instruction for Linux system.

## 2. Hardware Requirements
It is recommended to build Spartan-I Chain full nodes with the following requirements:

#### Minimum Requirements

- 2 CPU
- Memory: 4GB
- Disk: 100GB SSD
- Bandwidth: 20Mbps

#### Recommended Requirements

- 4 CPU
- Memory: 16GB
- Disk: 512GB SSD
- Bandwidth: 20Mbps

## 3. Full Node Installation

## 3.1. Geth Installation

There are 2 methods to install Geth: building from source and installing by Docker. Please refer to the installation method that is most applicable in your specific case.

#### 3.1.1 Building from Source

To build the node by commands, **Go 1.15** or above should be installed into your server first. Install `go` by following steps:

Download and untar the installation file

```
wget https://go.dev/dl/go1.18.5.linux-amd64.tar.gz

tar -C /usr/local -zxvf go1.18.5.linux-amd64.tar.gz
```

Change environment variables, for example in bash:

```
vi /etc/profile
```

Insert the parameter at the bottom of the file:

```shell
export PATH=$PATH:/usr/local/go/bin
```

Then, save the change and make the /etc/profile file take effect:

```
source /etc/profile
```

Now, check whether `go` has been correctly installed:

```
go version
```

![](https://raw.githubusercontent.com/BSN-Spartan/NC-Ethereum/main/.github/images/1.go_version.png)

Also, make sure that `gcc` has been successfully installed. Check by the following command:

```
gcc -v
```

![](https://raw.githubusercontent.com/BSN-Spartan/NC-Ethereum/main/.github/images/2.%20gcc.jpg)

Download the source code of Spartan NC Ethereum from github (`git` has been installed):

```
git clone https://github.com/BSN-Spartan/NC-Ethereum.git
```

![](https://raw.githubusercontent.com/BSN-Spartan/NC-Ethereum/main/.github/images/3.git.jpg)

Compile the source code in `NC-Ethereum` directory:

```
cd NC-Ethereum
make all
cp -r build/bin/* /usr/bin/
```

#### 3.1.2 Installing by Docker Images

If you build the node by Docker, Docker 18 or later version should be installed in your server first:

```shell
wget -qO- https://get.docker.com/ | sh
```

Grant your user permission to execute Docker commands:

```shell
sudo usermod -aG docker your-user
```

Now, check the docker version:

```
docker version
```

![](https://raw.githubusercontent.com/BSN-Spartan/NC-Ethereum/main/.github/images/4.1dockerversion.jpg)

Start docker:

```shell
systemctl start docker
```

Follow the commands below to install geth:

```
docker pull bsnspartan/nc-eth:latest
```

![](https://raw.githubusercontent.com/BSN-Spartan/NC-Ethereum/main/.github/images/4.docker.jpg)

## 3.2. Creating a Node

#### 3.2.1 Node Initialization

Create a new directory `node1/`:

```
mkdir node1
```

Copy [genesis.json](https://github.com/BSN-Spartan/NC-Ethereum/blob/main/spartan/genesis.json) file from `spartan/` directory to `node1/` directory:

```
cp ./spartan/genesis.json node1/
```

The structure is shown as follows (viewed by `tree node1` command):

```shell
node1
└── genesis.json

0 directories, 1 file
```

**Initialize `genesis.json` by command:**

```shell
geth --datadir node1 init node1/genesis.json
```

![](https://raw.githubusercontent.com/BSN-Spartan/NC-Ethereum/main/.github/images/6.initgenesis.jpg)

**Initialize `genesis.json` by Docker:**

```shell
docker run --rm -it -v $PWD/node1:/node1 bsnspartan/nc-eth:latest --datadir /node1 init /node1/genesis.json
```

![](https://raw.githubusercontent.com/BSN-Spartan/NC-Ethereum/main/.github/images/6.initgenesisdocker.jpg)

#### 3.2.2 Node Files Configuration

Copy [static-nodes.json](https://github.com/BSN-Spartan/NC-Ethereum/blob/main/spartan/static-nodes.json) and [trusted-nodes.json](https://github.com/BSN-Spartan/NC-Ethereum/blob/main/spartan/trusted-nodes.json) files from `spartan/` directory to `node1/geth/` directory:

```
cp ./spartan/static-nodes.json ./spartan/trusted-nodes.json node1/geth/
```
> For detailed explanation of the two files, please refer to
https://geth.ethereum.org/docs/interface/peer-to-peer

Now, the structure of `node1/` directory is like below:

```shell
node1
├── genesis.json
├── geth
│   ├── chaindata
│   │   ├── 000001.log
│   │   ├── CURRENT
│   │   ├── LOCK
│   │   ├── LOG
│   │   └── MANIFEST-000000
│   ├── lightchaindata
│   │   ├── 000001.log
│   │   ├── CURRENT
│   │   ├── LOCK
│   │   ├── LOG
│   │   └── MANIFEST-000000
│   ├── LOCK
│   ├── nodekey
│   ├── static-nodes.json
│   └── trusted-nodes.json
└── keystore

4 directories, 15 files
```

#### 3.2.3 Starting the Node

Each Data Center can only have one Default Node of Spartan-I Chain that interacts with the Data Center System. If a second Spartan-I full node is installed, this node then is called a regular full node. The two types of nodes launch in different ways.

#### 3.2.3.1 Starting the Node by Commands

##### Starting the Default Node:

```
geth --networkid 9090 --datadir node1/ --syncmode 'full' --nodiscover --maxpeers 300 --verbosity 6 --ipcdisable --port 30001 --http --http.addr 0.0.0.0 --http.port 8545 --http.api 'eth,net,web3,txpool' --ws --ws.port 8546 --ws.addr 0.0.0.0 --ws.api 'eth,net,web3' --ws.origins '*' --allow-insecure-unlock --censorship.admin.address 0x94109ebFB3d4153a266e7AC08E8C6F868360DEE6
```

After the node has been started, it will synchronize all blocks from Spartan-I Chain. This process will take time, and you can check it by the block number:

![](https://raw.githubusercontent.com/BSN-Spartan/NC-Ethereum/main/.github/images/8.startnode.jpg)

Or you can run the node in background by `nohup` command:

```
nohup geth --networkid 9090 --datadir node1/ --syncmode 'full' --nodiscover --maxpeers 300 --verbosity 6 --ipcdisable --port 30001 --http --http.addr 0.0.0.0 --http.port 8545 --http.api 'eth,net,web3,txpool' --ws --ws.port 8546 --ws.addr 0.0.0.0 --ws.api 'eth,net,web3' --ws.origins '*' --allow-insecure-unlock --censorship.admin.address 0x94109ebFB3d4153a266e7AC08E8C6F868360DEE6 > output.log 2>&1 &
```

Check the process of block synchronization from the log:

```
tail -f output.log
```

![](https://raw.githubusercontent.com/BSN-Spartan/NC-Ethereum/main/.github/images/9.nohuplog.jpg)

To stop the node in `nohup` mode, please refer to the below command:
```
pkill -INT geth
```

##### Starting a Regular Full Node:

```
geth --networkid 9090 --datadir node1/ --syncmode 'full' --nodiscover --maxpeers 300 --verbosity 6 --ipcdisable --port 30002 --http --http.addr 0.0.0.0 --http.port 8547 --http.api 'eth,net,web3' --ws --ws.port 8548 --ws.addr 0.0.0.0 --ws.api 'eth,net,web3' --ws.origins '*' --allow-insecure-unlock --censorship.admin.address 0x94109ebFB3d4153a266e7AC08E8C6F868360DEE6
```

**Important Parameters:**

* `--networkid` The network ID of Spartan-I Chain is 9090
* `--datadir` The diretory to store data after the node is started
* `--http` Enable the HTTP-RPC server
* `--http.addr` HTTP-RPC server listening interface (default: `localhost`)
* `--http.port` HTTP-RPC server listening port (default: `8545`)
* `--http.api` API's offered over the HTTP-RPC interface (default: `eth,net,web3`)
* `--ws` Enable the WS-RPC server
* `--ws.addr` WS-RPC server listening interface (default: `localhost`)
* `--ws.port` WS-RPC server listening port (default: `8546`)
* `--ws.api` API's offered over the WS-RPC interface (default: `eth,net,web3`)
* `--ws.origins` Origins from which to accept WebSocket requests

Please keep all other parameters unchanged.

#### 3.2.3.2 Starting the Node by Docker

##### Starting the Default Node:

```
docker run -d -p 30001:30001 -p 8545:8545 -p 8546:8546 -v $PWD/node1:/node1 --restart=always --name spartan-nc-eth bsnspartan/nc-eth:latest --networkid 9090 --datadir /node1/ --syncmode 'full' --nodiscover --maxpeers 300 --verbosity 6 --ipcdisable --port 30001 --http --http.addr 0.0.0.0 --http.port 8545 --http.api 'eth,net,web3,txpool' --ws --ws.port 8546 --ws.addr 0.0.0.0 --ws.api 'eth,net,web3' --ws.origins '*' --allow-insecure-unlock --censorship.admin.address 0x94109ebFB3d4153a266e7AC08E8C6F868360DEE6
```

##### Starting a Regular Full Node:

```
docker run -d -p 30001:30001 -p 8545:8545 -p 8546:8546 -v $PWD/node1:/node1 --restart=always --name spartan-nc-eth bsnspartan/nc-eth:latest --networkid 9090 --datadir /node1/ --syncmode 'full' --nodiscover --maxpeers 300 --verbosity 6 --ipcdisable --port 30001 --http --http.addr 0.0.0.0 --http.port 8545 --http.api 'eth,net,web3' --ws --ws.port 8546 --ws.addr 0.0.0.0 --ws.api 'eth,net,web3' --ws.origins '*' --allow-insecure-unlock --censorship.admin.address 0x94109ebFB3d4153a266e7AC08E8C6F868360DEE6
```

You can change the port to your own and remember to run this command where node1 directory is located.


## 4. Adding a New Node (Optional)

The process of adding new nodes to Spartan-I Chain is the same as building the regular full node above, including initializing the node, configuring the node files, and finally starting the node.

## 5. Generating the Node Signature

When joining the Spartan Network as a Data Center, the Data Center Operator will be rewarded a certain amount of NTT Incentives based on the quantity of the registered node. To achieve this, the Data Center Operator should firstly provide the signature of the node to verify the node's ownership.

#### Node installed by Commands:

Execute the following command after the node has been started:

```
geth validate --datadir node1/
```

`datadir` is the directory that stores the data of the node.

#### Node Installed by Docker

Execute the following command after the node has been started:

```
docker exec spartan-nc-eth geth validate --datadir node1/
```

### Node Signature

After executing the above commands, you will get the following information. You can fill it in the Data Center System when registering the node.

```shell
{
  "nodeId": "9ddd61e4f29d286228b0e4ea2fa0ab44bea60909f7633ad419a14a80ee7a5aa2",
  "address": "enode://5409333437067eea683b5671c7e846af1e7406e4d1fe18b4a3c9bc24c8fecdb729e1a47c6159dc4d4d99f18ea34250f3071c42d5c28599125a1f8ad758d4f0aa",
  "signature": "0xb5b6911b86cc3dfe8b3564bd6cdd978c80b24aff4487030e32c8678893ab598477286fc1f2c0b29822b3e060f2a8e37a44d95cdb32e52c70cce9b1a877a7cd6f01"
}

```

## 6.  Ethereum and Geth Documentation

Below is a list of useful online documentation about Ethereum and Geth:

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
