# go-matrix
------

Matrix Mainnet Update Notice


## Successful updates
We’ve made the following improvements in the latest update
1.	Updated the mining process
2.	Updated the miner election algorithm
3.	Updated the difficulty adjustment algorithm
4.	Adjusted the award distribution frequency and one time payout, while maintaining the cap of total available rewards.
5.	added compute penalty process.
6.	storage optimization 


### mining process
Block creation previously occurred with POS and POW running in serial execution. This lead to periods of time when miners were latent and compute was underutilized. This update introduces AI mining while increasing the productivity and utilization of compute.

During each broadcast period the 99 blocks other than the broadcast block will be divided into sets of 33, with each set representing 3 consecutive blocks. In each set the first block will be the AI block, the second will be a POS block, and the third will be a POW block.

**block**	**AI mining result**	**POW mining result**	**POS result**
**AI block**	√	×  	√
**POW block**	×  	×  	√
**POS block**	×  	√	√
			
Note: After changing rounds, and AI block doesn’t contain the AI mining results. Each election cycle consists of 300 blocks, comprised of 98 AI results, 99 POW results, and 297 POS results)


### AI block creation

-	In each cycle miners conduct AI mining based on the provided information from the previous cycle and the results are sent for validation.
-	The validation leader uses the AI results to create a validation block which will go through the POS consensus process. 
-	Once all validators have completed the POS consensus, they send the mining request for the new cycle to mining nodes to complete the AI block.
-	Miners can then begin POW mining once they receive the mining request for the election cycle. 


#### POW block creation

-	The validation leader produces the block initiating the POS consensus process.
-	Once the POS is complete the validator waits for the cycles’ POW result.
-	After the cycles’ POW is complete, the POS and POW information is combined and the POW block is issued.


#### Mining process

-	A miner begins POW mining once the information from the mining request is received.
-	POW mining is comprised of X11 compute and SM3 compute. X11 and SM3 mining are executed in parallel.
-	If the results of the X11 compute process satisfy the minimum difficulty requirements, the compute results can be sent to the validator and can continue to do POW mining. 


1.	The difficulty of finding the solution is the core characteristic of traditional HASH algorithms, however the validation is exceedingly simple. Therefore seamlessly implementing an AI mining algorithm needs to preserve the difficulty of finding the solution and ease of validation. Among AI algorithms, adversarial networks possess these characteristics, which is why we chose it as Matrix’s next mining algorithm. The nature of the algorithm is as follows:


(1)	A background pattern is generated based on the characteristics of the user.
(2)	A pattern cluster is presented according to the parameter configurations of the block on the data chain, and generate the  final image. 
(3)	Then detect objects in the image, according to the configuration provided the data chain. The detection results must satisfy a given requirements, thereby avoiding mistakes or generating duplicate patterns. 
(4)	Images must be recognized among randomly changing RGB pixel values within a specified region, with different recognized and non-recognized results.
(5)	Users detect changes in value from the original RGB value reporting the smallest differential in Euclidean distance. The purpose of the algorithm is to find the smallest differential to create a different recognizable image.  


2.	Implementing the updated AI algorithm.
GPUs will be needed to support the massive workloads to run GANs. The block creation time is too long in the current CPU model slowing down the entire network. For this reason Matrix will be updating our mainnet.  The update will include the following:
1)	Batch uploads of AI photos                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   
2)	Hash computing for images
Features implemented on David’s test chain include:
1) Batch uploads of AI photos,
2) AI image recognition (Yolo V3)
Features still awaiting testing on David’s test net.
1) The inconsistency of CPU and GPU floating point calculations
2) the issue that it takes longer than expected to validate of AI images


3.	The next phase of AI engineering improvements
Since these AI algorithms are heavily dependent on the size of the image library, the selection and updating of the AI image library is very important. We are currently preparing to use the following approaches for completion：
1)	Divide libraries of 16 images to 2 to the 16th power (2^16) then randomly select 16 images.
2)	Generating the background image will use fractal technology selecting the outer edge of a Mandelbrot set based on the prior hash and VFR results. Each topology is generated from the formula: Zn+1=(Zn) ^2+C. Each nonlinear iteration of the formula Zn+1=(Zn)^2+C, will preserve a finite number of copies for set C, creating a Mandelbrot set.
3)	Automatically batch updating images libraries within a finite specifications and time period, it will also set the numbering order MD5 validation, which will be recorded on the blockchain transaction records.
4)	Automatic updating of the image recognition parameters.
 

