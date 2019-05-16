/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// AccountBalance implements a simple chaincode to manage an asset
type AccountBalance struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *AccountBalance) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (a *AccountBalance) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	var result string
	var err error

	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryToken" {
		result, err = queryToken(APIstub, args)
	} else if function == "issueToken" {
		result, err = issueToken(APIstub, args)
	} else if function == "spendToken" {
		result, err = spendToken(APIstub, args)
	} else {
		return shim.Error("Invalid Smart Contract function name.")
	}
	

	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

func queryToken(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	accountBalanceAsBytes, err := APIstub.GetState(args[0])

	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if accountBalanceAsBytes == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}

	return string(accountBalanceAsBytes), nil
}

func issueToken(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	accountBalanceAsBytes, err := APIstub.GetState(args[0])

	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}

	if token, err := strconv.ParseFloat(args[1], 64); err == nil {
		if accountBalanceAsBytes == nil {

			err_tmp := APIstub.PutState(args[0], []byte(strconv.FormatFloat(token, 'f', 6, 64)))
		
			if err_tmp != nil {
				return "", fmt.Errorf("Failed to set asset: %s with error: %s", args[1], err_tmp)
			}
		
			return strconv.FormatFloat(token, 'f', 6, 64), nil
		} else {
			currentToken, parseError := strconv.ParseFloat(string(accountBalanceAsBytes), 64)
			
			if parseError != nil{
				return "", fmt.Errorf("Failed to parse current token error: %s", parseError)
			}

			err_tmp := APIstub.PutState(args[0], []byte(strconv.FormatFloat(token+currentToken, 'f', 6, 64)))
		
			if err_tmp != nil {
				return "", fmt.Errorf("Failed to set asset: %s with error: %s", args[1], err_tmp)
			}
		
			return strconv.FormatFloat(token+currentToken, 'f', 6, 64), nil
		}
	} else {
		return "", fmt.Errorf("Failed to parse token: %s with error: %s", args[1], err)
	}
}

func spendToken(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	accountBalanceAsBytes, err := APIstub.GetState(args[0])

	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if accountBalanceAsBytes == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}

	if token, err := strconv.ParseFloat(args[1], 64); err == nil {
		currentToken, parseError := strconv.ParseFloat(string(accountBalanceAsBytes), 64)
		
		if parseError != nil{
			return "", fmt.Errorf("Failed to parse current token error: %s", parseError)
		}

		if token > currentToken{
			return "", fmt.Errorf("The account balance is smaller than spending tokens")
		}

		err_tmp := APIstub.PutState(args[0], []byte(strconv.FormatFloat(currentToken-token, 'f', 6, 64)))
	
		if err_tmp != nil {
			return "", fmt.Errorf("Failed to update ledger")
		}
	
		return strconv.FormatFloat(currentToken-token, 'f', 6, 64), nil
	} else {
		return "", fmt.Errorf("Failed to parse token: %s with error: %s", args[1], err)
	}
}


func float64ToByte(f float64) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, f)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(AccountBalance)); err != nil {
		fmt.Printf("Error starting AccountBalance chaincode: %s", err)
	}
}
