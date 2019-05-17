/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

 package main

 import (
	 "encoding/json"
	 "fmt"
	 "strconv"
 
	 "github.com/hyperledger/fabric/core/chaincode/shim"
	 "github.com/hyperledger/fabric/protos/peer"
 )
 
 // AccountBalance implements a simple chaincode to manage an asset
 type AccountBalance struct {
 }

 // Response template
 type GetVoucherResponse struct {
	AccountID string
	CompanyID string
	SpendToken float64
	Voucher float64
}

type QueryTokenResponse struct {
	AccountID string
	AccountBalance float64
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
	 if function == "getVoucher" {
		 result, err = getVoucher(APIstub, args)
	 } else {
		 return shim.Error("Invalid Smart Contract function name.")
	 }
	 
	 if err != nil {
		 return shim.Error(err.Error())
	 }
 
	 // Return the result as success payload
	 return shim.Success([]byte(result))
 }


 func getVoucher(APIstub shim.ChaincodeStubInterface, args []string) (string, error) {
	 maxVoucher := 0.3
	 tokenToVoucherRate := 0.2
	 
	 var spendToken float64
	 var voucher float64
	 userID := args[0]
	 token, err := strconv.ParseFloat(args[1], 64)

	 if err != nil {
		 return "", fmt.Errorf("Failed to parse token: %s with error: %s", args[1], err)
	 }

	 chaincodeName := "contract_0"
	 invokeArgs := [][]byte{[]byte("queryToken"), []byte(userID)}
	 channel := "mychannel"

	 queryTokenResponse := APIstub.InvokeChaincode(chaincodeName, invokeArgs, channel)

	 if queryTokenResponse.GetStatus() == 500 {
		if err != nil {
			return "", fmt.Errorf("Failed to get current account balance: %s with error: %s", userID, err)
		}	 
	 }
	 
	var queryTokenResponsePayload QueryTokenResponse
	err = json.Unmarshal(queryTokenResponse.GetPayload(), &queryTokenResponsePayload)
	if err != nil {
		return "", fmt.Errorf("Failed to parse current account balance")
	}

	currentToken := queryTokenResponsePayload.AccountBalance

	if currentToken < token {
		return "", fmt.Errorf("The account balance is smaller than spending tokens")
	}

	voucher = token * tokenToVoucherRate

	if voucher > maxVoucher {
		spendToken = maxVoucher / tokenToVoucherRate
		voucher = maxVoucher
	} else {
		spendToken = token
	}

	chaincodeName = "contract_0"
	invokeArgs = [][]byte{[]byte("spendToken"), []byte(userID), []byte(strconv.FormatFloat(spendToken, 'f', 6, 64))}
	channel = "mychannel"

	spendTokenResponse := APIstub.InvokeChaincode(chaincodeName, invokeArgs, channel)

	if spendTokenResponse.GetStatus() == 500 {
	   if err != nil {
		   return "", fmt.Errorf("Failed to update current account balance: %s with error: %s", userID, err)
	   }	 
	}

	response := GetVoucherResponse{
		AccountID: args[0],
		CompanyID: "Company 1",
		SpendToken: spendToken,
		Voucher: voucher,
	}

	jsonData, _ := json.Marshal(response)
	return string(jsonData), nil
 }
 
 
 
 // main function starts up the chaincode in the container during instantiate
 func main() {
	 if err := shim.Start(new(AccountBalance)); err != nil {
		 fmt.Printf("Error starting AccountBalance chaincode: %s", err)
	 }
 }
 