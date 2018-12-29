package core

import (
	"encoding/json"
	"github.com/matrix/go-matrix/base58"
	"github.com/matrix/go-matrix/common"
)

var (
	DefaultJson = `{"nettopology": {
		"Type": 0,
		"NetTopologyData": [{
				"Account": "MAN.CrsnQSJJfGxpb2taGhChLuyZwZJo",
				"Position": 8192
			},
			{
				"Account": "MAN.2Uoz8g8jauMa2mtnwxrschj2qPJrE",
				"Position": 8193
			},
			{
				"Account": "MAN.4Uuy7yduAjku29WHeveHSNpnZTRGt",
				"Position": 8194
			},
			{
				"Account": "MAN.3FCfHj3uhyhKZvcSW6cxjKd4BR9YH",
				"Position": 8195
			},
			{
				"Account": "MAN.2CUi6tLr3pAKsUHsF84qLiG42jLHx",
				"Position": 8196
			},
			{
				"Account": "MAN.32TKCX1ScAFvy3HxfoUWmZptervkU",
				"Position": 8197
			},
			{
				"Account": "MAN.33genSvo3BXwUG1gxVi3dtH27Pasb",
				"Position": 8198
			},
			{
				"Account": "MAN.2mNVd1SLzC8ohGnp29e5CmRHEc3rQ",
				"Position": 8199
			},
			{
				"Account": "MAN.3sQ1A1tUuVLNsu6RoLrXjhNi8UwgK",
				"Position": 8200
			},
			{
				"Account": "MAN.3Bkn4SBhPADY2TSqkhZxQA9c1XJit",
				"Position": 8201
			},
			{
				"Account": "MAN.2g5Tv4M74nzxr2FiTiomonfZfEhgW",
				"Position": 8202
			},
			{
				"Account": "MAN.9HE223J2nC8HYjEBecdB1xGXFETG",
				"Position": 8203
			},
			{
				"Account": "MAN.2xKT9DfzrqR7yUvADfC7VLCbPSiBb",
				"Position": 8204
			},
			{
				"Account": "MAN.5m2XU6yGoSJiAmFFkKKM5cdURUtF",
				"Position": 8205
			},
			{
				"Account": "MAN.3UEQqTmEgH7BEBQ6Gmt2E7ZpJTSQN",
				"Position": 8206
			},
			{
				"Account": "MAN.3ThvRGTUA1sqN2AFuptttq4RWRa8x",
				"Position": 8207
			},
			{
				"Account": "MAN.3S7eB5LSCSv5FPcyzfcmjuyPryjho",
				"Position": 8208
			},
			{
				"Account": "MAN.hqoxtEV5UWGAabYWjiPnN1EH7N9K",
				"Position": 8209
			},
			{
				"Account": "MAN.DbarsHeYBhqPFhqHFUqPYv5KDX56",
				"Position": 8210
			},
			{
				"Account": "MAN.3i11mJFLnmCBJxpJfWEDQig1EeqKp",
				"Position": 0
			},
			{
				"Account": "MAN.3QjcY2wzgnE6HLAVv6YBidqxXaNMX",
				"Position": 1
			},
			{
				"Account": "MAN.4D2RKmdYm3pGgs1P8AWFyt1VydXuc",
				"Position": 2
			},
			{
				"Account": "MAN.47Vc4qqMT8nP2PJA46ggu2fNu957Y",
				"Position": 3
			},
			{
				"Account": "MAN.2rVjisdngpyFJG6W2z5SBqpBRDWtx",
				"Position": 4
			}
		]
	},
	"alloc": {
		"MAN.1111111111111111111B8": {
			"storage": {
				"0x0000000000000000000000000000000000000000000000000000000a444e554d": "0x000000000000000000000000000000000000000000000000000000000000001c",
				"0x0000000000000000000000000000000000000000000a44490000000000000000": "0x0000000000000000000000000ead6cdb8d214389909a535d4ccc21a393dddba9",
				"0x0000000000000000000000000000000000000000000a44490000000000000001": "0x0000000000000000000000006a3217d128a76e4777403e092bde8362d4117773",
				"0x0000000000000000000000000000000000000000000a44490000000000000002": "0x00000000000000000000000055fbba0496ef137be57d4c179a3a74c4d0c643be",
				"0x0000000000000000000000000000000000000000000a44490000000000000003": "0x000000000000000000000000915b5295dde0cebb11c6cb25828b546a9b2c9369",
				"0x0000000000000000000000000000000000000000000a44490000000000000004": "0x00000000000000000000000092e0fea9aba517398c2f0dd628f8cfc7e32ba984",
				"0x0000000000000000000000000000000000000000000a44490000000000000005": "0x0000000000000000000000007eb0bcd103810a6bf463d6d230ebcacc85157d19",
				"0x0000000000000000000000000000000000000000000a44490000000000000006": "0x000000000000000000000000cded44bd41476a69e8e68ba8286952c414d28af7",
				"0x0000000000000000000000000000000000000000000a44490000000000000007": "0x0000000000000000000000009cde10b889fca53c0a560b90b3cb21c2fc965d2b",
				"0x0000000000000000000000000000000000000000000a44490000000000000008": "0x0000000000000000000000007823a1bea7aca2f902b87effdd4da9a7ef1fc5fb",
				"0x0000000000000000000000000000000000000000000a44490000000000000009": "0x0000000000000000000000000f96b686b2c57a0f2d571a39eae66d61a74b5870",
				"0x0000000000000000000000000000000000000000000a4449000000000000000a": "0x000000000000000000000000328b4bb56a750ea86bd52329a3e368d051699bb2",
				"0x0000000000000000000000000000000000000000000a4449000000000000000b": "0x000000000000000000000000b09b89893fd55223ed2d9c06cda7afef867c7449",
				"0x0000000000000000000000000000000000000000000a4449000000000000000c": "0x000000000000000000000000aea37855eacb4b41ca0dbc6c744681f96fe09d94",
				"0x0000000000000000000000000000000000000000000a4449000000000000000d": "0x000000000000000000000000b142159adbfc2690b45da01e49cfa2379ddc2701",
				"0x0000000000000000000000000000000000000000000a4449000000000000000e": "0x000000000000000000000000f9e18acc86179925353713a4a5d0e9bf381fbc17",
				"0x0000000000000000000000000000000000000000000a4449000000000000000f": "0x000000000000000000000000a121e6670439ba37e7244d4eb18e42bd6724ef0f",
				"0x0000000000000000000000000000000000000000000a44490000000000000010": "0x0000000000000000000000000a3f28de9682df49f9f393931062c5204c2bc404",
				"0x0000000000000000000000000000000000000000000a44490000000000000011": "0x0000000000000000000000008c3d1a9504a36d49003f1652fadb9f06c32a4408",
				"0x0000000000000000000000000000000000000000000a44490000000000000012": "0x00000000000000000000000005e3c16931c6e578f948231dca609d754c18fc09",
				"0x0000000000000000000000000000000000000000000a44490000000000000013": "0x000000000000000000000000c24c56af638a788b76d7e9c058f29680b323344e",
				"0x0000000000000000000000000000000000000000000a44490000000000000014": "0x000000000000000000000000acee53a890b8418529e71112c0d5c1eb760b9f1a",
				"0x0000000000000000000000000000000000000000000a44490000000000000015": "0x000000000000000000000000e637b65ecf0646c6fdba29cf560039985a424d50",
				"0x0000000000000000000000000000000000000000000a44490000000000000016": "0x000000000000000000000000df5f4e53ea3e21cb1f5fa771a18584224078237e",
				"0x0000000000000000000000000000000000000000000a44490000000000000017": "0x000000000000000000000000850863159d8a203602163bdc593513a71d090535",
				"0x0000000000000000000000000000000000000000000a44490000000000000018": "0x00000000000000000000000048ae42f3a537598e2a0587abda9617ec7bde2610",
				"0x0000000000000000000000000000000000000000000a44490000000000000019": "0x000000000000000000000000452339d9e06de12574e4ef9727330d22c3a168f6",
				"0x0000000000000000000000000000000000000000000a4449000000000000001a": "0x000000000000000000000000fd84bdd98ad4b628c7174d5671ffaeccab883a22",
				"0x0000000000000000000000000000000000000000000a4449000000000000001b": "0x00000000000000000000000089ee5942605231a9bb95a4e6a688363465d157b4",
				"0x000000000000000000000005e3c16931c6e578f948231dca609d754c18fc0944": "0x000000000000000000000000000000000000000000084595161401484A000000",
				"0x00000000000000000000000a3f28de9682df49f9f393931062c5204c2bc40444": "0x000000000000000000000000000000000000000000084595161401484A000000",
				"0x00000000000000000000000ead6cdb8d214389909a535d4ccc21a393dddba944": "0x000000000000000000000000000000000000000000084595161401484A000000",
				"0x00000000000000000000000f96b686b2c57a0f2d571a39eae66d61a74b587044": "0x000000000000000000000000000000000000000000084595161401484A000000",
				"0x0000000000000000000000328b4bb56a750ea86bd52329a3e368d051699bb244": "0x000000000000000000000000000000000000000000084595161401484A000000",
				"0x000000000000000000000055fbba0496ef137be57d4c179a3a74c4d0c643be44": "0x00000000000000000000000000000000000000000000D3C21BCECCEDA1000000",
				"0x00000000000000000000006a3217d128a76e4777403e092bde8362d411777344": "0x00000000000000000000000000000000000000000000D3C21BCECCEDA1000000",
				"0x00000000000000000000007823a1bea7aca2f902b87effdd4da9a7ef1fc5fb44": "0x00000000000000000000000000000000000000000000D3C21BCECCEDA1000000",
				"0x00000000000000000000007eb0bcd103810a6bf463d6d230ebcacc85157d1944": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x00000000000000000000008c3d1a9504a36d49003f1652fadb9f06c32a440844": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x0000000000000000000000915b5295dde0cebb11c6cb25828b546a9b2c936944": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x000000000000000000000092e0fea9aba517398c2f0dd628f8cfc7e32ba98444": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x00000000000000000000009cde10b889fca53c0a560b90b3cb21c2fc965d2b44": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x0000000000000000000000a121e6670439ba37e7244d4eb18e42bd6724ef0f44": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x0000000000000000000000aea37855eacb4b41ca0dbc6c744681f96fe09d9444": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x0000000000000000000000b09b89893fd55223ed2d9c06cda7afef867c744944": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x0000000000000000000000b142159adbfc2690b45da01e49cfa2379ddc270144": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x0000000000000000000000cded44bd41476a69e8e68ba8286952c414d28af744": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x0000000000000000000000f9e18acc86179925353713a4a5d0e9bf381fbc1744": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x0000000000000000000000c24c56af638a788b76d7e9c058f29680b323344e44": "0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
				"0x0000000000000000000000acee53a890b8418529e71112c0d5c1eb760b9f1a44": "0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
				"0x0000000000000000000000e637b65ecf0646c6fdba29cf560039985a424d5044": "0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
				"0x0000000000000000000000df5f4e53ea3e21cb1f5fa771a18584224078237e44": "0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
				"0x0000000000000000000000850863159d8a203602163bdc593513a71d09053544": "0x00000000000000000000000000000000000000000000021e19e0c9bab2400000",
				"0x000000000000000000000048ae42f3a537598e2a0587abda9617ec7bde261044": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x0000000000000000000000452339d9e06de12574e4ef9727330d22c3a168f644": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x0000000000000000000000fd84bdd98ad4b628c7174d5671ffaeccab883a2244": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x000000000000000000000089ee5942605231a9bb95a4e6a688363465d157b444": "0x00000000000000000000000000000000000000000000152d02c7e14af6800000",
				"0x0000000000000000000005e3c16931c6e578f948231dca609d754c18fc094e58": "0x00000000000000000000000005e3c16931c6e578f948231dca609d754c18fc09",
				"0x000000000000000000000a3f28de9682df49f9f393931062c5204c2bc4044e58": "0x0000000000000000000000000a3f28de9682df49f9f393931062c5204c2bc404",
				"0x000000000000000000000ead6cdb8d214389909a535d4ccc21a393dddba94e58": "0x0000000000000000000000000ead6cdb8d214389909a535d4ccc21a393dddba9",
				"0x000000000000000000000f96b686b2c57a0f2d571a39eae66d61a74b58704e58": "0x0000000000000000000000000f96b686b2c57a0f2d571a39eae66d61a74b5870",
				"0x00000000000000000000328b4bb56a750ea86bd52329a3e368d051699bb24e58": "0x000000000000000000000000328b4bb56a750ea86bd52329a3e368d051699bb2",
				"0x0000000000000000000055fbba0496ef137be57d4c179a3a74c4d0c643be4e58": "0x00000000000000000000000055fbba0496ef137be57d4c179a3a74c4d0c643be",
				"0x000000000000000000006a3217d128a76e4777403e092bde8362d41177734e58": "0x0000000000000000000000006a3217d128a76e4777403e092bde8362d4117773",
				"0x000000000000000000007823a1bea7aca2f902b87effdd4da9a7ef1fc5fb4e58": "0x0000000000000000000000007823a1bea7aca2f902b87effdd4da9a7ef1fc5fb",
				"0x000000000000000000007eb0bcd103810a6bf463d6d230ebcacc85157d194e58": "0x0000000000000000000000007eb0bcd103810a6bf463d6d230ebcacc85157d19",
				"0x000000000000000000008c3d1a9504a36d49003f1652fadb9f06c32a44084e58": "0x0000000000000000000000008c3d1a9504a36d49003f1652fadb9f06c32a4408",
				"0x00000000000000000000915b5295dde0cebb11c6cb25828b546a9b2c93694e58": "0x000000000000000000000000915b5295dde0cebb11c6cb25828b546a9b2c9369",
				"0x0000000000000000000092e0fea9aba517398c2f0dd628f8cfc7e32ba9844e58": "0x00000000000000000000000092e0fea9aba517398c2f0dd628f8cfc7e32ba984",
				"0x000000000000000000009cde10b889fca53c0a560b90b3cb21c2fc965d2b4e58": "0x0000000000000000000000009cde10b889fca53c0a560b90b3cb21c2fc965d2b",
				"0x00000000000000000000a121e6670439ba37e7244d4eb18e42bd6724ef0f4e58": "0x000000000000000000000000a121e6670439ba37e7244d4eb18e42bd6724ef0f",
				"0x00000000000000000000aea37855eacb4b41ca0dbc6c744681f96fe09d944e58": "0x000000000000000000000000aea37855eacb4b41ca0dbc6c744681f96fe09d94",
				"0x00000000000000000000b09b89893fd55223ed2d9c06cda7afef867c74494e58": "0x000000000000000000000000b09b89893fd55223ed2d9c06cda7afef867c7449",
				"0x00000000000000000000b142159adbfc2690b45da01e49cfa2379ddc27014e58": "0x000000000000000000000000b142159adbfc2690b45da01e49cfa2379ddc2701",
				"0x00000000000000000000cded44bd41476a69e8e68ba8286952c414d28af74e58": "0x000000000000000000000000cded44bd41476a69e8e68ba8286952c414d28af7",
				"0x00000000000000000000f9e18acc86179925353713a4a5d0e9bf381fbc174e58": "0x000000000000000000000000f9e18acc86179925353713a4a5d0e9bf381fbc17",
				"0x00000000000000000000c24c56af638a788b76d7e9c058f29680b323344e4e58": "0x000000000000000000000000c24c56af638a788b76d7e9c058f29680b323344e",
				"0x00000000000000000000acee53a890b8418529e71112c0d5c1eb760b9f1a4e58": "0x000000000000000000000000acee53a890b8418529e71112c0d5c1eb760b9f1a",
				"0x00000000000000000000e637b65ecf0646c6fdba29cf560039985a424d504e58": "0x000000000000000000000000e637b65ecf0646c6fdba29cf560039985a424d50",
				"0x00000000000000000000df5f4e53ea3e21cb1f5fa771a18584224078237e4e58": "0x000000000000000000000000df5f4e53ea3e21cb1f5fa771a18584224078237e",
				"0x00000000000000000000850863159d8a203602163bdc593513a71d0905354e58": "0x000000000000000000000000850863159d8a203602163bdc593513a71d090535",
				"0x0000000000000000000048ae42f3a537598e2a0587abda9617ec7bde26104e58": "0x00000000000000000000000048ae42f3a537598e2a0587abda9617ec7bde2610",
				"0x00000000000000000000452339d9e06de12574e4ef9727330d22c3a168f64e58": "0x000000000000000000000000452339d9e06de12574e4ef9727330d22c3a168f6",
				"0x00000000000000000000fd84bdd98ad4b628c7174d5671ffaeccab883a224e58": "0x000000000000000000000000fd84bdd98ad4b628c7174d5671ffaeccab883a22",
				"0x0000000000000000000089ee5942605231a9bb95a4e6a688363465d157b44e58": "0x00000000000000000000000089ee5942605231a9bb95a4e6a688363465d157b4"
			},
			"balance": "30000000000000000000000000000000"
		},
		"MAN.2nRsUetjWAaYUizRkgBxGETimfUTz": {
			"balance": "10000000000000000000000000"
		},
		"MAN.2nRsUetjWAaYUizRkgBxGETimfUUs": {
			"balance": "25000000000000000000000000"
		},
		"MAN.2nRsUetjWAaYUizRkgBxGETimfUV2": {
			"balance": "10000000000000000000000000"
		},
		"MAN.2nRsUetjWAaYUizRkgBxGETimfUW7": {
			"balance": "5000000000000000000000000"
		},
		"MAN.2nRsUetjWAaYUizRkgBxGETimfUXN": {
			"balance": "10000000000000000000000000"
		},
		"MAN.CrsnQSJJfGxpb2taGhChLuyZwZJo": {
			"balance": "30000000000000000000000000"
		},
		"MAN.2Uoz8g8jauMa2mtnwxrschj2qPJrE": {
			"balance": "30000000000000000000000000"
		}
	},
	"mstate": {
			"curElect": [{
				"Account": "MAN.CrsnQSJJfGxpb2taGhChLuyZwZJo",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.2Uoz8g8jauMa2mtnwxrschj2qPJrE",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.4Uuy7yduAjku29WHeveHSNpnZTRGt",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.3FCfHj3uhyhKZvcSW6cxjKd4BR9YH",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.2CUi6tLr3pAKsUHsF84qLiG42jLHx",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.32TKCX1ScAFvy3HxfoUWmZptervkU",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.33genSvo3BXwUG1gxVi3dtH27Pasb",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.2mNVd1SLzC8ohGnp29e5CmRHEc3rQ",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.3sQ1A1tUuVLNsu6RoLrXjhNi8UwgK",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.3Bkn4SBhPADY2TSqkhZxQA9c1XJit",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.2g5Tv4M74nzxr2FiTiomonfZfEhgW",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.9HE223J2nC8HYjEBecdB1xGXFETG",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.2xKT9DfzrqR7yUvADfC7VLCbPSiBb",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.5m2XU6yGoSJiAmFFkKKM5cdURUtF",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.3UEQqTmEgH7BEBQ6Gmt2E7ZpJTSQN",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.3ThvRGTUA1sqN2AFuptttq4RWRa8x",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.3S7eB5LSCSv5FPcyzfcmjuyPryjho",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.hqoxtEV5UWGAabYWjiPnN1EH7N9K",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.DbarsHeYBhqPFhqHFUqPYv5KDX56",
				"Stock": 1,
				"Type": 2
			},
			{
				"Account": "MAN.3i11mJFLnmCBJxpJfWEDQig1EeqKp",
				"Stock": 1,
				"Type": 0
			},
			{
				"Account": "MAN.3QjcY2wzgnE6HLAVv6YBidqxXaNMX",
				"Stock": 1,
				"Type": 0
			},
			{
				"Account": "MAN.4D2RKmdYm3pGgs1P8AWFyt1VydXuc",
				"Stock": 1,
				"Type": 0
			},
			{
				"Account": "MAN.47Vc4qqMT8nP2PJA46ggu2fNu957Y",
				"Stock": 1,
				"Type": 0
			},
			{
				"Account": "MAN.2rVjisdngpyFJG6W2z5SBqpBRDWtx",
				"Stock": 1,
				"Type": 0
			}

		],
		"Broadcast": "MAN.38nGzwi5Xn5ApxHXquT8ALaMLpbyG",
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
			"BCInterval": 100
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
		"BlkRewardCfg": {
			"MinerMount": 3,
			"MinerHalf": 5000000,
			"ValidatorMount": 7,
			"ValidatorHalf": 5000000,
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
			"CalcInterval": 100,
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
			"MinerGen": 9,
			"MinerNetChange": 5,
			"ValidatorGen": 9,
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
		"ElectWhiteList": null
	},
	"config": {
		"chainID": 1,
		"byzantiumBlock": 0,
		"homesteadBlock": 0,
		"eip155Block": 0,
		"eip158Block": 0
	},
	"extraData": "0x00",
	"version": "1.0.0-stable",
  "versionSignatures": [
    [
      181,
      8,
      246,
      28,
      118,
      103,
      127,
      70,
      144,
      31,
      187,
      28,
      71,
      14,
      164,
      113,
      133,
      96,
      141,
      160,
      117,
      234,
      127,
      5,
      254,
      240,
      146,
      127,
      39,
      247,
      161,
      150,
      75,
      243,
      248,
      192,
      32,
      110,
      149,
      242,
      151,
      195,
      226,
      167,
      74,
      223,
      135,
      250,
      233,
      174,
      109,
      239,
      101,
      177,
      155,
      129,
      68,
      92,
      218,
      222,
      45,
      207,
      165,
      112,
      0
    ]
  ],
	"vrfvalue": "",
	"nextElect": [],
	"difficulty": "0x100",
	
	"signatures": [	],
	"coinbase": "MAN.1111111111111111111cs",
	"leader": "MAN.CrsnQSJJfGxpb2taGhChLuyZwZJo",
	"gasLimit": "0x2FEFD8",
	"nonce": "0x0000000000000050",
	"mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
	"parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
	"timestamp": "0x00"
}`
)

