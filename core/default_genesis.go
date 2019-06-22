// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package core

import (
	"encoding/json"

	"os"
	"reflect"
)

var (
	AllGenesisJson = `{
    "nettopology":{
        "Type":0,
        "NetTopologyData":[
            {
                "Account":"MAN.44EuST4f2vLeEMw2bsMWmBYqLMBhi",
                "Position":8192
            },
            {
                "Account":"MAN.3t9Ser2UjrXRT6erVKytjHtT4ohdX",
                "Position":8193
            },
            {
                "Account":"MAN.2MSWsig8iv45CDTrPn1XMzvZnR52v",
                "Position":8194
            },
            {
                "Account":"MAN.EXnVsEqjyPLHZySL9y53WgXjiFmN",
                "Position":8195
            },
            {
                "Account":"MAN.2QR75feL9KBfaezJwhuM2VCPpqkyT",
                "Position":8196
            },
            {
                "Account":"MAN.95Tro9wLULb6rNWNCT6QDwVBiDds",
                "Position":8197
            },
            {
                "Account":"MAN.375qsgtLc25bJv2Cf9ffnEZPDhyvd",
                "Position":8198
            },
            {
                "Account":"MAN.2RwBWjBykMiGRu7STzYdNzXUMyh8z",
                "Position":8199
            },
            {
                "Account":"MAN.2zjKitA4uydg5kSZLjtSNHzDtx6k8",
                "Position":8200
            },
            {
                "Account":"MAN.4kaowsz37i1WRrPrkg4g8qYQHtJ7",
                "Position":8201
            },
            {
                "Account":"MAN.3B1wx2wo5anTTMjnyWXAA4FQL5opx",
                "Position":8202
            },
            {
                "Account":"MAN.3pm3iXgrY9SYhhc9aS6DQCp4qj66t",
                "Position":8203
            },
            {
                "Account":"MAN.4M8Svax1v6yGitwB4E2ueBtw2TesK",
                "Position":8204
            },
            {
                "Account":"MAN.3abRtpG81TPY8YwenowwgmJ8SuLpy",
                "Position":8205
            },
            {
                "Account":"MAN.2pqHNWtbpKKaLU71RY6TZeaQjjQnR",
                "Position":8206
            },
            {
                "Account":"MAN.3Wf4yzVkbou1bFzkpbJUdu5rqq1se",
                "Position":8207
            },
            {
                "Account":"MAN.U5mNRj4q7jQzbVE4tWcjsvSXTseS",
                "Position":8208
            },
            {
                "Account":"MAN.2ZPbpaCyuEEf1WR3C6dCCFFqnH25J",
                "Position":8209
            },
            {
                "Account":"MAN.3ByNBjw4E7gcxD3uGKtAZYLYUD5xi",
                "Position":8210
            },
            {
                "Account":"MAN.2tSj5kjwwaiTPt4XnHAaEWBYBo9gK",
                "Position":0
            },
            {
                "Account":"MAN.4ZnJrUuM2bfFmdUaivrqGio28hZwd",
                "Position":1
            },
            {
                "Account":"MAN.49fxdGyiWPQ3evpMCNChqvCq3qzMC",
                "Position":2
            },
            {
                "Account":"MAN.3Ugik7ZsLoaNgX51kCJEWz1ZjxQgW",
                "Position":3
            },
            {
                "Account":"MAN.kwPCJkajT2op7rVgYKDqcQu2KEQn",
                "Position":4
            },
            {
                "Account":"MAN.ksFr4mKPfZhm2PrFdEUSoLoDsKAZ",
                "Position":12288
            },
            {
                "Account":"MAN.42bUyszBXL3feeDHztWMiUJCRzBRP",
                "Position":12289
            },
            {
                "Account":"MAN.2NVAVDc7AJGNP3Ghwfv8dz59kUvjM",
                "Position":12290
            },
            {
                "Account":"MAN.e33HPpmmZC98ADkZUXigS1nDFfaA",
                "Position":12291
            }
        ]
    },
    "alloc":{
        "MAN.1111111111111111111B8":{
            "storage":{
                "0x0000000000000000000000000000000000000000000000000000000a444e554d":"0x000000000000000000000000000000000000000000000000000000000000001c",
                "0x0000000000000000000000000000000000000000000a44490000000000000000":"0x000000000000000000000000db588e42D894EDE75000860fC0F4D969393ac514",
                "0x0000000000000000000000000000000000000000000a44490000000000000001":"0x000000000000000000000000Ceda8D254c1925a79fd4cDe842A86e5273400eA4",
                "0x0000000000000000000000000000000000000000000a44490000000000000002":"0x000000000000000000000000611318F1430AD50c540677089053dD7123e3B9b1",
                "0x0000000000000000000000000000000000000000000a44490000000000000003":"0x00000000000000000000000010BeC6Cd8fEf4393b072d1Ddf853E753Cd3D222c",
                "0x0000000000000000000000000000000000000000000a44490000000000000004":"0x00000000000000000000000064C1D85AF78c82Bf7a59030E87114025F76750CE",
                "0x0000000000000000000000000000000000000000000a44490000000000000005":"0x00000000000000000000000009fEEa325F03230969B02cAEC585256D1B26ca30",
                "0x0000000000000000000000000000000000000000000a44490000000000000006":"0x00000000000000000000000097161fE11Dba5a5daA85D1647eA79d72a3cf7385",
                "0x0000000000000000000000000000000000000000000a44490000000000000007":"0x00000000000000000000000066A2F34C7ce30A908F7ed9D5CAE33E939404cc47",
                "0x0000000000000000000000000000000000000000000a44490000000000000008":"0x0000000000000000000000008f3922BebdDECB25D5513E393abBaA00C5dD5Ee1",
                "0x0000000000000000000000000000000000000000000a44490000000000000009":"0x00000000000000000000000004A48442762386D1954895a3d1977457145e8Ca7",
                "0x0000000000000000000000000000000000000000000a4449000000000000000a":"0x0000000000000000000000009bf41d7a8aB4c11a4D0C5c5f451c06bf191E1593",
                "0x0000000000000000000000000000000000000000000a4449000000000000000b":"0x000000000000000000000000CaA9c4aFcA7584462C1663DDE671adAc5532605f",
                "0x0000000000000000000000000000000000000000000a4449000000000000000c":"0x000000000000000000000000f03F2c53784149f95535DBc707277363C6DB45A4",
                "0x0000000000000000000000000000000000000000000a4449000000000000000d":"0x000000000000000000000000b921CBa7b3b0eEB811E462bd096F7432cdCD7745",
                "0x0000000000000000000000000000000000000000000a4449000000000000000e":"0x00000000000000000000000082F984c3D66c4df8F2D68CedA0cd5F2D1897dD2b",
                "0x0000000000000000000000000000000000000000000a4449000000000000000f":"0x000000000000000000000000B442685E3730F1441aE712E4F9E27946b8471902",
                "0x0000000000000000000000000000000000000000000a44490000000000000010":"0x000000000000000000000000218416320cc42b8FCD2a3700bc54d5eD668d6d01",
                "0x0000000000000000000000000000000000000000000a44490000000000000011":"0x0000000000000000000000006fDcfade41750838F48f3Ec1fFDFdB4A3A42f746",
                "0x0000000000000000000000000000000000000000000a44490000000000000012":"0x0000000000000000000000009D22D3DD44b0d7F073C78BCe7c1352175FB7B997",
                "0x0000000000000000000000000000000000000000000a44490000000000000013":"0x000000000000000000000000877192DB751fD2b63C4CF0078221B2d0DAaa01eF",
                "0x0000000000000000000000000000000000000000000a44490000000000000014":"0x000000000000000000000000fFe7c96064D3b4C185B1E4B22D3E692A01Ab2FbE",
                "0x0000000000000000000000000000000000000000000a44490000000000000015":"0x000000000000000000000000e2117fF2836eD3f33B95bb8dBb4ACE9B4DB90e2E",
                "0x0000000000000000000000000000000000000000000a44490000000000000016":"0x000000000000000000000000B1D1CAD653D38B90b586F7755963711f5e37E469",
                "0x0000000000000000000000000000000000000000000a44490000000000000017":"0x0000000000000000000000003660302D6614EF96578EFC3301cc527054Bc8919",
                "0x0000000000000000000000000000000000000000000a44490000000000000018":"0x0000000000000000000000003649A589684ba446f119e7040c39cCd7Ab9865c5",
                "0x0000000000000000000000000000000000000000000a44490000000000000019":"0x000000000000000000000000d94F53F17a1D2B921B58f9BB1a3cAae1075ff3B0",
                "0x0000000000000000000000000000000000000000000a4449000000000000001a":"0x000000000000000000000000625E61a2Ec4aB70dF50ccD6CC46B85db6Aa001f0",
                "0x0000000000000000000000000000000000000000000a4449000000000000001b":"0x0000000000000000000000002dD55E3F620c8DE08CBEBB9BF6Cd88d7b29aEd25",
                "0x0000000000000000000000db588e42D894EDE75000860fC0F4D969393ac51444":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000Ceda8D254c1925a79fd4cDe842A86e5273400eA444":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000611318F1430AD50c540677089053dD7123e3B9b144":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000010BeC6Cd8fEf4393b072d1Ddf853E753Cd3D222c44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000064C1D85AF78c82Bf7a59030E87114025F76750CE44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000009fEEa325F03230969B02cAEC585256D1B26ca3044":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000097161fE11Dba5a5daA85D1647eA79d72a3cf738544":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000066A2F34C7ce30A908F7ed9D5CAE33E939404cc4744":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000008f3922BebdDECB25D5513E393abBaA00C5dD5Ee144":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000004A48442762386D1954895a3d1977457145e8Ca744":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000009bf41d7a8aB4c11a4D0C5c5f451c06bf191E159344":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000CaA9c4aFcA7584462C1663DDE671adAc5532605f44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000f03F2c53784149f95535DBc707277363C6DB45A444":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000b921CBa7b3b0eEB811E462bd096F7432cdCD774544":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x000000000000000000000082F984c3D66c4df8F2D68CedA0cd5F2D1897dD2b44":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000B442685E3730F1441aE712E4F9E27946b847190244":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000218416320cc42b8FCD2a3700bc54d5eD668d6d0144":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000006fDcfade41750838F48f3Ec1fFDFdB4A3A42f74644":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000009D22D3DD44b0d7F073C78BCe7c1352175FB7B99744":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000877192DB751fD2b63C4CF0078221B2d0DAaa01eF44":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
                "0x0000000000000000000000fFe7c96064D3b4C185B1E4B22D3E692A01Ab2FbE44":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
                "0x0000000000000000000000e2117fF2836eD3f33B95bb8dBb4ACE9B4DB90e2E44":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
                "0x0000000000000000000000B1D1CAD653D38B90b586F7755963711f5e37E46944":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
                "0x00000000000000000000003660302D6614EF96578EFC3301cc527054Bc891944":"0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
                "0x00000000000000000000003649A589684ba446f119e7040c39cCd7Ab9865c544":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000d94F53F17a1D2B921B58f9BB1a3cAae1075ff3B044":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x0000000000000000000000625E61a2Ec4aB70dF50ccD6CC46B85db6Aa001f044":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000002dD55E3F620c8DE08CBEBB9BF6Cd88d7b29aEd2544":"0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
                "0x00000000000000000000db588e42D894EDE75000860fC0F4D969393ac5144e58":"0x000000000000000000000000db588e42D894EDE75000860fC0F4D969393ac514",
                "0x00000000000000000000Ceda8D254c1925a79fd4cDe842A86e5273400eA44e58":"0x000000000000000000000000Ceda8D254c1925a79fd4cDe842A86e5273400eA4",
                "0x00000000000000000000611318F1430AD50c540677089053dD7123e3B9b14e58":"0x000000000000000000000000611318F1430AD50c540677089053dD7123e3B9b1",
                "0x0000000000000000000010BeC6Cd8fEf4393b072d1Ddf853E753Cd3D222c4e58":"0x00000000000000000000000010BeC6Cd8fEf4393b072d1Ddf853E753Cd3D222c",
                "0x0000000000000000000064C1D85AF78c82Bf7a59030E87114025F76750CE4e58":"0x00000000000000000000000064C1D85AF78c82Bf7a59030E87114025F76750CE",
                "0x0000000000000000000009fEEa325F03230969B02cAEC585256D1B26ca304e58":"0x00000000000000000000000009fEEa325F03230969B02cAEC585256D1B26ca30",
                "0x0000000000000000000097161fE11Dba5a5daA85D1647eA79d72a3cf73854e58":"0x00000000000000000000000097161fE11Dba5a5daA85D1647eA79d72a3cf7385",
                "0x0000000000000000000066A2F34C7ce30A908F7ed9D5CAE33E939404cc474e58":"0x00000000000000000000000066A2F34C7ce30A908F7ed9D5CAE33E939404cc47",
                "0x000000000000000000008f3922BebdDECB25D5513E393abBaA00C5dD5Ee14e58":"0x0000000000000000000000008f3922BebdDECB25D5513E393abBaA00C5dD5Ee1",
                "0x0000000000000000000004A48442762386D1954895a3d1977457145e8Ca74e58":"0x00000000000000000000000004A48442762386D1954895a3d1977457145e8Ca7",
                "0x000000000000000000009bf41d7a8aB4c11a4D0C5c5f451c06bf191E15934e58":"0x0000000000000000000000009bf41d7a8aB4c11a4D0C5c5f451c06bf191E1593",
                "0x00000000000000000000CaA9c4aFcA7584462C1663DDE671adAc5532605f4e58":"0x000000000000000000000000CaA9c4aFcA7584462C1663DDE671adAc5532605f",
                "0x00000000000000000000f03F2c53784149f95535DBc707277363C6DB45A44e58":"0x000000000000000000000000f03F2c53784149f95535DBc707277363C6DB45A4",
                "0x00000000000000000000b921CBa7b3b0eEB811E462bd096F7432cdCD77454e58":"0x000000000000000000000000b921CBa7b3b0eEB811E462bd096F7432cdCD7745",
                "0x0000000000000000000082F984c3D66c4df8F2D68CedA0cd5F2D1897dD2b4e58":"0x00000000000000000000000082F984c3D66c4df8F2D68CedA0cd5F2D1897dD2b",
                "0x00000000000000000000B442685E3730F1441aE712E4F9E27946b84719024e58":"0x000000000000000000000000B442685E3730F1441aE712E4F9E27946b8471902",
                "0x00000000000000000000218416320cc42b8FCD2a3700bc54d5eD668d6d014e58":"0x000000000000000000000000218416320cc42b8FCD2a3700bc54d5eD668d6d01",
                "0x000000000000000000006fDcfade41750838F48f3Ec1fFDFdB4A3A42f7464e58":"0x0000000000000000000000006fDcfade41750838F48f3Ec1fFDFdB4A3A42f746",
                "0x000000000000000000009D22D3DD44b0d7F073C78BCe7c1352175FB7B9974e58":"0x0000000000000000000000009D22D3DD44b0d7F073C78BCe7c1352175FB7B997",
                "0x00000000000000000000877192DB751fD2b63C4CF0078221B2d0DAaa01eF4e58":"0x000000000000000000000000877192DB751fD2b63C4CF0078221B2d0DAaa01eF",
                "0x00000000000000000000fFe7c96064D3b4C185B1E4B22D3E692A01Ab2FbE4e58":"0x000000000000000000000000fFe7c96064D3b4C185B1E4B22D3E692A01Ab2FbE",
                "0x00000000000000000000e2117fF2836eD3f33B95bb8dBb4ACE9B4DB90e2E4e58":"0x000000000000000000000000e2117fF2836eD3f33B95bb8dBb4ACE9B4DB90e2E",
                "0x00000000000000000000B1D1CAD653D38B90b586F7755963711f5e37E4694e58":"0x000000000000000000000000B1D1CAD653D38B90b586F7755963711f5e37E469",
                "0x000000000000000000003660302D6614EF96578EFC3301cc527054Bc89194e58":"0x0000000000000000000000003660302D6614EF96578EFC3301cc527054Bc8919",
                "0x000000000000000000003649A589684ba446f119e7040c39cCd7Ab9865c54e58":"0x0000000000000000000000003649A589684ba446f119e7040c39cCd7Ab9865c5",
                "0x00000000000000000000d94F53F17a1D2B921B58f9BB1a3cAae1075ff3B04e58":"0x000000000000000000000000d94F53F17a1D2B921B58f9BB1a3cAae1075ff3B0",
                "0x00000000000000000000625E61a2Ec4aB70dF50ccD6CC46B85db6Aa001f04e58":"0x000000000000000000000000625E61a2Ec4aB70dF50ccD6CC46B85db6Aa001f0",
                "0x000000000000000000002dD55E3F620c8DE08CBEBB9BF6Cd88d7b29aEd254e58":"0x0000000000000000000000002dD55E3F620c8DE08CBEBB9BF6Cd88d7b29aEd25"
            },
            "balance":"2350000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUTz":{
            "balance":"10000000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUUs":{
            "balance":"25000000000000000000000000"
        },
        "MAN.2nRsUetjWAaYUizRkgBxGETimfUV2":{
            "balance":"10000000000000000000000000"
        },
        "MAN.4L95KmR3e8eUJvzwK2thft1eKaFYa":{
            "balance":"300000000000000000000000000"
        },
        "MAN.4739r322TyL3xCpbbdohS8NhBgGwi":{
            "balance":"200000000000000000000000000"
        },
        "MAN.2zXWsDtyt7vhVADGTz2yXD6h7WJnF":{
            "balance":"87650000000000000000000000"
        }
    },
    "mstate":{
        "Broadcasts":["MAN.2y5fqzGDWVznvkd49qqWpXiqjcmJF"],
        "curElect":[
            {
                "Account":"MAN.44EuST4f2vLeEMw2bsMWmBYqLMBhi",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3t9Ser2UjrXRT6erVKytjHtT4ohdX",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2MSWsig8iv45CDTrPn1XMzvZnR52v",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.EXnVsEqjyPLHZySL9y53WgXjiFmN",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2QR75feL9KBfaezJwhuM2VCPpqkyT",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.95Tro9wLULb6rNWNCT6QDwVBiDds",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.375qsgtLc25bJv2Cf9ffnEZPDhyvd",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2RwBWjBykMiGRu7STzYdNzXUMyh8z",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2zjKitA4uydg5kSZLjtSNHzDtx6k8",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.4kaowsz37i1WRrPrkg4g8qYQHtJ7",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3B1wx2wo5anTTMjnyWXAA4FQL5opx",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3pm3iXgrY9SYhhc9aS6DQCp4qj66t",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.4M8Svax1v6yGitwB4E2ueBtw2TesK",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3abRtpG81TPY8YwenowwgmJ8SuLpy",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2pqHNWtbpKKaLU71RY6TZeaQjjQnR",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3Wf4yzVkbou1bFzkpbJUdu5rqq1se",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.U5mNRj4q7jQzbVE4tWcjsvSXTseS",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2ZPbpaCyuEEf1WR3C6dCCFFqnH25J",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.3ByNBjw4E7gcxD3uGKtAZYLYUD5xi",
                "Stock":1,
                "Type":2
            },
            {
                "Account":"MAN.2tSj5kjwwaiTPt4XnHAaEWBYBo9gK",
                "Stock":1,
                "Type":0
            },
            {
                "Account":"MAN.4ZnJrUuM2bfFmdUaivrqGio28hZwd",
                "Stock":1,
                "Type":0
            },
            {
                "Account":"MAN.49fxdGyiWPQ3evpMCNChqvCq3qzMC",
                "Stock":1,
                "Type":0
            },
            {
                "Account":"MAN.3Ugik7ZsLoaNgX51kCJEWz1ZjxQgW",
                "Stock":1,
                "Type":0
            },
            {
                "Account":"MAN.kwPCJkajT2op7rVgYKDqcQu2KEQn",
                "Stock":1,
                "Type":0
            },
            {
                "Account":"MAN.ksFr4mKPfZhm2PrFdEUSoLoDsKAZ",
                "Stock":1,
                "Type":3
            },
            {
                "Account":"MAN.42bUyszBXL3feeDHztWMiUJCRzBRP",
                "Stock":1,
                "Type":3
            },
            {
                "Account":"MAN.2NVAVDc7AJGNP3Ghwfv8dz59kUvjM",
                "Stock":1,
                "Type":3
            },
            {
                "Account":"MAN.e33HPpmmZC98ADkZUXigS1nDFfaA",
                "Stock":1,
                "Type":3
            }
        ],
		"Foundation": "MAN.2zXWsDtyt7vhVADGTz2yXD6h7WJnF",
		"VersionSuperAccounts": [
			"MAN.4739r322TyL3xCpbbdohS8NhBgGwi"
		],
		"BlockSuperAccounts": [
			"MAN.4L95KmR3e8eUJvzwK2thft1eKaFYa"
		],
		"InnerMiners": [
		"MAN.3SPbc3M7bK8zCT8VbvjMGW2eCaBgY"
		],
		"BroadcastInterval": {
			"LastBCNumber": 0,
			"LastReelectNumber": 0,
			"BCInterval": 100,
			"BackupEnableNumber": 0,
			"BackupBCInterval": 0
		},
		"VIPCfg": [
					{
				"MinMoney": 0,
				"InterestRate": 5,
				"ElectUserNum": 0,
				"StockScale": 1000
			},
			{
				"MinMoney": 1000000,
				"InterestRate": 10,
				"ElectUserNum": 3,
				"StockScale": 1600
			},
			{
				"MinMoney": 10000000,
				"InterestRate": 15,
				"ElectUserNum": 5,
				"StockScale": 2000
			}
		],
        "BlkCalcCfg":"1",
        "TxsCalcCfg":"1",
        "InterestCalcCfg":"1",
        "LotteryCalcCfg":"1",
        "SlashCalcCfg":"1",
		"BlkRewardCfg": {
			"MinerMount": 3,
            "MinerAttenuationRate":8500,
            "MinerAttenuationNum":5000000,
			"ValidatorMount": 5,
            "ValidatorAttenuationRate":8500,
			"ValidatorAttenuationNum": 5000000,
			"RewardRate": {
				"MinerOutRate": 4000,
				"ElectedMinerRate": 5000,
				"FoundationMinerRate": 1000,
				"LeaderRate": 4000,
				"ElectedValidatorsRate": 5000,
				"FoundationValidatorRate": 1000,
				"OriginElectOfflineRate": 5000,
				"BackupRewardRate": 5000
			}
		},
		"TxsRewardCfg": {
			"MinersRate": 0,
			"ValidatorsRate": 10000,
			"RewardRate": {
				"MinerOutRate": 4000,
				"ElectedMinerRate": 6000,
				"FoundationMinerRate":0,
				"LeaderRate": 4000,
				"ElectedValidatorsRate": 6000,
				"FoundationValidatorRate": 0,
				"OriginElectOfflineRate": 5000,
				"BackupRewardRate": 5000
			}
		},
		"LotteryCfg": {
			"LotteryCalc": "1",
			"LotteryInfo": [{
				"PrizeLevel": 0,
				"PrizeNum": 1,
				"PrizeMoney": 6
			}]
		},
		"InterestCfg": {
            "RewardMount":2,
            "AttenuationRate":8500,
            "AttenuationPeriod":5000000,
			"PayInterval": 3600
		},
		"LeaderCfg": {
			"ParentMiningTime": 20,
			"PosOutTime": 20,
			"ReelectOutTime": 40,
			"ReelectHandleInterval": 3
		},
		"SlashCfg": {
			"SlashRate": 7500
		},
		"EleTime": {
			"MinerGen": 6,
			"MinerNetChange": 5,
			"ValidatorGen": 4,
			"ValidatorNetChange": 3,
			"VoteBeforeTime": 7
		},
		"EleInfo": {
			"ValidatorNum": 19,
			"BackValidator": 5,
			"ElectPlug": "layerd"
		},
		"ElectMinerNum": {
			"MinerNum": 21
		},
		"ElectBlackList": null,
		"ElectWhiteListSwitcherCfg":  {
			"Switcher" : false
		},
		"ElectWhiteList": null,
		"BlkProduceSlashCfg": {
			"Switcher" : true,
			"LowTHR"   : 1,
			"ProhibitCycleNum" : 2
		},
		"BlkProduceStatus": {
			"Number" : 0
		},
		"BlkProduceStats": {
			"StatsList": []
		},
       "BlkProduceBlackList" : {
			"BlackList" : []
		}
    },
  "config": {
					"chainID": 1,
					"byzantiumBlock": 0,
					"homesteadBlock": 0,
					"eip155Block": 0,
			"eip158Block": 0                        				             
	},
  "versionSignatures": [ "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"],
      "difficulty":"0x100",
    "timestamp":"0x5c26f140",
		"version": "1.0.0.0",
  
	"signatures": [	],
      "coinbase": "MAN.1111111111111111111cs",
      "leader":"MAN.CrsnQSJJfGxpb2taGhChLuyZwZJo", 
       "gasLimit": "0x2FEFD8",   
       "nonce": "0x0000000000000050",
       "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
       "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
	     "extraData": "0x0000000000000000"
}
`
	DefaultGenesisJson = `{
    "nettopology":{
    },
    "alloc":{},
    "mstate":{
		"BroadcastInterval" :{
			"LastBCNumber" : 0,
			"LastReelectNumber" : 0,
			"BCInterval" : 100,
			"BackupEnableNumber" : 0,
			"BackupBCInterval" : 0
		},
		"VIPCfg": [
					{
				"MinMoney": 0,
				"InterestRate": 5,
				"ElectUserNum": 0,
				"StockScale": 1000
			},
			{
				"MinMoney": 1000000,
				"InterestRate": 10,
				"ElectUserNum": 3,
				"StockScale": 1000      
			},
		{
				"MinMoney": 10000000,
				"InterestRate": 15,
				"ElectUserNum": 5,
				"StockScale": 1000      
			}
		],
        "BlkCalcCfg":"1",
        "TxsCalcCfg":"1",
        "InterestCalcCfg":"1",
        "LotteryCalcCfg":"1",
        "SlashCalcCfg":"1",
		"BlkRewardCfg": {
			"MinerMount": 3,
            "MinerAttenuationRate":8500,
			"MinerAttenuationNum": 5000000,
			"ValidatorMount": 5,
            "ValidatorAttenuationRate":8500,
			"ValidatorAttenuationNum": 5000000,
			"RewardRate": {
				"MinerOutRate": 4000,
				"ElectedMinerRate": 5000,
				"FoundationMinerRate": 1000,
				"LeaderRate": 4000,
				"ElectedValidatorsRate": 5000,
				"FoundationValidatorRate": 1000,
				"OriginElectOfflineRate": 5000,
				"BackupRewardRate": 5000
			}
		},
		"TxsRewardCfg": {
			"MinersRate": 0,
			"ValidatorsRate": 10000,
			"RewardRate": {
				"MinerOutRate": 4000,
				"ElectedMinerRate": 6000,
				"FoundationMinerRate":0,
				"LeaderRate": 4000,
				"ElectedValidatorsRate": 6000,
				"FoundationValidatorRate": 0,
				"OriginElectOfflineRate": 5000,
				"BackupRewardRate": 5000
			}
		},
		"LotteryCfg": {
			"LotteryCalc": "1",
			"LotteryInfo": [{
				"PrizeLevel": 0,
				"PrizeNum": 1,
				"PrizeMoney": 6
			}]
		},
		"InterestCfg": {
            "RewardMount":2,
            "AttenuationRate":8500,
            "AttenuationPeriod":5000000,
			"PayInterval": 3600
		},
		"LeaderCfg": {
			"ParentMiningTime": 20,
			"PosOutTime": 20,
			"ReelectOutTime": 40,
			"ReelectHandleInterval": 3
		},
		"SlashCfg": {
			"SlashRate": 7500
		},
		"EleTime": {
			"MinerGen": 6,
			"MinerNetChange": 5,
			"ValidatorGen": 4,
			"ValidatorNetChange": 3,
			"VoteBeforeTime": 7
		},
		"EleInfo": {
			"ValidatorNum": 19,
			"BackValidator": 5,
			"ElectPlug": "layerd"
		},
		"ElectMinerNum": {
			"MinerNum": 21
		},
		"ElectBlackList": null,
		"ElectWhiteListSwitcherCfg":  {
			"Switcher" : false
		},
		"ElectWhiteList": null,
		"BlkProduceSlashCfg": {
			"Switcher" : true,
			"LowTHR"   : 1,
			"ProhibitCycleNum" : 2
		},
		"BlkProduceStatus": {
			"Number" : 0
		},
		"BlkProduceStats": {
			"StatsList": []
		},
       "BlkProduceBlackList" : {
			"BlackList" : []
		}
    },
	"Currencys":{},
  "config": {
					"chainID": 1,           
					"byzantiumBlock": 0,
					"homesteadBlock": 0,
					"eip155Block": 0,
			"eip158Block": 0                        				             
	},
  "versionSignatures": [],
      "difficulty":"0x100",
    "timestamp":"0x5c26f140",
	"version": "1.0.0.0",
	"signatures": [	],
      "coinbase": "MAN.1111111111111111111cs",
    "leader":"MAN.1111111111111111111cs",      
       "gasLimit": "0x2FEFD8",   
    "nonce": "0x00000000000000178",            
       "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
       "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
	"extraData": "0x68656c6c6f2c77656c636f6d6520746f206d617472697820776f726c6421"  
}
`
)

