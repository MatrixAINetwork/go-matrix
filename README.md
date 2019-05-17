# go-matrix
---

## About
It is with great excitement that the Matrix AI Network team releases the first major update to the Matrix AI Network! The primary goal of this update is to increase the stability of the network. This update includes tweaks to several attributes including:

### Verification Masternode Selection
Any Verification Masternode with a stake exceeding 1/19th of the total MAN tokens staked by all Verification Masternodes is guaranteed selection. All other Verification Masternodes are selected according to their stake. The probability of selection can be estimated using a simple formula:

〖Individual Stake〗^1.45/〖Total Verification Masternodes Stake〗^1.45 

The real formula is considerably more complex but yields similar results. The reason we use a more complex algorithm to manage selection is to reduce selection variability in small time periods.

As before, once selected, all Verification Masternodes in a given cycle (300 blocks) serve as the Block Leader, in turn, regardless of stake. The probability of everyone’s starting positions can also be estimated using a simple formula:

〖Individual Stake〗^1.45/〖Total Stake of Selected Verification Masternodes〗^1.45  


### VIP System 
The previous VIP system is completely eliminated and does not impact the selection of Masternodes in any way 

### Voting Timeout
As before, in order to verify and package transactions in a block, at least 13 of the selected Verification Masternodes must agree via a voting mechanism. If a Verification Masternode fails to vote, they timeout and are replaced. The time to timeout is changed from 20 seconds to 40 seconds
### Block Timeout    
If a block leader fails to perform their duties, they will timeout and be replaced. The time to timeout is changed from 40 seconds to 60 seconds. 
### Penalties
Penalties assigned to malicious, mis-deployed and offline Verification Masternodes are increased. Any Verification Masternode that fails to perform its duties for the entirety of a selection cycle will forfeit its next 10 selections.
### Target Block Time
The target block time is increased from 6 seconds to 9 seconds.
### Block Rewards 
Block rewards are increased from 10 MAN tokens to 16 MAN tokens.

### Blockchain Explorer