func DefaultGenesisToEthGensis(gensis1 *Genesis1, gensis *Genesis) *Genesis {
	if nil != gensis1.Config {
		gensis.Config = gensis1.Config
	}
	if gensis1.Nonce != 0 {
		gensis.Nonce = gensis1.Nonce
	}
	if gensis1.Timestamp != 0 {
		gensis.Timestamp = gensis1.Timestamp
	}
	if len(gensis1.ExtraData) != 0 {
		gensis.ExtraData = gensis1.ExtraData
	}
	if gensis1.Version != "" {
		gensis.Version = gensis1.Version
	}
	if len(gensis1.VersionSignatures) != 0 {
		gensis.VersionSignatures = gensis1.VersionSignatures
	}
	if len(gensis1.VrfValue) != 0 {
		gensis.VrfValue = gensis1.VrfValue
	}
	if len(gensis1.Signatures) != 0 {
		gensis.Signatures = gensis1.Signatures
	}
	if nil != gensis1.Difficulty {
		gensis.Difficulty = gensis1.Difficulty
	}
	if gensis1.Mixhash.Equal(common.Hash{}) == false {
		gensis.Mixhash = gensis1.Mixhash
	}
	if gensis1.Number != 0 {
		gensis.Number = gensis1.Number
	}
	if gensis1.GasUsed != 0 {
		gensis.GasUsed = gensis1.GasUsed
	}
	if gensis1.ParentHash.Equal(common.Hash{}) == false {
		gensis.ParentHash = gensis1.ParentHash
	}

	if gensis1.Leader != "" {
		gensis.Leader = base58.Base58DecodeToAddress(gensis1.Leader)
	}
	if gensis1.Coinbase != "" {
		gensis.Coinbase = base58.Base58DecodeToAddress(gensis1.Coinbase)
	}
	if gensis1.Root.Equal(common.Hash{}) == false {
		gensis.Root = gensis1.Root
	}
	if gensis1.TxHash.Equal(common.Hash{}) == false {
		gensis.TxHash = gensis1.TxHash
	}
	//nextElect
	if nil != gensis1.NextElect {
		sliceElect := make([]common.Elect, 0)
		for _, elec := range gensis1.NextElect {
			tmp := new(common.Elect)
			tmp.Account = base58.Base58DecodeToAddress(elec.Account)
			tmp.Stock = elec.Stock
			tmp.Type = elec.Type
			sliceElect = append(sliceElect, *tmp)
		}
		gensis.NextElect = sliceElect
	}

	//NetTopology
	if len(gensis1.NetTopology.NetTopologyData) != 0 {
		sliceNetTopologyData := make([]common.NetTopologyData, 0)
		for _, netTopology := range gensis1.NetTopology.NetTopologyData {
			tmp := new(common.NetTopologyData)
			tmp.Account = base58.Base58DecodeToAddress(netTopology.Account)
			tmp.Position = netTopology.Position
			sliceNetTopologyData = append(sliceNetTopologyData, *tmp)
		}
		gensis.NetTopology.NetTopologyData = sliceNetTopologyData
		gensis.NetTopology.Type = gensis1.NetTopology.Type
	}

	//Alloc
	if nil != gensis1.Alloc {
		gensis.Alloc = make(GenesisAlloc)
		for kString, vGenesisAccount := range gensis1.Alloc {
			tmpk := base58.Base58DecodeToAddress(kString)
			gensis.Alloc[tmpk] = vGenesisAccount
		}
	}

	if nil != gensis1.MState {
		if gensis.MState == nil {
			gensis.MState = new(GenesisMState)
		}
		if nil != gensis1.MState.Broadcast {
			gensis.MState.Broadcast = new(common.Address)
			*gensis.MState.Broadcast = base58.Base58DecodeToAddress(*gensis1.MState.Broadcast)
		}
		if nil != gensis1.MState.Foundation {
			gensis.MState.Foundation = new(common.Address)
			*gensis.MState.Foundation = base58.Base58DecodeToAddress(*gensis1.MState.Foundation)
		}
		if nil != gensis1.MState.VersionSuperAccounts {
			versionSuperAccounts := make([]common.Address, 0)
			for _, v := range *gensis1.MState.VersionSuperAccounts {
				versionSuperAccounts = append(versionSuperAccounts, base58.Base58DecodeToAddress(v))
			}
			gensis.MState.VersionSuperAccounts = &versionSuperAccounts
		}
		if nil != gensis1.MState.BlockSuperAccounts {
			blockSuperAccounts := make([]common.Address, 0)
			for _, v := range *gensis1.MState.BlockSuperAccounts {
				blockSuperAccounts = append(blockSuperAccounts, base58.Base58DecodeToAddress(v))
			}
			gensis.MState.BlockSuperAccounts = &blockSuperAccounts
		}
		if nil != gensis1.MState.InnerMiners {
			innerMiners := make([]common.Address, 0)
			for _, v := range *gensis1.MState.InnerMiners {
				innerMiners = append(innerMiners, base58.Base58DecodeToAddress(v))
			}
			gensis.MState.InnerMiners = &innerMiners
		}
		if nil != gensis1.MState.ElectBlackListCfg {
			blackList := make([]common.Address, 0)
			for _, v := range *gensis1.MState.ElectBlackListCfg {
				blackList = append(blackList, base58.Base58DecodeToAddress(v))
			}
			gensis.MState.ElectBlackListCfg = &blackList
		}
		if nil != gensis1.MState.ElectWhiteListCfg {
			whiteList := make([]common.Address, 0)
			for _, v := range *gensis1.MState.ElectWhiteListCfg {
				whiteList = append(whiteList, base58.Base58DecodeToAddress(v))
			}
			gensis.MState.ElectBlackListCfg = &whiteList
		}
		if nil != gensis1.MState.ElectMinerNumCfg {
			gensis.MState.ElectMinerNumCfg = gensis1.MState.ElectMinerNumCfg
		}
		if nil != gensis1.MState.BlkRewardCfg {
			gensis.MState.BlkRewardCfg = gensis1.MState.BlkRewardCfg
		}
		if nil != gensis1.MState.TxsRewardCfg {
			gensis.MState.TxsRewardCfg = gensis1.MState.TxsRewardCfg
		}
		if nil != gensis1.MState.InterestCfg {
			gensis.MState.InterestCfg = gensis1.MState.InterestCfg
		}
		if nil != gensis1.MState.LotteryCfg {
			gensis.MState.LotteryCfg = gensis1.MState.LotteryCfg
		}
		if nil != gensis1.MState.SlashCfg {
			gensis.MState.SlashCfg = gensis1.MState.SlashCfg
		}
		if nil != gensis1.MState.BCICfg {
			gensis.MState.BCICfg = gensis1.MState.BCICfg
		}
		if nil != gensis1.MState.VIPCfg {
			gensis.MState.VIPCfg = gensis1.MState.VIPCfg
		}
		if nil != gensis1.MState.LeaderCfg {
			gensis.MState.LeaderCfg = gensis1.MState.LeaderCfg
		}
		if nil != gensis1.MState.EleTimeCfg {
			gensis.MState.EleTimeCfg = gensis1.MState.EleTimeCfg
		}
		if nil != gensis1.MState.EleInfoCfg {
			gensis.MState.EleInfoCfg = gensis1.MState.EleInfoCfg
		}
		//curElect
		if nil != gensis1.MState.CurElect {
			sliceElect := make([]common.Elect, 0)
			for _, elec := range *gensis1.MState.CurElect {
				tmp := new(common.Elect)
				tmp.Account = base58.Base58DecodeToAddress(elec.Account)
				tmp.Stock = elec.Stock
				tmp.Type = elec.Type
				sliceElect = append(sliceElect, *tmp)
			}
			gensis.MState.CurElect = &sliceElect
		}
	}
	return gensis
}

func GetDefaultGeneis() (*Genesis, error) {
	genesis := new(Genesis)
	defaultGenesis1 := new(Genesis1)
	err := json.Unmarshal([]byte(DefaultJson), defaultGenesis1)
	if err != nil {
		return nil, err
	}
	genesis = DefaultGenesisToEthGensis(defaultGenesis1, genesis)
	return genesis, nil

}