package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// AutomaticNetworkConfigurationChaincode example simple Chaincode implementation
type AutomaticNetworkConfigurationChaincode struct {
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
	err := shim.Start(new(AutomaticNetworkConfigurationChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode - %s", err)
	}
}

// ============================================================================================================================
// Init - initialize the chaincode
// ============================================================================================================================
func (t *AutomaticNetworkConfigurationChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
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
func (t *AutomaticNetworkConfigurationChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println(" ")
	fmt.Println("starting invoke, for - " + function)

	// Handle different functions
	if function == "completeOrder" { //create a new marble
		return completeOrder(stub, args)
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
func (t *AutomaticNetworkConfigurationChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call - Query()")
}

// ============================================================================================================================
// Get Answer - get a answer asset from ledger
// ============================================================================================================================
func getOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var order Order
	if len(args) != 1 {
		fmt.Println("initAnswer(): Incorrect number of arguments. Expecting 1")
		return shim.Error("intAnswer(): Incorrect number of arguments. Expecting 1")
	}

	//input sanitation
	err1 := sanitizeArguments(args)
	if err1 != nil {
		return shim.Error("Cannot sanitize arguments")
	}
	orderID := args[0]

	fmt.Println("========================= recieved args ==========================")
	fmt.Println(args)

	orderBytes, err := stub.GetState(orderID) //getState retreives a key/value from the ledger
	if err == nil {                           //this seems to always succeed, even if key didn't exist
		return shim.Error("Failed to find marble - " + orderID)
	}

	if orderBytes == nil { //test if marble is actually here or just nil
		return shim.Error("Answer does not exist - " + orderID)
	}

	err = json.Unmarshal([]byte(orderBytes), &order)
	if err != nil {
		fmt.Println("Unmarshal failed : ", err)
		return shim.Error("unable to unmarshall")
	}

	str := fmt.Sprintf("%s", orderBytes)
	fmt.Println("string is " + str)

	return shim.Success(orderBytes)
}

func completeOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting completeOrder")

	if len(args) != 4 {
		fmt.Println("initAnswer(): Incorrect number of arguments. Expecting 4")
		return shim.Error("intAnswer(): Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	err1 := sanitizeArguments(args)
	if err1 != nil {
		return shim.Error("Cannot sanitize arguments")
	}

	OrderID := args[0]

	fmt.Println("========================= recieved args ==========================")
	fmt.Println(args)

	// ===================================== save order into ledger ============================================
	orderAsBytes, err := stub.GetState(OrderID)
	if err != nil { //this seems to always succeed, even if key didn't exist
		return shim.Error("error in finding Order for - " + OrderID)
	}
	if orderAsBytes == nil {
		fmt.Println("This Order does not exists - " + OrderID)
		return shim.Error("This Order does not exists - " + OrderID) //all stop a marble by this id exists
	}

	orderObject, err := CreateOrderObject(args[0:])
	if err != nil {
		errorStr := "addNewDataCircuit() : Failed Cannot create object buffer for write : " + args[0]
		fmt.Println(errorStr)
		return shim.Error(errorStr)
	}

	fmt.Println(orderObject)
	buff, err := OrderToJSON(orderObject)
	if err != nil {
		return shim.Error("unable to convert DataCircuit to json")
	}

	err = stub.PutState(OrderID, buff) //store marble with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end completeOrder")
	return shim.Success(nil)
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
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

// CreateAssetObject creates an asset
func CreateOrderObject(args []string) (Order, error) {
	var myOrder Order

	// Check there are 10 Arguments provided as per the the struct
	if len(args) != 4 {
		fmt.Println("CreateAnswerObject(): Incorrect number of arguments. Expecting 4")
		return myOrder, errors.New("CreateAnswerObject(): Incorrect number of arguments. Expecting 4")
	}

	orderBandwidth, _ := strconv.Atoi(args[2])

	myOrder = Order{args[0], args[1], orderBandwidth, args[3], true, time.Now().Format("20060102150405")}
	return myOrder, nil
}

func OrderToJSON(ans Order) ([]byte, error) {

	djson, err := json.Marshal(ans)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return djson, nil
}

func JSONtoOrder(data []byte) (Order, error) {

	eval := Order{}
	err := json.Unmarshal([]byte(data), &eval)
	if err != nil {
		fmt.Println("Unmarshal failed : ", err)
		return eval, err
	}

	return eval, nil
}

func contains(techRepuArray []string, match string) bool {
	flag := false
	for _, data := range techRepuArray {
		if data == match {
			flag = true
			break
		}
	}
	return flag
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