4.	Resolving the compatibility of using CPU and GPUs
Since the floating points definition and algorithm GPUs use is substantially different from the traditional IEEE 754 floating point compute, all of the algorithm libraries will need to undergo compatibility standardization to maintain the consistency of AI recognition results.


### Compute penalty

#### Background

At present a number of miners elected on the mainnet, have only undergone staking, but have not activated their nodes and were still receiving mining rewards without supporting the development of Matrix mainnet. This has been a rather large impact on the compute available across the mining network of the mainnet.


#### Purpose
To have each staked miner contribute actual compute.

#### Solution
Once a node is elected to be a mining node, for each POW cycle they will need to simultaneously compute a minimum low difficulty POW result, which is reported to the validator. The validator will publish the minimum compute results to the blockhead. Before miners are elected in each election cycle, there will be an assessment of the minimum compute requirement results, and node that didn’t qualify will be added to a blacklist. 
Testing principle: During an election cycle nodes that failed to pass the test twice will be considered unfit.


Implementing penalties:
Blacklisted nodes will
1. be denied mining rewards from the current cycle.
2. lose eligibility to participate in multiple subsequent mining cycles (For details on subsequent cycles “mining masternode election algorithm”, for this update the suspension time will be one cycle.)


### Mining masternode election algorithm
#### Background
A random election method was used on the earlier data chain update to generate mining masternodes. Should the number of staked miners exceed 1000, users will have to await a sufficient amount of time before being elected as a miner.
#### Purpose
The probability of being elected for a mining cycle in a given time frame is now much greater without altering the overall probability of getting elected. The probability under the earlier edition can be problematic, for instance: for a staked account that would be elected 10 times in a given month, one would happen in the first 20 days, and the remaining 9 times would happen in the following 10 days. This update will make a more even distribution of being elected once every 3 days. 

    
#### Solution offered by the current update
A round-robin algorithm will be used for this update. The fundamental rationale is getting a snapshot of staked nodes during each election cycle. The mining masternodes for the next election cycle will be chosen from this random snapshot, then the elected nodes will be removed from the snapshot until there are no longer qualified nodes in the snapshot and the cycle is completed. Once the available miners for a given round-robin is lower than the requisite number, the round is completed and a new round starts. The remaining qualified nodes are elected and a new round begins (ineligible nodes in the snapshot are ignored.)


To prevent elected nodes from later dropping out and impacting the compute pool, nodes that have dropped out or have compute penalties will be automatically filtered out.

The data chain currently has 24 cycles per day, with 32 nodes participating per cycle, meaning 768 nodes will be elected each day. Assuming 1024 staked nodes, each node would be elected once every 1.34 days. Accounting for block creation time you could round that to once every 1.5 days, or twice in a 3 day period.


In order to support this average election time, we’ve strengthened the function to adjust the upper limit of mining masternodes. When there are more staked miners, more mining masternodes can be supported, when staking is low the mining masternodes will be less. This mechanism will work according to the following principles
1.	Regardless of the number of staked nodes, 32 is the minimum number of masternodes.
2.	After reaching 1024 staked nodes, for every additional 64 staked nodes, there will be an additional 2 mining masternodes.
For example at 1023 staked nodes there will be 32 mining masternodes, and will remain that number through 1087 staked nodes. Once the staked nodes reach 1088, there would be a total of 34 mining masternodes and remain so through 1151 staked nodes. Once the staked nodes reach 1152, there would be a total of 36 mining masternodes.


### Difficulty adjustment algorithm
This update will introduce and entirely new block creation process. The block difficulty for the entire network will increase considerably. The anticipated mining compute will change quickly, due to the unique Matrix characteristic masternodes participating in POW in mining network election cycles. The newly adjusted difficulty algorithm is suited for the following dramatic changes in compute:
1.	Difficulty for each election cycle will have a base difficulty
2.	The first N POW blocks in rapid succession will increase the difficulty
3.	Subsequently a “weighted moving average” method will be used to adjust the difficulty.


Currently each POW mining process lasts for the creation of 3 blocks, the mainnet hopes to achieve a blocktime of 12 seconds, creating 3 blocks in 36 seconds. Due to additional time delays for the validation and sending blocks, there is an additional 1 second, so the actual POW mining time is 33 seconds.


In order to prevent faults in powerful nodes, and cause problems for the data chain, if elected nodes time out during mining, mining results will be taken from “mining nodes in the Matrix foundation” (the nodes from the foundation do not participate in the distribution of rewards). If a miner times out on the previous POW block, the difficulty algorithm will be adjusted downward. With this update the mining time out window will be set at 100 seconds.


