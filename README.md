# GO-MATRIX
---

### About
The MATRIX repository is based on go-ethereum which contains protocol changes to support the MATRIX protocol and a few other distinct features. This implements the MATRIX cryptocurrency, which maintains a separate ledger from the ETHEREUM network, for several reasons, the most immediate of which is that the consensus protocol is different.

### Highlights

+ High-performace TPS
+ Highly-regulated network hierarchy
+ Support various transaction type: One2Multiple, Offline, Smartcontract, AI transactions as well as support for rich texts, images and videos
+ AI Features: Formal Verification（Trial）、Natural Language Input、AI server

+ 高性能 TPS
+ 一对多交易
+ 智慧合约
+ 人工智能支持
+ AI server

### Documentation Guide

+ If you want to know more about MATRIX Web Wallet, please check out ['Guide to Web Wallet'](https://github.com/MatrixAINetwork/MATRIX_docs/blob/master/ENGLISH_DOCS/MATRIX_WEB_WALLET/Guide_to_Web_Wallet.md)

+ If you want to know more about MATRIX Block Explorer, please check out ['Guide to BlockChain Explorer'](https://github.com/MatrixAINetwork/MATRIX_docs/blob/master/ENGLISH_DOCS/MATRIX_Blockchain_Explorer/Guide_to_Blockchain_Explorer.md)

URL: [BlockChain Explorer](http://tom.matrix.io/)

+ If you want to know more about How to deploy MATRIX, please check out ['User Guide'](https://github.com/MatrixAINetwork/MATRIX_docs/blob/master/ENGLISH_DOCS/MATRIX_User_guide/User%20Guide.md)

+ If you want to better understand what a specific term refers to, please check out ['glossary'](https://github.com/MatrixAINetwork/MATRIX_docs/blob/master/ENGLISH_DOCS/Glossary/Glossary.md)


### 文档指引

+ 如果你想了解 MATRIX 网页钱包的操作方法，请查看[《MATRIX 网页钱包教程》](https://github.com/MatrixAINetwork/MATRIX_docs/blob/master/%E4%B8%AD%E6%96%87%E6%96%87%E6%A1%A3/%E7%BD%91%E9%A1%B5%E9%92%B1%E5%8C%85%E6%89%8B%E5%86%8C.md)

+ 如果你想了解 MATRIX 区块链浏览器的操作方法，请查看[《MATRIX 区块链浏览器指南》](https://github.com/MatrixAINetwork/MATRIX_docs/blob/master/%E4%B8%AD%E6%96%87%E6%96%87%E6%A1%A3/%E5%8C%BA%E5%9D%97%E9%93%BE%E6%B5%8F%E8%A7%88%E5%99%A8%E6%8C%87%E5%8D%97.md)

[MATRIX 区块链浏览器](http://tom.matrix.io/)

+ MATRIX 部署手册请查看[《部署手册》](https://github.com/MatrixAINetwork/MATRIX_docs/blob/master/%E4%B8%AD%E6%96%87%E6%96%87%E6%A1%A3/%E9%83%A8%E7%BD%B2%E6%89%8B%E5%86%8C.md)


### Getting Started
Welcome! This guide is intended to get you running on the MATRIX testnet. To ensure your client behaves gracefully throughout the setup process, please check your system meets the following requirements:


| OS      | Windows, Mac                                 |
|---------|----------------------------------------------|
| CPU     | 6 Core (Intel(R) Xeon(R) CPU X5670 @2.93GHz) |
| RAM     | 8G                                           |
| Free HD | 500G                                         |



### Build from Source

First of all, you need to clone the source code from MATRIX repository:

Git clone https://github.com/MatrixAINetwork/MATRIX-TESTNET.git, or

wget https://github.com/MatrixAINetwork/MATRIX-TESTNET/archive/master.zip

Building gman requires both a Go (version 1.7 or later) and a C compiler. You can install them using your favourite package manager. Once the dependencies are installed, run:

    cd MATRIX-TESTNET_GO
    make gman
or, to build the full suite of utilities:

    make all


### Docker Quick Start

One of the quickest ways to get MATRIX up and running on your machine is by using Docker:

    docker build –t image_name

    docker run -d  -p 8545:8545 –p 30303:30303 –p 40404:40404 imagename --fast --cache=512

This will start gman in fast-sync mode with a DB memory allowance of 1GB just as the above command does.  It will also create a persistent volume in your home directory for saving your blockchain as well as map the default ports. There is also an `alpine` tag available for a slim version of the image.

Do not forget `--rpcaddr 0.0.0.0`, if you want to access RPC from other containers and/or hosts. By default, `gman` binds to the local interface and RPC endpoints is not accessible from the outside.


### Operating a private network

Maintaining your own private network is more involved as a lot of configurations taken for granted in
the official networks need to be manually set up.

#### Defining the private genesis state

First, you'll need to create the genesis state of your networks, which all nodes need to be aware of
and agree upon. This consists of a small JSON file (e.g. name it `genesis.json`):


    {
        "config": {
        				"chainID": 1,
        				"byzantiumBlock": 0,
        				"eip155Block": 0,
                "eip158Block": 0                        				             
        },
		
        "alloc": {
	"0x000000000000000000000000000000000000000A": {
            "storage": {
    "0x0000000000000000000000000000000000000000000000000000000a444e554d":"0x0000000000000000000000000000000000000000000000000000000000000014",
    "0x0000000000000000000000000000000000000000000a44490000000000000000":"0x0000000000000000000000006a3217d128a76e4777403e092bde8362d4117773",
    "0x0000000000000000000000000000000000000000000a44490000000000000001":"0x0000000000000000000000000ead6cdb8d214389909a535d4ccc21a393dddba9",
    "0x0000000000000000000000000000000000000000000a44490000000000000002":"0x00000000000000000000000055fbba0496ef137be57d4c179a3a74c4d0c643be",
    "0x0000000000000000000000000000000000000000000a44490000000000000003":"0x000000000000000000000000915b5295dde0cebb11c6cb25828b546a9b2c9369",
    "0x0000000000000000000000000000000000000000000a44490000000000000004":"0x00000000000000000000000092e0fea9aba517398c2f0dd628f8cfc7e32ba984",
    "0x0000000000000000000000000000000000000000000a44490000000000000005":"0x0000000000000000000000007eb0bcd103810a6bf463d6d230ebcacc85157d19",
    "0x0000000000000000000000000000000000000000000a44490000000000000006":"0x000000000000000000000000cded44bd41476a69e8e68ba8286952c414d28af7",
    "0x0000000000000000000000000000000000000000000a44490000000000000007":"0x0000000000000000000000009cde10b889fca53c0a560b90b3cb21c2fc965d2b",
    "0x0000000000000000000000000000000000000000000a44490000000000000008":"0x0000000000000000000000007823a1bea7aca2f902b87effdd4da9a7ef1fc5fb",
    "0x0000000000000000000000000000000000000000000a44490000000000000009":"0x0000000000000000000000000f96b686b2c57a0f2d571a39eae66d61a74b5870",
    "0x0000000000000000000000000000000000000000000a4449000000000000000a":"0x000000000000000000000000328b4bb56a750ea86bd52329a3e368d051699bb2",
    "0x0000000000000000000000000000000000000000000a4449000000000000000b":"0x000000000000000000000000b09b89893fd55223ed2d9c06cda7afef867c7449",
    "0x0000000000000000000000000000000000000000000a4449000000000000000c":"0x000000000000000000000000aea37855eacb4b41ca0dbc6c744681f96fe09d94",
    "0x0000000000000000000000000000000000000000000a4449000000000000000d":"0x000000000000000000000000b142159adbfc2690b45da01e49cfa2379ddc2701",
    "0x0000000000000000000000000000000000000000000a4449000000000000000e":"0x000000000000000000000000b7efab17215a43983d766114feb69172587a4090",
    "0x0000000000000000000000000000000000000000000a4449000000000000000f":"0x000000000000000000000000f9e18acc86179925353713a4a5d0e9bf381fbc17",
    "0x0000000000000000000000000000000000000000000a44490000000000000010":"0x000000000000000000000000a121e6670439ba37e7244d4eb18e42bd6724ef0f",
    "0x0000000000000000000000000000000000000000000a44490000000000000011":"0x0000000000000000000000000a3f28de9682df49f9f393931062c5204c2bc404",
    "0x0000000000000000000000000000000000000000000a44490000000000000012":"0x0000000000000000000000008c3d1a9504a36d49003f1652fadb9f06c32a4408",
    "0x0000000000000000000000000000000000000000000a44490000000000000013":"0x00000000000000000000000005e3c16931c6e578f948231dca609d754c18fc09",
    "0x000000000000000000000005e3c16931c6e578f948231dca609d754c18fc0944":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
    "0x00000000000000000000000a3f28de9682df49f9f393931062c5204c2bc40444":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
    "0x00000000000000000000000ead6cdb8d214389909a535d4ccc21a393dddba944":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x00000000000000000000000f96b686b2c57a0f2d571a39eae66d61a74b587044":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x0000000000000000000000328b4bb56a750ea86bd52329a3e368d051699bb244":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x000000000000000000000055fbba0496ef137be57d4c179a3a74c4d0c643be44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x00000000000000000000006a3217d128a76e4777403e092bde8362d411777344":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x00000000000000000000007823a1bea7aca2f902b87effdd4da9a7ef1fc5fb44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x00000000000000000000007eb0bcd103810a6bf463d6d230ebcacc85157d1944":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x00000000000000000000008c3d1a9504a36d49003f1652fadb9f06c32a440844":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
    "0x0000000000000000000000915b5295dde0cebb11c6cb25828b546a9b2c936944":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x000000000000000000000092e0fea9aba517398c2f0dd628f8cfc7e32ba98444":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x00000000000000000000009cde10b889fca53c0a560b90b3cb21c2fc965d2b44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x0000000000000000000000a121e6670439ba37e7244d4eb18e42bd6724ef0f44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x0000000000000000000000aea37855eacb4b41ca0dbc6c744681f96fe09d9444":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x0000000000000000000000b09b89893fd55223ed2d9c06cda7afef867c744944":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x0000000000000000000000b142159adbfc2690b45da01e49cfa2379ddc270144":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x0000000000000000000000b7efab17215a43983d766114feb69172587a409044":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x0000000000000000000000cded44bd41476a69e8e68ba8286952c414d28af744":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x0000000000000000000000f9e18acc86179925353713a4a5d0e9bf381fbc1744":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
    "0x0000000000000000000005e3c16931c6e578f948231dca609d754c18fc094e58":"0x43b553fae2184b25e76b69a2386bfc9a014486db7da3df75bba9fa2e3eed8aaf",
    "0x0000000000000000000005e3c16931c6e578f948231dca609d754c18fc094e59":"0x063a5f1aab68488a8645fd6a230a27bfe4e8d3393232fe107ba0f68a9bf541ad",
    "0x000000000000000000000a3f28de9682df49f9f393931062c5204c2bc4044e58":"0xa9f94b62067e993f3f02ada1a448c70ae90bdbe4c6b281f8ff16c6f4832e0e9a",
    "0x000000000000000000000a3f28de9682df49f9f393931062c5204c2bc4044e59":"0xba1827531c260b380c776938b9975ac7170a7e822f567660333622ea3e529313",
    "0x000000000000000000000ead6cdb8d214389909a535d4ccc21a393dddba94e58":"0xb624a3fb585a48b4c96e4e6327752b1ba82a90a948f258be380ba17ead7c01f6",
    "0x000000000000000000000ead6cdb8d214389909a535d4ccc21a393dddba94e59":"0xd4ad43d665bb11c50475c058d3aad1ba9a35c0e0c4aa118503bf3ce79609bef6",
    "0x000000000000000000000f96b686b2c57a0f2d571a39eae66d61a74b58704e58":"0x6f2d48a825fb4fb82dc1e08e2fc9fc1fd82628548d0b544df15a44a5637a642d",
    "0x000000000000000000000f96b686b2c57a0f2d571a39eae66d61a74b58704e59":"0xc4be69768c105c6fc5d024c1166b6ffaae1d4d20325f988d5ef9b97e073cd7c9",
    "0x00000000000000000000328b4bb56a750ea86bd52329a3e368d051699bb24e58":"0xac0eac73f5ffc009085dccfe22b9f749e9982cc416170fb0f219411049fb8b21",
    "0x00000000000000000000328b4bb56a750ea86bd52329a3e368d051699bb24e59":"0xf774cccbe251d9db0437fcd0d0fefb52c7c933dfaafbabb5567ac8389ff491d1",
    "0x0000000000000000000055fbba0496ef137be57d4c179a3a74c4d0c643be4e58":"0x8ce7defe2dde8297f7b55dd9ba8c5e13e0274371b716250ea0dd725974fa076c",
    "0x0000000000000000000055fbba0496ef137be57d4c179a3a74c4d0c643be4e59":"0xa379fc7226789a91678f4e38f8f60f8e6405ec9539cab77d4822614e80f743cf",
    "0x000000000000000000006a3217d128a76e4777403e092bde8362d41177734e58":"0xdbf8dcc4c82eb2ea2e1350b0ea94c7e29f5be609736b91f0faf334851d18f8de",
    "0x000000000000000000006a3217d128a76e4777403e092bde8362d41177734e59":"0x1a518def870c774649db443fbce5f72246e1c6bc4a901ef33429fdc3244a93b3",
    "0x000000000000000000007823a1bea7aca2f902b87effdd4da9a7ef1fc5fb4e58":"0xdf57387d6505d0f71d7000da9642cf16d44feb7fcaa5f3a8a7d9fa58b6cbb6d3",
    "0x000000000000000000007823a1bea7aca2f902b87effdd4da9a7ef1fc5fb4e59":"0x3d145746d4fb544c049d3ff9b534bf9245a5b8052231c51695fd298032bd4a79",
    "0x000000000000000000007eb0bcd103810a6bf463d6d230ebcacc85157d194e58":"0xbc5e761c9d0ba42f22433be14973b399662456763f033a4cdbb8ec37b8026652",
    "0x000000000000000000007eb0bcd103810a6bf463d6d230ebcacc85157d194e59":"0x6e6c56f92d0591825c7d644e487fcee828d537c58ce583a72578309ec6ebbd39",
    "0x000000000000000000008c3d1a9504a36d49003f1652fadb9f06c32a44084e58":"0x80606b6c1eecb8ce91ca8a49a5a183aa42f335eb0d8628824e715571c1f9d1d7",
    "0x000000000000000000008c3d1a9504a36d49003f1652fadb9f06c32a44084e59":"0x57911b80ebc3afab06647da228f36ecf1c39cb561ef7684467c882212ce55cdb",
    "0x00000000000000000000915b5295dde0cebb11c6cb25828b546a9b2c93694e58":"0x9f237f9842f70b0417d2c25ce987248c991310b2bd4034e300a6eec46b517bd8",
    "0x00000000000000000000915b5295dde0cebb11c6cb25828b546a9b2c93694e59":"0xc4f7f31f157128d0732786181a481bcf725c41a655bdcce282a4bc95638d9aae",
    "0x0000000000000000000092e0fea9aba517398c2f0dd628f8cfc7e32ba9844e58":"0x68315573b123b44367f9fefcce38c4d5c4d5d2caf04158a9068de2060380b81f",
    "0x0000000000000000000092e0fea9aba517398c2f0dd628f8cfc7e32ba9844e59":"0x26b220543de7402745160141f932012a792722fd4dd2a7a2751771097eeef5f2",
    "0x000000000000000000009cde10b889fca53c0a560b90b3cb21c2fc965d2b4e58":"0x14f62dfd8826734fe75120849e11614b0763bc584fba4135c2f32b19501525d5",
    "0x000000000000000000009cde10b889fca53c0a560b90b3cb21c2fc965d2b4e59":"0x5d217742893801ecc871023fc42ed7e80196357fb5b1f762d181e827e626637d",
    "0x00000000000000000000a121e6670439ba37e7244d4eb18e42bd6724ef0f4e58":"0x23353a739cc08d5d9b52809562f452e375a4ebdab142cff5369e9fe27d3b1bfe",
    "0x00000000000000000000a121e6670439ba37e7244d4eb18e42bd6724ef0f4e59":"0xd4c03517f7f4815b0f22cb1e4af6bd4c08e3fc80376135f2473983336f04f0b1",
    "0x00000000000000000000aea37855eacb4b41ca0dbc6c744681f96fe09d944e58":"0xb00e07027e771793bd142323570686667e8ae7770afc1db5fe249784b890f277",
    "0x00000000000000000000aea37855eacb4b41ca0dbc6c744681f96fe09d944e59":"0x34517232f764af201ac04451cf314db10d5bc1eb7c50cdd035fdda7b29068ba1",
    "0x00000000000000000000b09b89893fd55223ed2d9c06cda7afef867c74494e58":"0xa71a2a108f5da1fc3841d9f09b2e6a29b9720b225c65d63e97ac4c1f2f05367c",
    "0x00000000000000000000b09b89893fd55223ed2d9c06cda7afef867c74494e59":"0x7bab15e05d3380d9f66cde1b0b216c2846db331de8dad9c1937105297f103427",
    "0x00000000000000000000b142159adbfc2690b45da01e49cfa2379ddc27014e58":"0x139839c1a4fa6a9fef82caf050fde90cab8a3149312d40ca964d9c88713a0003",
    "0x00000000000000000000b142159adbfc2690b45da01e49cfa2379ddc27014e59":"0x3ae0521a67f0c9344a6eb40d3c09c1a6344793266807b4051e8bc9c500031702",
    "0x00000000000000000000b7efab17215a43983d766114feb69172587a40904e58":"0x1e5c7c07fcf6608e84c970b172b865284832a01d09e80e80a035809ed7a8aa01",
    "0x00000000000000000000b7efab17215a43983d766114feb69172587a40904e59":"0xe85b2156aab81d4ba9e08e7312acc6751614a16ad0deca792572959ff74366c7",
    "0x00000000000000000000cded44bd41476a69e8e68ba8286952c414d28af74e58":"0x25ea3bca7679192612aed14d5e83a4f2a30824ff2af705d2d7c6795470f9cbbc",
    "0x00000000000000000000cded44bd41476a69e8e68ba8286952c414d28af74e59":"0x258d9b102a726c3982cda6c4732ba3715551b6fbf9c0ae4ddca4a6c80bc4bbe9",
    "0x00000000000000000000f9e18acc86179925353713a4a5d0e9bf381fbc174e58":"0x037a7c7a2acbfbde6d2ad243c9ef929a3f782c3aff1839e7c40dba0534fbe65c",
    "0x00000000000000000000f9e18acc86179925353713a4a5d0e9bf381fbc174e59":"0x5a82dba58e1f8b0dd767e87da0e9f95e3c566279    422a215e0b06ddca411ac361"
    },
    "balance":"3000000000000"},

	"0x0ead6cdb8d214389909a535d4ccc21a393dddba9":{"balance":"30000000000000000000000000000000"},
	"0x6a3217d128a76e4777403e092bde8362d4117773":{"balance":"30000000000000000000000000000000"},
	"0x0a3f28de9682df49f9f393931062c5204c2bc404":{"balance":"30000000000000000000000000000000"},
	"0x8c3d1a9504a36d49003f1652fadb9f06c32a4408":{"balance":"30000000000000000000000000000000"},
	"0x05e3c16931c6e578f948231dca609d754c18fc09":{"balance":"30000000000000000000000000000000"},
	"0x55fbba0496ef137be57d4c179a3a74c4d0c643be":{"balance":"30000000000000000000000000000000"},
	"0x915b5295dde0cebb11c6cb25828b546a9b2c9369":{"balance":"30000000000000000000000000000000"},
	"0x92e0fea9aba517398c2f0dd628f8cfc7e32ba984":{"balance":"30000000000000000000000000000000"},
	"0x7eb0bcd103810a6bf463d6d230ebcacc85157d19":{"balance":"30000000000000000000000000000000"},
	"0xcded44bd41476a69e8e68ba8286952c414d28af7":{"balance":"30000000000000000000000000000000"},
	"0x9cde10b889fca53c0a560b90b3cb21c2fc965d2b":{"balance":"30000000000000000000000000000000"},
	"0x7823a1bea7aca2f902b87effdd4da9a7ef1fc5fb":{"balance":"30000000000000000000000000000000"},
	"0x992fcd5f39a298e58776a87441f5ee3319a101a0":{"balance":"30000000000000000000000000000000"},
	"0x0f96b686b2c57a0f2d571a39eae66d61a74b5870":{"balance":"30000000000000000000000000000000"},
	"0x328b4bb56a750ea86bd52329a3e368d051699bb2":{"balance":"30000000000000000000000000000000"},
	"0xb09b89893fd55223ed2d9c06cda7afef867c7449":{"balance":"30000000000000000000000000000000"},
	"0xaea37855eacb4b41ca0dbc6c744681f96fe09d94":{"balance":"30000000000000000000000000000000"},
	"0xb142159adbfc2690b45da01e49cfa2379ddc2701":{"balance":"30000000000000000000000000000000"},
	"0xb7efab17215a43983d766114feb69172587a4090":{"balance":"30000000000000000000000000000000"},
	"0xd51ef175736ec808f775c6b687e5be224f1de458":{"balance":"30000000000000000000000000000000"}
    },   

		"coinbase": "0x0000000000000000000000000000000000000000", 
	  "leader":"0x0ead6cdb8d214389909a535d4ccc21a393dddba9", 
    "difficulty": "0x400", 
    "extraData": "0x1000",
	   
	   "elect":[
	   {"Account":"0x0ead6cdb8d214389909a535d4ccc21a393dddba9","Stock":1,"Type":2},
	   {"Account":"0x6a3217d128a76e4777403e092bde8362d4117773","Stock":1,"Type":2},
	   {"Account":"0x8c3d1a9504a36d49003f1652fadb9f06c32a4408","Stock":1,"Type":2},
	   {"Account":"0x05e3c16931c6e578f948231dca609d754c18fc09","Stock":1,"Type":2},
	   {"Account":"0x55fbba0496ef137be57d4c179a3a74c4d0c643be","Stock":1,"Type":2},
	   {"Account":"0x915b5295dde0cebb11c6cb25828b546a9b2c9369","Stock":1,"Type":2},
	   {"Account":"0x92e0fea9aba517398c2f0dd628f8cfc7e32ba984","Stock":1,"Type":2},
	   {"Account":"0x7eb0bcd103810a6bf463d6d230ebcacc85157d19","Stock":1,"Type":2},
	   {"Account":"0xcded44bd41476a69e8e68ba8286952c414d28af7","Stock":1,"Type":2},
	   {"Account":"0x9cde10b889fca53c0a560b90b3cb21c2fc965d2b","Stock":1,"Type":2},
	   {"Account":"0x7823a1bea7aca2f902b87effdd4da9a7ef1fc5fb","Stock":1,"Type":2},
	   {"Account":"0x0a3f28de9682df49f9f393931062c5204c2bc404","Stock":1,"Type":0}],
	   
	   "version": "0x5000", 
     "gasLimit": "0xEE00000",
	   
	   "nettopology":{"Type":0,
						"NetTopologyData":[
						{"Account":"0x0ead6cdb8d214389909a535d4ccc21a393dddba9","Position":8192},
						{"Account":"0x6a3217d128a76e4777403e092bde8362d4117773","Position":8193},
						{"Account":"0x8c3d1a9504a36d49003f1652fadb9f06c32a4408","Position":8194},
						{"Account":"0x05e3c16931c6e578f948231dca609d754c18fc09","Position":8195},
						{"Account":"0x55fbba0496ef137be57d4c179a3a74c4d0c643be","Position":8196},
						{"Account":"0x915b5295dde0cebb11c6cb25828b546a9b2c9369","Position":8197},
						{"Account":"0x92e0fea9aba517398c2f0dd628f8cfc7e32ba984","Position":8198},
						{"Account":"0x7eb0bcd103810a6bf463d6d230ebcacc85157d19","Position":8199},
						{"Account":"0xcded44bd41476a69e8e68ba8286952c414d28af7","Position":8200},
						{"Account":"0x9cde10b889fca53c0a560b90b3cb21c2fc965d2b","Position":8201},
						{"Account":"0x7823a1bea7aca2f902b87effdd4da9a7ef1fc5fb","Position":8202},
						{"Account":"0x0a3f28de9682df49f9f393931062c5204c2bc404","Position":0}]},
						
	   "signatures":[
	   [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,100],
	   [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,100],
	   [0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,100]
	   ],
       "nonce": "0x0000000000000050",
       "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
       "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
       "timestamp": "0x00"
    }

Path: can be located anywhere. You can direct to this place via --datadir when executing gman


In the file genesis.json :

- smartcontract-candidate list: includes the NodeID, accounts, deposits, online time, etc of current candidates

How candidate list is generated:

Step 1: Create a txt file named genesis.txt under go-matrix/core/core, where you should specify the nodeid, account and role of each node

Step 2: Run the file create_genesis_data_test.go under go-matrix/core

Step 3: Then a txt file named saveGenesis.txt will be generated, where there's a candidate list. You can copy these to the genesis.json


- election info: includes the account address, stakes and ID types of all miners as well as validators during the first election cycle

elect PART：

type=2 indicates a validator; type=0 indicates a miner

- topology: includes the accounts and addresses of all miners as well as validators during the first election cycle （the account information shall be consistent with election information）

nettopology PART: 

The role in the topology can be judged by its position information, where 8XXX indicates a validator，and 0-1-2-3... indicates a miner

The above fields should be fine for most purposes, although we'd recommend changing the `nonce` to
some random value so you prevent unknown remote nodes from being able to connect to you. If you'd
like to pre-fund some accounts for easier testing, you can populate the `alloc` field with account
configs:


    "alloc": {
     "0x0000000000000000000000000000000000000001": {"balance": "111111111"},
     "0x0000000000000000000000000000000000000002": {"balance": "222222222"}
    }



#### Common Profile (name it 'man.json', for example)
   
    {
	"BootNode":[
	"enode://b624a3fb585a48b4c96e4e6327752b1ba82a90a948f258be380ba17ead7c01f6d4ad43d665bb11c50475c058d3aad1ba9a35c0e0c4aa118503bf3ce79609bef6@192.168.3.162:30303"
	],
	"BroadNode":[
	{"NodeID":"4b2f638f46c7ae5b1564ca7015d716621848a0d9be66f1d1e91d566d2a70eedc2f11e92b743acb8d97dec3fb412c1b2f66afd7fbb9399d4fb2423619eaa514c7","Address":"0x992fcd5f39a298e58776a87441f5ee3319a101a0"}
	]
    }

Please note: You need to create a directory named chaindata first. This common profile should be put under this directory. Otherwise, you can't start gman due to the inability to read the profile


- BootNode: includes NodeID and IP of all boot nodes for node discovery
- BroadNode: includes NodeID and accounts of all broadcast nodes for broadcast service

With the genesis state defined in the above JSON file, you'll need to initialize **every** gman node
with it prior to starting it up to ensure all blockchain parameters are correctly set:

```
./gman --datadir ./chaindata/ init ./root/genesis.json
```

#### Starting up your member nodes

step 1: With the bootnode operational and externally reachable (you can try `telnet <ip> <port>` to ensure
it's indeed reachable), start every subsequent gman node pointed to the bootnode for peer discovery
via the `--bootnodes` flag. It will probably also be desirable to keep the data directory of your
private network separated, so do also specify a custom `--datadir` flag.

```
$ gman --datadir=path/to/custom/data/folder --bootnodes=<bootnode-enode-url-from-above>
```

step 2: start up all validator nodes （see genesis.json for configurations of validators）

step 3: start up all broadcast nodes (see man.json for configurations of broadcast nodes)

step 4: start up all miner nodes (see genesis.json for configurations of validators)


#### Execute the following command:

./gman --identity "MyNodeName" --datadir ./chaindata/ --rpc --rpcaddr 0.0.0.0 --rpccorsdomain "*" --networkid 1 --password ./chaindata/password.txt

Note:  password.txt contains your password of the wallet, which can also be placed under /chaindata.



#### License
Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors

The go-matrix-ethereum library (i.e. all code outside of the `cmd` directory) is licensed under MIT.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
