package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// OrderManagementChaincode example simple Chaincode implementation
type OrderManagementChaincode struct {
}

func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

// ============================================================================================================================
// Asset Definitions - The ledger will store answers with hash id and cid
// ============================================================================================================================
type Order struct {
	OrderID        string `json:"QuestionHashID"`
	DataCircuitID  string `json:"QuestionerID"`
	OrderBandwidth int    `json:"OrderBandwidth"`
	OperatorID     string `json:"OperatorID"`
	OrderSatus     bool   `json:"OrderSatus"`
	CreatedOn      string `json:"CreatedOn'`
}

// Internal data maps
type DataCircuit struct {
	CircuitID            string `json:"CircuitID"`
	CircuitNetwork       string `json:"CircuitNetwork"`
	ProviderID           string `json:"ProviderID"`
	IsConfigured         bool   `json:"IsConfigured"`
	TotalBandwidth       int    `json:"TotalBandwidth"`
	AllocatedBandwidth   int    `json:"AllowedBandwidth"`
	UnallocatedBandwidth int    `json:"unallowedBandwidth"`
	CreatedOn            string `json:"CreatedOn'`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(OrderManagementChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode - %s", err)
	}
}

// ============================================================================================================================
// Init - initialize the chaincode
// ============================================================================================================================
func (t *OrderManagementChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Answer Store Channel Is Starting Up")
	funcName, args := stub.GetFunctionAndParameters()
	var err error
	txId := stub.GetTxID()

	fmt.Println("  Init() is running")
	fmt.Println("  Transaction ID: ", txId)
	fmt.Println("  GetFunctionAndParameters() function: ", funcName)
	fmt.Println("  GetFunctionAndParameters() args count: ", len(args))
	fmt.Println("  GetFunctionAndParameters() args found: ", args)

	// expecting 1 arg for instantiate or upgrade
	if len(args) == 2 {
		fmt.Println("  GetFunctionAndParameters() : Number of arguments", len(args))
	}

	err = stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(err.Error()) //self-test fail
	}

	fmt.Println("Ready for action") //self-test pass
	return shim.Success(nil)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *OrderManagementChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println(" ")
	fmt.Println("starting invoke, for - " + function)

	// Handle different functions
	if function == "prepareOrder" { //create a new marble
		return prepareOrder(stub, args)
	} else if function == "getOrder" { //update_answer
		return getOrder(stub, args)
	}
	// error out
	fmt.Println("Received unknown invoke function name - " + function)
	return shim.Error("Received unknown invoke function name - '" + function + "'")
}

// ============================================================================================================================
// Query - legacy function
// ============================================================================================================================
func (t *OrderManagementChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call - Query()")
}

// ============================================================================================================================
// Get Answer - get a answer asset from ledger
// ============================================================================================================================
func getOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	fmt.Println("starting prepareOrder")

	if len(args) != 2 {
		fmt.Println("initAnswer(): Incorrect number of arguments. Expecting 2")
		return shim.Error("intAnswer(): Incorrect number of arguments. Expecting 2")
	}

	//input sanitation
	err1 := sanitizeArguments(args)
	if err1 != nil {
		return shim.Error("Cannot sanitize arguments")
	}
	ANCSChaincode := args[0]
	orderID := args[1]

	fmt.Println("========================= recieved args ==========================")
	fmt.Println(args)

	// ==================================== check the valid question ===========================================
	channelID := ""
	chainCodeToCall := ANCSChaincode //"questions2"
	functionName := "getOrder"

	queryArgs := toChaincodeArgs(functionName, orderID)

	response := stub.InvokeChaincode(chainCodeToCall, queryArgs, channelID)
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return shim.Error("error in finding order for - " + orderID)
	}
	orderBytes := response.Payload

	str := fmt.Sprintf("%s", orderBytes)
	fmt.Println("string is " + str)

	return shim.Success(orderBytes)
}

func prepareOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting prepareOrder")

	if len(args) != 7 {
		fmt.Println("initAnswer(): Incorrect number of arguments. Expecting 7")
		return shim.Error("intAnswer(): Incorrect number of arguments. Expecting 7")
	}

	//input sanitation
	err1 := sanitizeArguments(args)
	if err1 != nil {
		return shim.Error("Cannot sanitize arguments")
	}

	BPMChaincode := args[0]
	NIMSChaincode := args[1]
	ANCSChaincode := args[2]

	orderID := args[3]
	operatorID := args[4]
	dataCircuitID := args[5]
	orderBandwidth := args[6]

	fmt.Println("========================= recieved args ==========================")
	fmt.Println(args)

	// ==================================== check the valid question ===========================================
	channelID := ""
	chainCodeToCall := BPMChaincode //"questions2"
	functionName := "checkOnNIMSAndRespond"
	dataCircuitIDAsQueryKey := dataCircuitID
	operatorIDToProcess := operatorID
	orderBandwidthToProcess := orderBandwidth

	queryArgs := toChaincodeArgs(functionName, BPMChaincode, NIMSChaincode, ANCSChaincode, dataCircuitIDAsQueryKey, orderBandwidthToProcess, orderID, operatorIDToProcess)

	response := stub.InvokeChaincode(chainCodeToCall, queryArgs, channelID)
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to prepare order. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return shim.Error(errStr)
	}

	// now send for testing and cabling

	fmt.Println("- end submitAnswer")
	return shim.Success(nil)
}

func queryOtherChaincodeByKeyOnly(stub shim.ChaincodeStubInterface, args []string) (pb.Response, error) {

	fmt.Println("starting thumbsUpToAnswer")
	// var arr []byte

	if len(args) != 4 {
		return shim.Error(""), errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	err := sanitizeArguments(args)
	if err != nil {
		return shim.Error(""), errors.New("error in sanitization")
	}
	channelID := args[0]
	chainCodeToCall := args[1]
	functionName := args[2]
	queryKey := args[3]

	queryArgs := toChaincodeArgs(functionName, queryKey)
	response := stub.InvokeChaincode(chainCodeToCall, queryArgs, channelID)
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return shim.Error(""), errors.New(errStr)
	}
	bytesResponse := response.Payload

	return shim.Success(bytesResponse), nil
}

// for thumbsup first validate the registered evaluator by evaluator secret from the evaluator chaincode
// then allow the evaluator to do a thumsup against an answer hash id
// iff the evaluator has a tech reputation more than 1000

// ====================================================== Private Library ====================================================

func sanitizeArguments(strs []string) error {
	for i, val := range strs {
		if len(val) <= 0 {
			return errors.New("Argument " + strconv.Itoa(i) + " must be a non-empty string")
		}
		// if len(val) > 32 {
		// 	return errors.New("Argument " + strconv.Itoa(i) + " must be <= 32 characters")
		// }
	}
	return nil
}

func JSONtoOrder(data []byte) (Order, error) {

	order := Order{}
	err := json.Unmarshal([]byte(data), &order)
	if err != nil {
		fmt.Println("Unmarshal failed : ", err)
		return order, err
	}

	return order, nil
}
