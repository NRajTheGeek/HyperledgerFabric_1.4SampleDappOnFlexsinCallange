package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// BusinessProcessManagementChaincode example simple Chaincode implementation
type BusinessProcessManagementChaincode struct {
}

func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

// ============================================================================================================================
// Asset Definitions - The ledger will store questions with hash id and cid
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
	err := shim.Start(new(BusinessProcessManagementChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode - %s", err)
	}
}

// ============================================================================================================================
// Init - initialize the chaincode
// ============================================================================================================================
func (t *BusinessProcessManagementChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Question Store Channel Is Starting Up")
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
	// this is a very simple test. let's write to the ledger and error out on any errors
	// it's handy to read this right away to verify network is healthy if it wrote the correct value
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
func (t *BusinessProcessManagementChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println(" ")
	fmt.Println("starting invoke, for - " + function)

	// Handle different functions
	if function == "checkOnNIMSAndRespond" { //create a new marble
		return checkOnNIMSAndRespond(stub, args)
	}

	// error out
	fmt.Println("Received unknown invoke function name - " + function)
	return shim.Error("Received unknown invoke function name - '" + function + "'")
}

// ============================================================================================================================
// Query - legacy function
// ============================================================================================================================
func (t *BusinessProcessManagementChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call - Query()")
}

// ============================================================================================================================
// Get Question - get a question asset from ledger
// ============================================================================================================================

func checkOnNIMSAndRespond(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting submitQuestion")

	if len(args) != 6 {
		fmt.Println("initQuestion(): Incorrect number of arguments. Expecting 6")
		return shim.Error("intQuestion(): Incorrect number of arguments. Expecting 6")
	}

	//input sanitation
	err1 := sanitizeArguments(args)
	if err1 != nil {
		return shim.Error("Cannot sanitize arguments")
	}

	NIMSChaincode := args[0]
	ANCSChaincode := args[1]
	dataCircuitIDAsQueryKey := args[2]
	orderBandwidthToProcess, _ := strconv.Atoi(args[3])
	OrderID := args[4]
	operatorIDToProcess := args[5]

	//===================================================================================

	channelId := ""
	chainCodeToCall := NIMSChaincode //"questions2"
	functionName := "checkBandwithAllowanceOnCircuit"

	queryArgs := toChaincodeArgs(functionName, dataCircuitIDAsQueryKey)

	response := stub.InvokeChaincode(chainCodeToCall, queryArgs, channelId)
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", err.Error())
		fmt.Printf(errStr)
		return shim.Error("error in finding evaluator for - " + dataCircuitIDAsQueryKey)
	}
	circuitDataBytes := response.Payload

	str := fmt.Sprintf("%s", circuitDataBytes)
	fmt.Println("string is " + str)

	circuitData, err := JSONtoCircuitData(circuitDataBytes)
	if err != nil { //this seems to always succeed, even if key didn't exist
		fmt.Println("Error in unmarshelling - " + dataCircuitIDAsQueryKey)
		return shim.Error("Error in unmarshelling - " + dataCircuitIDAsQueryKey)
	}
	fmt.Println("captured circuit data ")
	fmt.Println(circuitData)

	fmt.Println("======================= orderBandwidthToProcess: ")
	fmt.Println(orderBandwidthToProcess)
	fmt.Println("======================= circuitData.UnallocatedBandwidth: ")
	fmt.Println(circuitData.UnallocatedBandwidth)
	fmt.Println("==========================================================")

	if orderBandwidthToProcess <= circuitData.UnallocatedBandwidth {
		// then it auto triggers the signal to Automatic Network Configuration Engine
		// to assign and configure it to a particular network according to client’s demand

		channelId = ""
		chainCodeToCall = ANCSChaincode //"questions2"
		functionName = "completeOrder"
		orderIDToProcess := OrderID
		DataCircuitID := dataCircuitIDAsQueryKey
		OrderBandwidth := strconv.Itoa(orderBandwidthToProcess)
		OperatorID := operatorIDToProcess

		queryArgs = toChaincodeArgs(functionName, orderIDToProcess, DataCircuitID, OrderBandwidth, OperatorID)

		response = stub.InvokeChaincode(chainCodeToCall, queryArgs, channelId)
		if response.Status != shim.OK {
			errStr := fmt.Sprintf("Failed to query chaincode. Got error: %s", err.Error())
			fmt.Printf(errStr)
			return shim.Error("error in finding evaluator for - " + dataCircuitIDAsQueryKey)
		}

	} else {
		errStr := fmt.Sprintf("Required bandwidth is out of allowance range: ", err.Error())
		fmt.Printf(errStr)
		return shim.Error("Required bandwidth is out of allowance range:  " + dataCircuitIDAsQueryKey)
	}

	fmt.Println("- end checkOnNIMSAndRespond")
	return shim.Success(nil)
}

// =========================================== Private Libraries ========================================================

// ========================================================
// Input Sanitation - dumb input checking, look for empty strings
// ========================================================
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

func CircuitDatatoJSON(dc DataCircuit) ([]byte, error) {

	fmt.Println("dc before being marshelled")
	fmt.Println(dc)

	djson, err := json.Marshal(dc)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return djson, nil
}

func JSONtoCircuitData(data []byte) (DataCircuit, error) {

	dc := DataCircuit{}
	err := json.Unmarshal([]byte(data), &dc)
	if err != nil {
		fmt.Println("Unmarshal failed : ", err)
		return dc, err
	}

	return dc, nil
}

func JSONtoOrder(data []byte) (Order, error) {

	dc := Order{}
	err := json.Unmarshal([]byte(data), &dc)
	if err != nil {
		fmt.Println("Unmarshal failed : ", err)
		return dc, err
	}

	return dc, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
