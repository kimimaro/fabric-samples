/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/*
 * The sample smart contract for documentation topic:
 * Writing Your First Blockchain Application
 */

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the habit structure, with 4 properties.  Structure tags are used by encoding/json library
type Habit struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Attendees []string `json:"attendees"`
	Owner     string   `json:"owner"`
}

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryHabit" {
		return s.queryHabit(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "createHabit" {
		return s.createHabit(APIstub, args)
	} else if function == "queryAllHabits" {
		return s.queryAllHabits(APIstub)
	} else if function == "changeHabitOwner" {
		return s.changeHabitOwner(APIstub, args)
	} else if function == "changeHabitAttendees" {
		return s.changeHabitAttendees(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryHabit(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(carAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	habits := []Habit{
		Habit{Name: "Running", Type: "Health", Attendees: []string{"Ruby", "Kathy"}, Owner: "Kathy"},
		Habit{Name: "English", Type: "Learning", Attendees: []string{"Kathy"}, Owner: "Kathy"},
		Habit{Name: "Workout", Type: "Health", Attendees: []string{"Kimi", "Rocky"}, Owner: "Kimi"},
		Habit{Name: "bark", Type: "Nature", Attendees: []string{"Ruby", "Rocky"}, Owner: "Rocky"},
		Habit{Name: "Blockchain", Type: "Learning", Attendees: []string{"Kimi", "Ruby", "Rocky"}, Owner: "Kimi"},
	}

	people := []Person{
		Person{Name: "Kimi", Age: 29},
		Person{Name: "Kathy", Age: 28},
		Person{Name: "Ruby", Age: 5},
		Person{Name: "Rocky", Age: 3},
	}

	i := 0
	for i < len(habits) {
		habitAsBytes, _ := json.Marshal(habits[i])
		APIstub.PutState("HABIT"+strconv.Itoa(i), habitAsBytes)
		i = i + 1
	}

	j := 0
	for j < len(people) {
		personAsBytes, _ := json.Marshal(people[j])
		APIstub.PutState("PERSON"+strconv.Itoa(j), personAsBytes)
		j = j + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createHabit(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	attendees := []string{args[3]}
	var habit = Habit{Name: args[1], Type: args[2], Attendees: attendees, Owner: args[4]}

	habitAsBytes, _ := json.Marshal(habit)
	APIstub.PutState(args[0], habitAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) queryAllHabits(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "HABIT0"
	endKey := "HABIT999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
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

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) changeHabitOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	habitAsBytes, _ := APIstub.GetState(args[0])
	habit := Habit{}

	json.Unmarshal(habitAsBytes, &habit)
	habit.Owner = args[1]

	habitAsBytes, _ = json.Marshal(habit)
	APIstub.PutState(args[0], habitAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) changeHabitAttendees(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	habitAsBytes, _ := APIstub.GetState(args[0])
	habit := Habit{}

	json.Unmarshal(habitAsBytes, &habit)

	attendees := append(habit.Attendees, args[1])
	habit.Attendees = attendees

	habitAsBytes, _ = json.Marshal(habit)
	APIstub.PutState(args[0], habitAsBytes)

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
