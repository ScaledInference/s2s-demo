package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		header := w.Header()
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		header.Set("Access-Control-Allow-Headers", "Accept, Accept-Language, Content-Language, Content-Type")

		if req.Method == "OPTIONS" {
			_, _ = w.Write([]byte("OK"))
			return
		}

		// *** Parse the donation amount and ampToken from the body
		incomingBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: can't read body: %v", err.Error()), http.StatusBadRequest)
			return
		}

		var incomingMap map[string]interface{}
		if err := json.Unmarshal(incomingBody, &incomingMap); err != nil {
			http.Error(w, fmt.Sprintf("Error: can't unmarshall body: %v", err.Error()), http.StatusBadRequest)
			return
		}

		ampToken := incomingMap["ampToken"].(string)
		var amount float64
		if s, ok := incomingMap["amount"].(string); ok {
			if amount, err = strconv.ParseFloat(s, 10); err != nil {
				http.Error(w, fmt.Sprintf("Error: amount is not a real number: %v", err.Error()), http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "Error: amount is missing", http.StatusBadRequest)
			return
		}

		// POST outcome to amp.ai
		outgoingMap := map[string]interface{}{
			"ampToken": ampToken,
			"name":     "Donation",
			"properties": map[string]interface{}{
				"amount": amount,
			},
			"ts": time.Now().UnixNano() / 1000000,
		}

		observeUrl := "http://amp.ai/api/core/v2/observeWithToken"
		log.Println("Posting outcome request to ", observeUrl)
		log.Println("Request object: ", outgoingMap)
		outgoingBody, err := json.Marshal(outgoingMap)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: can't marshall body: %v", err.Error()), http.StatusInternalServerError)
			return
		}

		outgoingReq, err := http.NewRequest("POST", observeUrl, bytes.NewBuffer(outgoingBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(outgoingReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: while sending outcome event to backend: %v", err.Error()), http.StatusInternalServerError)
			return
		}

		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("bad observe response: %v", resp.Status), http.StatusInternalServerError)
			return
		}

		_, _ = w.Write([]byte("OK"))
	})

	log.Println("Starting metrics reporting server. Listening on port 8090 ...")
	if err := http.ListenAndServe(":8090", nil); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
