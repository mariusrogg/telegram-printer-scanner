package main

import (
	"fmt"
	"io"
	"log"
)

type scanner struct {
	endpoint  string
	functions []scannerFunction
	deviceId  string
}

func newScanner(endpoint string, functions []scannerFunction, deviceId string) *scanner {
	return &scanner{
		endpoint:  endpoint,
		functions: functions,
		deviceId:  deviceId,
	}
}

func (scanner scanner) getFunction(functionName string) *scannerFunction {
	for _, function := range scanner.functions {
		if function.name == functionName {
			return &function
		}
	}
	return nil
}

func (scanner scanner) scan(functionName string, callback func(file io.ReadCloser, fileName string)) {
	fmt.Println("Received requirement to scan " + functionName)
	function := scanner.getFunction(functionName)
	if function == nil {
		log.Fatal("Function not found: " + functionName)
	}

	function.scan(scanner.endpoint, scanner.deviceId, callback)
}