When difficulty allows miners to create blocks quickly in close succession even outpacing the POS process, the difficult index will increase. According to the situation of the mainnet, the POS process takes approximately 46 seconds, the configuration for the fastest blocktime will be 7 seconds. When the POW mining time is lower than 3 times the lowest blocktime, or 21 seconds, the difficult will quickly increase. If POW mining time is higher than 3 times the lowest blocktime, the system will use an iterate process to linearly adjust the difficulty level in relation to blocktime. 


### Storage optimization 
The present the storage space for the data chain is fairly large, giving nodes supporting operation of the data chain a considerable storage burden. 
Analysis of the data chain found the following factors impacting the storage burden.

1. Data chain rewards transactions: each block saves dozens of reward transaction records (undergoing one to many distribution) somewhat impacting the state tree. Currently the data chain has a high volume of federated staked accounts, and each time rewards are distributed there is a cascade of redistributions which increase the storage burden.
2. Blocks on the prior data chain recorded interest rewards, increasing the storage burden.
3. The prior data chain supports multiple currency types with each type of currency adding storage requirements.


#### Optimization methods

1.	Increased storage capacity for multi-currency, and only storing changes in the balance, and reporting the prior value if there is no change.
2.	Mining participation and rewards will only be recorded once per cycle.
3.	Validator participation and rewards will only be recorded once per cycle.
4.	Staking information will go from being calculated and stored once a block to once per broadcast cycle


Testing from our lab stores that for 3000 staked accounts, 212000 have balance accounts, and 7 currency types （same with the data chain）, 4 alliance stakes, each alliance stake has 300 participating accounts” . After editing blocks will be reduced to about 10% of the previous amount.
## Rewards


### Fixed mining block rewards

1.	In the previous edition block mining rewards were issued once per block. In the update distribution will be once per election  cycle. The cycle payout will be 297 times greater than the single block payout (broadcast nodes don’t participate), the current number is 570.25 MAN.
2.	An individual miner participation reward is equal to the total miner participation count / miner masternode count.
3.	In an election cycle, blacklisted nodes don’t receive mining participation awards.
4.	The mining participation rewards on the update will be reduced from 50% to 40%.
5.	Miner POW rewards will change from being issued each block to issuing three times the volume (5.6MAN) every third block.
6.	A fixed 10% of block rewards will be distributed for AI rewards every third block. At the current rate 1.44 MAN (compute method: 10% of the individual block mining rewards times 3).


The change in mining reward distribution and distribution cycle is as follows:

| reward types  | former mainnet reward percentages | updated mainnet reward percentages | former mainnet reward distribution cycle | updated mainnet reward distribution cycle | former mainnet payout amounts | updated mainnet payout amounts |
| ----------- | ---- | ---- | -------- | -------- | -------- | -------- | 
| pow miner rewards | 40% | 40%  | 1 block | 3 blocks | 1.92MAN| 5.76MAN |
| AI miner rewards  | none | 10%  | none | 3 blocks | none | 1.44MAN |
| participation rewards | 50% | 40%  | 1 block | 300 blocks | 2.4MAN | 570.24MAN |
| Foundation redards | 10% | 10%  | 1 block | 1 block | 0.48MAN |  0.48MAN |

（None：1 percent represents total mining awards ratio） 


### validator fixed block rewards

1.	The earlier mainnet validator participation rewards were released once per block, with the update they will be issued once per election cycle. The cycle payout will be 297 times greater than the single block payout (broadcast nodes not participating), at the present rates equal to 1544.4MAN.
2.	The distribution for validator fixed block participation rewards will not change. A difference in the update is the distribution rate is that formerly the participation rewards were issued once per block, and this time participation rewards accumulate in the state tree and are distributed at the beginning of the next election cycle.


| reward type   | former mainnet reward percentage | updated mainnet reward percentage | former mainnet distribution rate | updated mainnet distribution rate | former mainnet distribution amount | updated mainnet distribution amount |
| ---------- | ------------ | ------------ | ------------ | ------------ | ------------ | ------------ |
| leader reward | 25%          | 25%          | 1 block         | 1 block        | 3MAN        | 3MAN        |
| participation reward   | 65%          | 65%          | 1 block         | 300 blocks     |5.2MAN      | 1544.4MAN       |
| Foundation reward | 10%          | 10%          | 1 block         | 1 block       | 0.8MAN        | 0.8MAN       |


### Transaction fee reward

