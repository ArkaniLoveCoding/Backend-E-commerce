package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type MidtransSnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

func CreateSnapMidtrans (amount int, orderID int) (*MidtransSnapResponse, error) {
	payload := map[string]interface{}{
		"transactions_id": map[string]interface{}{
			"order_id": orderID,
			"amount": amount,
		},
	}

	request, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpRequest, err := http.NewRequest(
		"POST",
		os.Getenv("MIDTRANSBASEURL")+"/v1/snap/transactions",
		bytes.NewBuffer(request),
	)
	if err != nil {
		return nil, err
	}

	httpRequest.SetBasicAuth(os.Getenv("MIDTRANSSERVERKEY"), "")
	httpRequest.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var result MidtransSnapResponse
	json.NewDecoder(resp.Body).Decode(&resp)

	return &result, nil
}