[http://tom.matrix.io/home](http://tom.matrix.io/home)

### MATRIX WEB WALLET

[https://wallet.matrix.io/](https://wallet.matrix.io/)

### Getting Started
Welcome! This guide is intended to get you running on the MATRIX network. To ensure your client behaves gracefully throughout the setup process, please check your system meets the following requirements:


| OS        | Windows, Linux                               |   |
|-----------|----------------------------------------------|---|
| CPU       | 8 Core (Intel(R) Xeon(R) CPU X5670 @2.93GHz) |   |
| RAM       | 16G                                          |   |
| Free HD   | 300G                                         |   |
| Bandwidth | 20M                                          |   |
|           |                                              |   |

### Build from Source

First of all, you need to clone the source code from MATRIX repository:

Git clone https://github.com/MatrixAINetwork/go-matrix.git, or

wget https://github.com/MatrixAINetwork/go-matrix/archive/master.zip

- Branch: Master

- Tag: v1.1.4

Building gman requires both a Go (version 1.7 or later) and a C compiler. You can install them using your favourite package manager. Once the dependencies are installed, run your 'make gman' command 

You can also obtain our compiled gman from github [https://github.com/MatrixAINetwork/GMAN_CLIENT/tree/master/MAINNET/5.17](https://github.com/MatrixAINetwork/GMAN_CLIENT/tree/master/MAINNET/5.17)



### Starting up your member nodes (Linux & Mac) - for deposited users

Step 1: Check out what you need to prepare (most of them can be obtaind from go-matrix repository)

    /gman: exe file

    /MANGenesis.json: genesis file

    /chaindata: a folder which you should create

    man.json: common profile which shall be put under /chaindata

Step 2: Run Initiate command

    ./gman  --datadir  ./chaindata/   init    ./MANGenesis.json

Step 3: Visit our web wallet to create a new wallet address, and save your keystore file as well as password.

Please refer to [['Guide to Web Wallet']](https://github.com/MatrixAINetwork/MATRIX_docs/blob/master/ENGLISH_DOCS/MATRIX_Web_Wallet/MATRIX%20Online%20Wallet%20Manual.pdf)

Carry out your deposit actions if you want to run for a miner or validator node (you can find steps on the above guide)

Step 4: Copy your keystore file to folder keystore which is generated at Step 2 (/chaindata/keystore)

Step 5: Create a file named signAccount.json under root, and its content is like:

    [
      {
        "Address":" MAN.gQAAHUeTBxvgbzf8tFgUtavDceJP ",
        "Password":" pass123456"
      }

    ]
Then, run: 

    ./gman --datadir ./chaindata aes --aesin ./signAccount.json --aesout entrust.json

Upon the window prompt, you will be asked to set a password (which should contain upper-case letter[s], lower-case letter[s], number[s] and special character[s])

Step 6: Copy the generated entrust.json to root

Step 7: Start gman

    ./gman --datadir ./chaindata --networkid 1 --debug --verbosity 5  --manAddress [your man.address here] --entrust ./entrust.json --gcmode archive --outputinfo 1 --syncmode full 

    for example, 

./gman --datadir ./chaindata --networkid 1 --debug --verbosity 5 --manAddress MAN.gQAAHUeTBxvgbzf8tFgUtavDceJP --entrust ./entrust.json --gcmode archive --outputinfo 1 --syncmode full

In this step, you will need to input the password set in step 5.

Step 8: Run 'Attach': ./gman attach /chaindata/gman.ipc (gman.ipc is generated under /chaindata when starting gman)


### Starting up your member nodes (Linux & Mac) - for non-deposited users

Step 1: Check out what you need to prepare (most of them can be obtaind from go-matrix repository)

    /gman: exe file

    /MANGenesis.json: genesis file

    /chaindata: a folder which you should create

    man.json: common profile which shall be put under /chaindata

Step 2: Run Initiate command

    ./gman  --datadir  ./chaindata/   init    ./MANGenesis.json

Step 3: Start
    ./gman --datadir ./chaindata --networkid 1  --outputinfo 1 --syncmode 'full'

### Starting up your member nodes (Windows) - for deposited users
Step 1: Check out what you need to prepare (most of them can be obtaind from go-matrix repository)

    /gman: exe file

    /MANGenesis.json: genesis file

    /chaindata: a folder which you should create

    man.json: common profile which shall be put under /chaindata

Step 2: Run Initiate command

    gman.exe --datadir chaindata\ init MANGenesis.json

Step 3: Create a file named signAccount.json, whose contents are:

    [
      {
        "Address":"MAN.2skMrkoEkecKjJLPz6qTdi8B3NgjU ",
        "Password":"haolin0123"
      }

    ]

Step 4: Run:

    gman.exe --datadir chaindata aes --aesin signAccount.json --aesout entrust.json

Upon the window prompt, you will be asked to set a password (which should contain upper-case letter[s], lower-case letter[s], number[s] and special character[s])

Step 5: Start gman

    gman --datadir chaindata  --networkid 1 --debug --verbosity 5  --manAddress  MAN.2skMrkoEkecKjJLPz6qTdi8B3NgjU --entrust entrust.json --gcmode archive --outputinfo 1 --syncmode full

In this step, you will need to input the password set in step 5.

Step 8: Open another window

    gman attach ipc:\\.\pipe\gman.ipc 

gman.ipc is generated under /chaindata when starting gman)

### Starting up your member nodes (Windows) - for non-deposited users

Step 1: Check out what you need to prepare (most of them can be obtaind from go-matrix repository)

    /gman: exe file

    /MANGenesis.json: genesis file

    /chaindata: a folder which you should create

    man.json: common profile which shall be put under /chaindata

Step 2: Run Initiate command

    gman.exe --datadir chaindata\ init MANGenesis.json

Step 3: Start gman

    gman --datadir chaindata  --networkid 1 --outputinfo 1 -- syncmode full

### License
Copyright 2018-2019 The MATRIX Authors

The go-matrix library is licensed under MIT.

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