func DefaultGenesis(genesisFile string) (*Genesis, error) {
	defGenesis := make(map[string]interface{})
	err := json.Unmarshal([]byte(DefaultGenesisJson), &defGenesis)
	if err != nil {
		return nil, err
	}
	if len(genesisFile) > 0 {
		file, err := os.Open(genesisFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		fileGenesis := make(map[string]interface{})
		if err := json.NewDecoder(file).Decode(&fileGenesis); err != nil {
			return nil, err
		}
		defGenesis = mergeGenesis(defGenesis, fileGenesis)
	}
	val, err := json.Marshal(defGenesis)
	if err != nil {
		return nil, err
	}
	genesis := new(Genesis)
	err = json.Unmarshal(val, genesis)
	if err != nil {
		return nil, err
	}
	return genesis, nil
}
func mergeGenesis(src, merge map[string]interface{}) map[string]interface{} {
	for key, value := range merge {
		if value == nil {
			src[key] = value
			continue
		}
		srcValue, exist := src[key]
		if exist {
			if reflect.TypeOf(value).Kind() == reflect.Map {
				src[key] = mergeGenesis(srcValue.(map[string]interface{}), value.(map[string]interface{}))
			} else {
				src[key] = value
			}
		} else {
			src[key] = value
		}
	}
	return src
}