1. In the old version participation reward is distributed to each block only once. After the update participation reward will be distributed after each election cycle.
2. The way transaction fee reward is calculated hasn’t changed. However, in the old version participation reward is distributed along with the fixed reward of each block; after the update, reward will be saved in the status tree and distributed by the time the first block is created in the following election cycle.



| reward type   | former mainnet reward percentage  | updated mainnet reward percentage | former mainnet distribution rate | updated mainnet distribution rate
| ---------- | ------------ | ------------ | ------------ | ------------ |
| leader reward | 40%          | 40%          |1 block         |1 block         |
| participation reward   | 60%          | 60%          | 1 block         |300 blocks         |


### interest reward

The current mainnet interest rewards are calculated during each block and saved into the state tree. In order to optimize storage, in the update interest rewards will be calculated every 100 blocks. The interest algorithm will be unchanged.


## Bug fixes
This update fixed the following bugs.
1.	Repaired part of the RPC connector.
2.	Repaired the bug with VRF not participating in a signature.
3.	Repaired the bug for not duplicating capture of the same height in the Fetch module.
4.	Repaired the PEER link problem.


### Repairing the RPC connector
Previous some of the API parameters for the RPC connector didn’t support MAN addresses and lacked some functionality. This update fixed the following API functions

1. man_newFilter
2. man_getFilterLogs
3. man_coinbase
4. man_getLogs
5. man_getFilterChanges
6. man_getEntrustFrom
7. man_getEntrustFromByTime
8. man_getAuthFrom
9. man_getAuthFromByTime
10. man_getAuthGasAddress
11. man_getMatrixCoinConfig
12. trace_traceTransaction

### Repaired the bug with VRF not participating in a signature.
In the earlier data chain the VRF did not participate in the POS signature, raising the possibility of the leader forging the VRF. After the update the VRF content will participate in the POS signature.

   
### Repaired the bug for not duplicating capture of the same height in the Fetch module.

When the Fetch module doesn’t duplicate capture of the same height there was a very low percentage for discrepancy between a leader, multiple broadcast and validator consensus, causing block stalling. This update fixes that.  


### Repaired the PEER link problem.

The current peer link is based on a tiered mechanism raising the possibility that if a higher level node is offline then lower level peers may be unable to connect. This could lead to an inability to download data causing the network and real world situation to get out of sync. This is now updated.



Update methodology:
1.	When a peer is unable to connect to a node on a higher tier, there’s an increase in dynamic proactive peers attempts to sync the data.
2.	An increase in the dynamic peer link monitoring mechanism which periodically reconnects a set number of peers. This helps preclude a passively disconnected peer from being unable to connect.
3.	Increase the mechanism for dynamically connecting nodes, periodically reducing the peer node dynamic connection to half, then after the next dynamic search re-establishing connections. This mechanism is primarily to prevent peer nodes from falling into a closed isolated network. Dynamically reestablishing connections will accomplish this.
4.	Increase information transfer in the tiered mechanism between dynamic connections and static connections. While preserving the stability of static links in the tiered mechanism, dynamic link nodes need to account for data from the static links, which requires the peer for dynamic links to convert to static links.
Strengthening the robustness and flexibility of node links to avoid extreme situations of an inability to connect to the network, thereby enhancing reliability. 


## improvements
This update has the following improvements for ease of use and efficiency:

1.	Optimized the generation process for the signature document
2.	Optimized on-chain speed


### Optimized the generation process for the signature document

Once a signature and address document are saved, the program will not regenerate this document. If there is a change to the user’s startup command account, leading to an incompatibility with the system, it could cause an issue with the node being unable to connect.

This update improved the logic for generating these documents. The document is regenerated each time at startup, enhancing ease of use.


### Optimized on-chain speed 
Our analysis finds that currently the on-chain process calculating interest takes a significant amount of time when accessing the state tree and stakes. This update will change the interest calculation from once per block to once each broadcast cycle, increasing on-chain speed. Furthermore, some of the current code for reading the state tree of the election modules is inefficient. This update enhances the efficiency of code for reading the tree state.



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

- Tag: v1.1.6

Building gman requires both a Go (version 1.7 or later) and a C compiler. You can install them using your favourite package manager. Once the dependencies are installed, run your 'make gman' command 

You can also obtain our compiled gman from github [https://github.com/MatrixAINetwork/GMAN_CLIENT/tree/master/MAINNET/1022](https://github.com/MatrixAINetwork/GMAN_CLIENT/tree/master/MAINNET/1022)



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
