// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/internal/manapi"
	"github.com/matrix/go-matrix/log"
	"github.com/matrix/go-matrix/signer/core"
	"github.com/matrix/go-matrix/signer/rules/deps"
	"github.com/matrix/go-matrix/signer/storage"
	"github.com/robertkrimen/otto"
)

var (
	BigNumber_JS = deps.MustAsset("bignumber.js")
)

// consoleOutput is an override for the console.log and console.error methods to
// stream the output into the configured output stream instead of stdout.
func consoleOutput(call otto.FunctionCall) otto.Value {
	output := []string{"JS:> "}
	for _, argument := range call.ArgumentList {
		output = append(output, fmt.Sprintf("%v", argument))
	}
	fmt.Fprintln(os.Stdout, strings.Join(output, " "))
	return otto.Value{}
}

// rulesetUI provides an implementation of SignerUI that evaluates a javascript
// file for each defined UI-method
type rulesetUI struct {
	next        core.SignerUI // The next handler, for manual processing
	storage     storage.Storage
	credentials storage.Storage
	jsRules     string // The rules to use
}

func NewRuleEvaluator(next core.SignerUI, jsbackend, credentialsBackend storage.Storage) (*rulesetUI, error) {
	c := &rulesetUI{
		next:        next,
		storage:     jsbackend,
		credentials: credentialsBackend,
		jsRules:     "",
	}

	return c, nil
}

func (r *rulesetUI) Init(javascriptRules string) error {
	r.jsRules = javascriptRules
	return nil
}
func (r *rulesetUI) execute(jsfunc string, jsarg interface{}) (otto.Value, error) {

	// Instantiate a fresh vm engine every time
	vm := otto.New()
	// Set the native callbacks
	consoleObj, _ := vm.Get("console")
	consoleObj.Object().Set("log", consoleOutput)
	consoleObj.Object().Set("error", consoleOutput)
	vm.Set("storage", r.storage)

	// Load bootstrap libraries
	script, err := vm.Compile("bignumber.js", BigNumber_JS)
	if err != nil {
		log.Warn("Failed loading libraries", "err", err)
		return otto.UndefinedValue(), err
	}
	vm.Run(script)

	// Run the actual rule implementation
	_, err = vm.Run(r.jsRules)
	if err != nil {
		log.Warn("Execution failed", "err", err)
		return otto.UndefinedValue(), err
	}

	// And the actual call
	// All calls are objects with the parameters being keys in that object.
	// To provide additional insulation between js and go, we serialize it into JSON on the Go-side,
	// and deserialize it on the JS side.

	jsonbytes, err := json.Marshal(jsarg)
	if err != nil {
		log.Warn("failed marshalling data", "data", jsarg)
		return otto.UndefinedValue(), err
	}
	// Now, we call foobar(JSON.parse(<jsondata>)).
	var call string
	if len(jsonbytes) > 0 {
		call = fmt.Sprintf("%v(JSON.parse(%v))", jsfunc, string(jsonbytes))
	} else {
		call = fmt.Sprintf("%v()", jsfunc)
	}
	return vm.Run(call)
}

func (r *rulesetUI) checkApproval(jsfunc string, jsarg []byte, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	v, err := r.execute(jsfunc, string(jsarg))
	if err != nil {
		log.Info("error occurred during execution", "error", err)
		return false, err
	}
	result, err := v.ToString()
	if err != nil {
		log.Info("error occurred during response unmarshalling", "error", err)
		return false, err
	}
	if result == "Approve" {
		log.Info("Op approved")
		return true, nil
	} else if result == "Reject" {
		log.Info("Op rejected")
		return false, nil
	}
	return false, fmt.Errorf("Unknown response")
}

func (r *rulesetUI) ApproveTx(request *core.SignTxRequest) (core.SignTxResponse, error) {
	jsonreq, err := json.Marshal(request)
	approved, err := r.checkApproval("ApproveTx", jsonreq, err)
	if err != nil {
		log.Info("Rule-based approval error, going to manual", "error", err)
		return r.next.ApproveTx(request)
	}

	if approved {
		return core.SignTxResponse{
				Transaction: request.Transaction,
				Approved:    true,
				Password:    r.lookupPassword(request.Transaction.From.Address()),
			},
			nil
	}
	return core.SignTxResponse{Approved: false}, err
}

func (r *rulesetUI) lookupPassword(address common.Address) string {
	return r.credentials.Get(strings.ToLower(address.String()))
}

func (r *rulesetUI) ApproveSignData(request *core.SignDataRequest) (core.SignDataResponse, error) {
	jsonreq, err := json.Marshal(request)
	approved, err := r.checkApproval("ApproveSignData", jsonreq, err)
	if err != nil {
		log.Info("Rule-based approval error, going to manual", "error", err)
		return r.next.ApproveSignData(request)
	}
	if approved {
		return core.SignDataResponse{Approved: true, Password: r.lookupPassword(request.Address.Address())}, nil
	}
	return core.SignDataResponse{Approved: false, Password: ""}, err
}

func (r *rulesetUI) ApproveExport(request *core.ExportRequest) (core.ExportResponse, error) {
	jsonreq, err := json.Marshal(request)
	approved, err := r.checkApproval("ApproveExport", jsonreq, err)
	if err != nil {
		log.Info("Rule-based approval error, going to manual", "error", err)
		return r.next.ApproveExport(request)
	}
	if approved {
		return core.ExportResponse{Approved: true}, nil
	}
	return core.ExportResponse{Approved: false}, err
}

func (r *rulesetUI) ApproveImport(request *core.ImportRequest) (core.ImportResponse, error) {
	// This cannot be handled by rules, requires setting a password
	// dispatch to next
	return r.next.ApproveImport(request)
}

func (r *rulesetUI) ApproveListing(request *core.ListRequest) (core.ListResponse, error) {
	jsonreq, err := json.Marshal(request)
	approved, err := r.checkApproval("ApproveListing", jsonreq, err)
	if err != nil {
		log.Info("Rule-based approval error, going to manual", "error", err)
		return r.next.ApproveListing(request)
	}
	if approved {
		return core.ListResponse{Accounts: request.Accounts}, nil
	}
	return core.ListResponse{}, err
}

func (r *rulesetUI) ApproveNewAccount(request *core.NewAccountRequest) (core.NewAccountResponse, error) {
	// This cannot be handled by rules, requires setting a password
	// dispatch to next
	return r.next.ApproveNewAccount(request)
}

func (r *rulesetUI) ShowError(message string) {
	log.Error(message)
	r.next.ShowError(message)
}

func (r *rulesetUI) ShowInfo(message string) {
	log.Info(message)
	r.next.ShowInfo(message)
}
func (r *rulesetUI) OnSignerStartup(info core.StartupInfo) {
	jsonInfo, err := json.Marshal(info)
	if err != nil {
		log.Warn("failed marshalling data", "data", info)
		return
	}
	r.next.OnSignerStartup(info)
	_, err = r.execute("OnSignerStartup", string(jsonInfo))
	if err != nil {
		log.Info("error occurred during execution", "error", err)
	}
}

func (r *rulesetUI) OnApprovedTx(tx manapi.SignTransactionResult) {
	jsonTx, err := json.Marshal(tx)
	if err != nil {
		log.Warn("failed marshalling transaction", "tx", tx)
		return
	}
	_, err = r.execute("OnApprovedTx", string(jsonTx))
	if err != nil {
		log.Info("error occurred during execution", "error", err)
	}
}
