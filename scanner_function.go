package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ScannerSource string

const (
	flatbed ScannerSource = "Flatbed"
	adf     ScannerSource = "ADF"
)

var scannerSource = map[ScannerSource]string{
	flatbed: string(flatbed),
	adf:     string(adf),
}

func (ss ScannerSource) String() string {
	return scannerSource[ss]
}

type ScannerMode string

const (
	color ScannerMode = "Color"
	gray  ScannerMode = "Gray"
)

var scannerMode = map[ScannerMode]string{
	color: string(color),
	gray:  string(gray),
}

func (ss ScannerMode) String() string {
	return scannerMode[ss]
}

type ScannerTarget string

const (
	telegram  ScannerTarget = "telegram"
	paperless ScannerTarget = "paperless"
)

var scannerTarget = map[ScannerTarget]string{
	telegram:  string(telegram),
	paperless: string(paperless),
}

func (ss ScannerTarget) String() string {
	return scannerTarget[ss]
}

type ScannerFunction struct {
	mode   ScannerMode
	source ScannerSource
	target ScannerTarget
}

type scanBody struct {
	Params struct {
		DeviceID       string `json:"deviceId"`
		Top            int    `json:"top"`
		Left           int    `json:"left"`
		Width          int    `json:"width"`
		Height         int    `json:"height"`
		PageWidth      int    `json:"pageWidth"`
		PageHeight     int    `json:"pageHeight"`
		Resolution     int    `json:"resolution"`
		Mode           string `json:"mode"`
		Source         string `json:"source"`
		AdfMode        string `json:"adfMode"`
		Brightness     int    `json:"brightness"`
		Contrast       int    `json:"contrast"`
		DynamicLineart bool   `json:"dynamicLineart"`
		Ald            string `json:"ald"`
	} `json:"params"`
	Filters  []string `json:"filters"`
	Pipeline string   `json:"pipeline"`
	Batch    string   `json:"batch"`
	Index    int      `json:"index"`
}

func newScanBody(function ScannerFunction, scannerId string) *scanBody {
	batch := "none"
	if function.source == adf {
		batch = "auto"
	}
	body := scanBody{
		Params: struct {
			DeviceID       string "json:\"deviceId\""
			Top            int    "json:\"top\""
			Left           int    "json:\"left\""
			Width          int    "json:\"width\""
			Height         int    "json:\"height\""
			PageWidth      int    "json:\"pageWidth\""
			PageHeight     int    "json:\"pageHeight\""
			Resolution     int    "json:\"resolution\""
			Mode           string "json:\"mode\""
			Source         string "json:\"source\""
			AdfMode        string "json:\"adfMode\""
			Brightness     int    "json:\"brightness\""
			Contrast       int    "json:\"contrast\""
			DynamicLineart bool   "json:\"dynamicLineart\""
			Ald            string "json:\"ald\""
		}{
			DeviceID:       scannerId,
			Top:            0,
			Left:           0,
			Width:          215,
			Height:         297,
			PageWidth:      215,
			PageHeight:     297,
			Resolution:     200,
			Mode:           string(function.mode),
			Source:         string(function.source),
			AdfMode:        "Simplex",
			Brightness:     0,
			Contrast:       0,
			DynamicLineart: false,
			Ald:            "yes",
		},
		Filters:  []string{},
		Pipeline: "PDF (TIF | @:pipeline.uncompressed)",
		Batch:    batch,
		Index:    0,
	}
	return &body
}

type scanResponseBody struct {
	Image any `json:"image"`
	Index int `json:"index"`
	File  struct {
		Fullname     string    `json:"fullname"`
		Extension    string    `json:"extension"`
		LastModified time.Time `json:"lastModified"`
		Size         int64     `json:"size"`
		SizeString   string    `json:"sizeString"`
		IsDirectory  bool      `json:"isDirectory"`
		Name         string    `json:"name"`
		Path         string    `json:"path"`
	} `json:"file"`
}

func (function ScannerFunction) scan(endpoint string, scannerId string) (io.ReadCloser, string, error) {
	var scanClientWithTimeout = &http.Client{
		Timeout: time.Minute * 20,
	}
	fmt.Println("Starting scan")
	var err error
	body := newScanBody(function, scannerId)
	marshalled, err := json.Marshal(body)
	if err != nil {
		fmt.Println("Cannot encode JSON: " + err.Error())
		return nil, "", err
	}
	resp, err := scanClientWithTimeout.Post(endpoint+"/api/v1/scan", "application/json", bytes.NewReader(marshalled))
	if err != nil {
		fmt.Println("Post failed: " + err.Error())
		return nil, "", err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Post failed with status code: " + resp.Status)
		if resp.StatusCode == http.StatusInternalServerError {
			fmt.Printf("Trying to reload scanners\n")
			req, err := http.NewRequest(http.MethodDelete, endpoint+"/api/v1/context", nil)
			if err != nil {
				fmt.Println("Could not create delete request for scanners")
				return nil, "", err
			}
			client := &http.Client{}
			fmt.Printf("Delete scanners\n")
			resp, err = client.Do(req)
			if err != nil {
				fmt.Println("Could not delete scanners " + err.Error())
				return nil, "", err
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Println("Failed to delete scanners: " + resp.Status)
				return nil, "", fmt.Errorf("failed to delete scanners: %s", resp.Status)
			}
			fmt.Printf("Get scanners\n")
			resp, err = http.Get(endpoint + "/api/v1/context")
			if err != nil {
				fmt.Println("Failed to reload scanners " + err.Error())
				return nil, "", err
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Println("Could not reload scanners: " + resp.Status)
				return nil, "", fmt.Errorf("could not reload scanners: %s", resp.Status)
			}
			fmt.Printf("Retry scan\n")
			resp, err = scanClientWithTimeout.Post(endpoint+"/api/v1/scan", "application/json", bytes.NewReader(marshalled))
			if err != nil {
				fmt.Println("Post failed: " + err.Error())
				return nil, "", err
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Println("Post failed with status code: " + resp.Status)
				return nil, "", fmt.Errorf("post failed with status code: %s", resp.Status)
			}
		} else {
			return nil, "", fmt.Errorf("post failed with status code: %s", resp.Status)
		}
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read response failed: " + err.Error())
		return nil, "", err
	}
	var result scanResponseBody
	err = json.Unmarshal(respBody, &result)
	fmt.Printf("Result (%s): \n", resp.Status)
	j, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(j))
	if err != nil {
		fmt.Println("Cannot unmarshal JSON: " + err.Error())
		return nil, "", err
	}
	return function.getScannedFile(result.File.Name, endpoint)

}

func (function ScannerFunction) getScannedFile(fileName string, endpoint string) (io.ReadCloser, string, error) {
	fmt.Printf("Trying to get file %s\n", fileName)
	resp, err := http.Get(endpoint + "/api/v1/files/" + fileName)
	if err == nil {
		return resp.Body, fileName, nil
	}
	return nil, fileName, err
}
