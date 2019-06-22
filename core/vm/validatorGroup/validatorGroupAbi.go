// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php
package validatorGroup

import (
	"strings"
	"github.com/MatrixAINetwork/go-matrix/accounts/abi"
)

var (
	validatorGroupJson = `[
    {
      "constant": true,
      "inputs": [],
      "name": "owner",
      "outputs": [
        {
          "name": "",
          "type": "address"
        }
      ],
      "payable": false,
      "stateMutability": "view",
      "type": "function"
    },
    {
      "inputs": [],
      "payable": false,
      "stateMutability": "payable",
      "type": "constructor"
    },
    {
      "anonymous": false,
      "inputs": [],
      "name": "WithDrawAll",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "name": "from",
          "type": "address"
        },
        {
          "indexed": false,
          "name": "amount",
          "type": "uint256"
        },
        {
          "indexed": false,
          "name": "dType",
          "type": "uint256"
        }
      ],
      "name": "AddDeposit",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "name": "from",
          "type": "address"
        },
        {
          "indexed": false,
          "name": "amount",
          "type": "uint256"
        },
        {
          "indexed": false,
          "name": "dType",
          "type": "uint256"
        }
      ],
      "name": "Withdraw",
      "type": "event"
    },
    {
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "name": "from",
          "type": "address"
        }
      ],
      "name": "Refund",
      "type": "event"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "newSigner",
          "type": "address"
        }
      ],
      "name": "setSignAccount",
      "outputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [],
      "name": "withdrawAll",
      "outputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "dType",
          "type": "uint256"
        }
      ],
      "name": "addDeposit",
      "outputs": [],
      "payable": true,
      "stateMutability": "payable",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "amount",
          "type": "uint256"
        },
        {
          "name": "position",
          "type": "uint256"
        }
      ],
      "name": "withdraw",
      "outputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [],
      "name": "getReward",
      "outputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    },
    {
      "constant": false,
      "inputs": [{
          "name": "position",
          "type": "uint256"
        }],
      "name": "refund",
      "outputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    }
  ]`
  validatorGroupContractJson = `[
    {
      "inputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "constructor"
    },
	{
      "anonymous": false,
      "inputs": [
        {
          "indexed": true,
          "name": "from",
          "type": "address"
        },
        {
          "indexed": true,
          "name": "signAccount",
          "type": "address"
        },
        {
          "indexed": true,
          "name": "contractAddr",
          "type": "address"
        },
        {
          "indexed": false,
          "name": "amount",
          "type": "uint256"
        },
        {
          "indexed": false,
          "name": "dType",
          "type": "uint256"
        }
      ],
      "name": "CreateValidatorGroup",
      "type": "event"
    },
    {
      "constant": false,
      "inputs": [
        {
          "name": "signAcount",
          "type": "address"
        },
        {
          "name": "dType",
          "type": "uint256"
        },
        {
          "name": "ownerRate",
          "type": "uint256"
        },
        {
          "name": "nodeRate",
          "type": "uint256"
        },
        {
          "name": "lvlRate",
          "type": "uint256[]"
        }
      ],
      "name": "createValidatorGroup",
      "outputs": [],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    }
  ]`
    ValidatorGroupAbi, Abierr = abi.JSON(strings.NewReader(validatorGroupJson))
	ValidatorGroupContractAbi, Abi1err = abi.JSON(strings.NewReader(validatorGroupContractJson))
)
