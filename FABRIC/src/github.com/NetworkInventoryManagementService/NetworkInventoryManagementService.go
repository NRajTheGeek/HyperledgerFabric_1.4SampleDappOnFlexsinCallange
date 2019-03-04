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

// NetworkInventoryManagementChaincode example simple Chaincode implementation
type NetworkInventoryManagementChaincode struct {
}

// ============================================================================================================================
// Structure of assets
// ============================================================================================================================

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
	err := shim.Start(new(NetworkInventoryManagementChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode - %s", err)
	}
}

// ============================================================================================================================
// Init - initialize the chaincode
// ============================================================================================================================
func (t *NetworkInventoryManagementChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("DataCircuit Store Channel Is Starting Up")
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
func (t *NetworkInventoryManagementChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println(" ")
	fmt.Println("starting invoke, for - " + function)

	// Handle different functions
	if function == "addNewDataCircuit" { //create a new marble
		return addNewDataCircuit(stub, args)
	} else if function == "allocateDataCircuitBandwidth" { //create a new marble
		return allocateDataCircuitBandwidth(stub, args)
	} else if function == "checkBandwithAllowanceOnCircuit" {
		return checkBandwithAllowanceOnCircuit(stub, args)
	} else if function == "queryDataCircuitBandwidthDataById" {
		return queryDataCircuitBandwidthDataById(stub, args)
	}

	// error out
	fmt.Println("Received unknown invoke function name - " + function)
	return shim.Error("Received unknown invoke function name - '" + function + "'")
}

// ============================================================================================================================
// Query - legacy function
// ============================================================================================================================
func (t *NetworkInventoryManagementChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call - Query()")
}

// ========================================================
// Input Sanitation - dumb input checking, look for empty strings
// ========================================================
func sanitize_arguments(strs []string) error {
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

func addNewDataCircuit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting addNewDataCircuit")

	if len(args) != 4 {
		fmt.Println("addNewDataCircuit(): Incorrect number of arguments. Expecting 4")
		return shim.Error("addNewDataCircuit(): Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	err1 := sanitize_arguments(args)
	if err1 != nil {
		return shim.Error("Cannot sanitize arguments")
	}

	dataCircuitID := args[0]
	fmt.Println(args)
	//check if marble id already exists
	dataCircuitAsBytes, err := stub.GetState(dataCircuitID)
	if err != nil { //this seems to always succeed, even if key didn't exist
		return shim.Error("error in finding DataCircuit for - " + dataCircuitID)
	}
	if dataCircuitAsBytes != nil {
		fmt.Println("This DataCircuit already exists - " + dataCircuitID)
		return shim.Error("This DataCircuit already exists - " + dataCircuitID) //all stop a marble by this id exists
	}

	dataCircuitObject, err := createDataCircuitObject(args[0:])
	if err != nil {
		errorStr := "addNewDataCircuit() : Failed Cannot create object buffer for write : " + args[0]
		fmt.Println(errorStr)
		return shim.Error(errorStr)
	}

	fmt.Println(dataCircuitObject)
	buff, err := dataCircuitToJSON(dataCircuitObject)
	if err != nil {
		return shim.Error("unable to convert DataCircuit to json")
	}

	err = stub.PutState(dataCircuitID, buff) //store marble with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end addNewDataCircuit")
	return shim.Success(nil)
}

func allocateDataCircuitBandwidth(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("starting allocateDataCircuitBandwidth")

	if len(args) != 2 {
		fmt.Println("addNewDataCircuit(): Incorrect number of arguments. Expecting 3 ")
		return shim.Error("addNewDataCircuit(): Incorrect number of arguments. Expecting 3 ")
	}

	//input sanitation
	err1 := sanitize_arguments(args)
	if err1 != nil {
		return shim.Error("Cannot sanitize arguments")
	}

	dataCircuitID := args[0]
	toAllocateBandwidth, _ := strconv.Atoi(args[1])
	fmt.Println(args)
	//check if marble id already exists
	dataCircuitAsBytes, err := stub.GetState(dataCircuitID)
	if err != nil { //this seems to always succeed, even if key didn't exist
		return shim.Error("error in finding DataCircuit for - " + dataCircuitID)
	}
	if dataCircuitAsBytes == nil {
		fmt.Println("This DataCircuit does not exists - " + dataCircuitID)
		return shim.Error("This DataCircuit does not exists - " + dataCircuitID) //all stop a marble by this id exists
	}

	dataCircuitObject, err := jsonToDataCircuit(dataCircuitAsBytes)

	if toAllocateBandwidth <= dataCircuitObject.UnallocatedBandwidth {
		dataCircuitObject.AllocatedBandwidth = dataCircuitObject.AllocatedBandwidth + toAllocateBandwidth
		dataCircuitObject.UnallocatedBandwidth = dataCircuitObject.UnallocatedBandwidth - toAllocateBandwidth

	} else {
		errorStr := "allocateDataCircuitBandwidth() : Failed Cannot allocate Data Circuit Bandwidth for write : " + args[0]
		fmt.Println(errorStr)
		return shim.Error(errorStr)
	}
	if err != nil {
		errorStr := "allocateDataCircuitBandwidth() : Failed Cannot allocate Data Circuit Bandwidth for write : " + args[0]
		fmt.Println(errorStr)
		return shim.Error(errorStr)
	}

	fmt.Println(dataCircuitObject)
	buff, err := dataCircuitToJSON(dataCircuitObject)
	if err != nil {
		return shim.Error("unable to convert DataCircuit to json")
	}

	err = stub.PutState(dataCircuitID, buff) //store marble with id as key
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end allocateDataCircuitBandwidth")
	return shim.Success(nil)
}

// CreateAssetObject creates an asset
func createDataCircuitObject(args []string) (DataCircuit, error) {
	var myDataCircuit DataCircuit

	fmt.Println(args)
	// Check there are 10 Arguments provided as per the the struct
	if len(args) != 4 {
		strErr := "createDataCircuitObject(): Incorrect number of arguments. Expecting 4 but got " + strconv.Itoa(len(args))
		fmt.Println(strErr)
		return myDataCircuit, errors.New(strErr)
	}

	ttlBandwidth, _ := strconv.Atoi(args[3])

	myDataCircuit = DataCircuit{args[0], args[1], args[2], false, ttlBandwidth, 0, ttlBandwidth, time.Now().Format("20060102150405")}
	return myDataCircuit, nil
}

func dataCircuitToJSON(eval DataCircuit) ([]byte, error) {

	djson, err := json.Marshal(eval)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return djson, nil
}

func jsonToDataCircuit(data []byte) (DataCircuit, error) {

	eval := DataCircuit{}
	err := json.Unmarshal([]byte(data), &eval)
	if err != nil {
		fmt.Println("Unmarshal failed : ", err)
		return eval, err
	}

	return eval, nil
}

// query callback representing the query of a chaincode
func checkBandwithAllowanceOnCircuit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("sarting the checkBandwithAllowanceOnCircuit() with the args: ")
	fmt.Println(args)
	fmt.Println("========================")
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	circuitID := args[0]

	// Get the state from the ledger
	circuitbytes, err := stub.GetState(circuitID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + circuitID + "\"}"
		return shim.Error(jsonResp)
	}

	if circuitbytes == nil {
		jsonResp := "{\"Error\":\"Nil data for " + circuitID + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"CircuitID\":\"" + circuitID + "\",\"data\":\"" + string(circuitbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)

	// respond back to BPm for this

	return shim.Success(circuitbytes)
}

// very important as it is required by the Answer chaincode to query
func queryDataCircuitBandwidthDataById(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	CircuitID := args[0]

	queryString := fmt.Sprintf("{\"selector\":{\"CircuitID\":\"%s\"}}", CircuitID)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
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
