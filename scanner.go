package main

import (
	"slices"
)

var Scanner scanner

type scanner struct {
	endpoint  string
	functions []ScannerFunction
	deviceId  string
}

func newScanner(endpoint string, functions []ScannerFunction, deviceId string) *scanner {
	return &scanner{
		endpoint:  endpoint,
		functions: functions,
		deviceId:  deviceId,
	}
}

func (scanner scanner) getFunctions() []ScannerFunction {
	return scanner.functions
}

func (scanner scanner) getTargets() []ScannerTarget {
	targets := []ScannerTarget{}
	for _, function := range scanner.functions {
		if !slices.Contains(targets, function.target) {
			targets = append(targets, function.target)
		}
	}
	return targets
}

func (scanner scanner) getSources(target ScannerTarget) []ScannerSource {
	sources := []ScannerSource{}
	for _, function := range scanner.functions {
		if function.target == target {
			if !slices.Contains(sources, function.source) {
				sources = append(sources, function.source)
			}
		}
	}
	return sources
}

func (scanner scanner) getModes(target ScannerTarget, source ScannerSource) []ScannerMode {
	modes := []ScannerMode{}
	for _, function := range scanner.functions {
		if function.target == target && function.source == source {
			if !slices.Contains(modes, function.mode) {
				modes = append(modes, function.mode)
			}
		}
	}
	return modes
}

func (scanner scanner) getFunction(target ScannerTarget, source ScannerSource, mode ScannerMode) *ScannerFunction {
	for _, function := range scanner.functions {
		if (function.target == target || target == "") && (function.source == source || source == "") && (function.mode == mode || mode == "") {
			return &function
		}
	}
	return nil
}
