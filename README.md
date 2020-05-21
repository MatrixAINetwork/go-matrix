# go-matrix
------

Matrix Mainnet Update Notice


## Successful updates
This patch will bring the following changes to Matrix AI Network:
1. Adjustments to the penalty policy;
2. Adjustments to the difficulty algorithm;
3. More search space for POW mining;
4. Bug fixes


### Penalty Policy
In the current version, mining nodes that are elected but fail to report their base compute will get
blacklisted and lose all rewards for that round.
The new patch will raise the penalty for backlisted mining nodes. If blacklisted, your node is also
unable to become a candidate mining Masternode for the next election cycle


### Difficulty Adjustment Algorithm

We have found two flaws in the current version:
-   Difficulty level drops when a new validation leader is elected.
If there is a new validation leader in one mining cycle (the time for generating 3 blocks), this
mining cycle will take longer to finish, and difficulty for mining the following block will lower
down. (The validators take turns to generate blocks, and a validator which has generated a block
is considered the validation leader of that block.) Therefore, the increase in mining time caused
in this way has nothing to do with the actual computing power. (In fact, a problem with one
validator can cause this to happen repeatedly to the point where difficulty adjustment fails. As a
result, the difficulty level will be too low for miners to make full use of their computing power.)
When this happens, it is impossible to know a miner’s actual mining time. Therefore, when
calculating the difficulty level, we’ll use “expected block-generation time” as the intermediate
solution in place of the actual time.
-   In the quick setup and the tracking stages, mining difficulty is slow to stabilise due to lack of
synergy.
Currently, the difficulty adjustment algorithm on the mainnet goes through two stages：
(1) Quick setup: Quickly establish an algorithm for the first n mining cycles. Adjust difficulty
level exponentially to increase the estimation accuracy.
(2) Tracking: Use exponential weighted moving average to track compute change through the
mainnet.
In the current version, when calculating an exponential weighted moving average, the block
information (difficulty level and time) of the quick setup stage will be used at the beginning.
This information is not an accurate reflection of computing power. But since the last block of the
quick setup stage contains relatively accurate information, in the new version, we’ll use this
information instead of trying to get the difficulty level in other ways, when the calculating
exponential weighted moving average.


#### More search space for POW mining

In the current version, the search space for POW mining is only 4 bytes large. When the difficulty
level is too high, an ideal target value may not be found after searching the entire space. In the
new version, we’ll add a 12-byte space, at the initial 12 bytes of the mixDigest field of a block
head

## Bug fixes
The new version will introduce the following bug fixes.
1. In the current version, validators in the mainnet only accept mining results that are one
block higher than the local height. When the POW mining difficulty is too low, this can
cause mining results to arrive two blocks earlier. In consequence, miners with greater
computing power may not see their mining results accepted. This bug will get fixed with the
patch.
2. In the current version, the nonce cannot be 0 for CPU mining. This bug will get fixed with
the patch.
3. In the current version, the P2P module does not lock up visits to the map using certain codes,
causing the system to crash sometimes when there are too many connected nodes. This issue
will get fixed with the patch


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

- Tag: v1.1.7

Building gman requires both a Go (version 1.7 or later) and a C compiler. You can install them using your favourite package manager. Once the dependencies are installed, run your 'make gman' command 

You can also obtain our compiled gman from github [https://github.com/MatrixAINetwork/GMAN_CLIENT/tree/master/MAINNET/20200520]https://github.com/MatrixAINetwork/GMAN_CLIENT/tree/master/MAINNET/20200520)



### Starting up your member nodes (Linux & Mac) - for deposited users

Step 1: Check out what you need to prepare (most of them can be obtaind from go-matrix repository)

    /gman: exe file

    /MANGenesis.json: genesis file

    /chaindata: a folder which you should create

    man.json: common profile which shall be put under /chaindata
	picstore：a folder which shall be put under /chaindata

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
	picstore：a folder which shall be put under /chaindata

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
	picstore：a folder which shall be put under /chaindata

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
	picstore：a folder which shall be put under /chaindata

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
