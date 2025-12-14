package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type MidtransSnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

func CreateSnapMidtrans(amount int, orderID int) (*MidtransSnapResponse, error) {
	payload := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     fmt.Sprintf("ORDER-%d", orderID),
			"gross_amount": amount,
		},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	godotenv.Load()

	req, err := http.NewRequest(
		"POST",
		os.Getenv("MIDTRANSBASEURL")+"/snap/v1/transactions",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(os.Getenv("MIDTRANSSERVERKEY"), "")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Println("MIDTRANS STATUS:", resp.StatusCode)
	fmt.Println("MIDTRANS BODY:", string(bodyBytes))

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("midtrans error: %s", string(bodyBytes))
	}

	var result MidtransSnapResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
