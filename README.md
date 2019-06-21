# go-matrix
------
## About
Following the successful deployment of the Matrix AI Network’s first major update in May, the Matrix team is happy to announce that a second major update is ready! It brings with it two major new features. They are: Fixed/Flexible Stakes and Verification Masternode Pools. This article will outline Fixed and Flexible Staking.

### Introducing Fixed and Flexible Staking
Once this update goes live, the Matrix AI Network will support two different stake types: Fixed Stakes and Flexible Stakes. Both have slightly different behaviors, rules and attributes.


|Flexible Stake||
|--------------|---------------------------------------------|
|Minimum Stake |100 MAN tokens each time you add to your stake|
|Lockup Time   | None|
|Stake Reward  | None. Flexible stakes are given a “Stake Weight” of 1 when calculating stake rew|

|Fixed Stake  ||
|---------------|---------------------------------------------|
|Minimum Stake  | 2000 MAN tokens each time you add to your stake|
|Lockup Time    | Fixed. When staking tokens, you must select a 1-, 3-, 6-, or 12-month lockup period.Prior to the end of your stake period, you must initiate unstake procedures or else it will be automatically renewed with the same lock-up period|
|Stake Reward   | Fixed stakes are given increasingly heavy “Stake Weight” as the lock-up time increases|


### Other Notes and Updates
#### You can combine fixed and flexible stakes.
#### If you have a fixed stake, you must initiate unstake procedures prior to the end of your stake period or else it will be automatically renewed at the same lock-up period.
#### If you unstake a fixed stake, you must wait at least 2 hours before being able to transfer the funds to your personal wallet. You will not be able to transfer the funds until your lockup time is expired.
#### If you unstake a flexible stake, you must wait at least 7 days before being able to transfer the funds your personal wallet.


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

- Tag: v1.1.5

Building gman requires both a Go (version 1.7 or later) and a C compiler. You can install them using your favourite package manager. Once the dependencies are installed, run your 'make gman' command 

You can also obtain our compiled gman from github [https://github.com/MatrixAINetwork/GMAN_CLIENT/tree/master/MAINNET/0620](https://github.com/MatrixAINetwork/GMAN_CLIENT/tree/master/MAINNET/0620)



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
