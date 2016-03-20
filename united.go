package united

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const shipmentTrackingURL = "http://www.unitedvanlines.com/vanlines/api/shipmentTracking-v1"

type shipmentTrackingRequest struct {
	Name           string
	OrderNumber    string
	SubsidiaryCode string
}

type Header struct {
	Response string `json:"response"`
}

type Van struct {
	LastReportedCity     string `json:"lastReportedCity"`
	LastReportedState    string `json:"lastReportedState"`
	LastReportedDateTime string `json:"lastReportedDateTime"`
	TrackingCity         string `json:"trackingCity"`
	TrackingDate         string `json:"trackingDate"`
	TrackingDirection    string `json:"trackingDirection"`
	TrackingMiles        string `json:"trackingMiles"`
	TrackingState        string `json:"trackingState"`
	TrackingTime         string `json:"trackingTime"`
}

type Shipment struct {
	ActualLoadDate string `json:"actualLoadDate"`
	Van            `json:"van"`
}

type ShipmentTrackingResponse struct {
	Header       `json:"header"`
	OrderNumbers []string `json:"orderNumbers"`
	Shipment     `json:"shipment"`
}

//unpacks the null() wrapper
func cleanResponse(resp []byte) []byte {
	return resp[5 : len(resp)-1]
}

func GetTrackingUpdate(name, orderNumber string) (ShipmentTrackingResponse, error) {
	var response ShipmentTrackingResponse

	requestJSON, err := json.Marshal(shipmentTrackingRequest{
		Name:           name,
		OrderNumber:    orderNumber,
		SubsidiaryCode: "U",
	})

	if err != nil {
		return response, err
	}

	query := url.QueryEscape(fmt.Sprintf("{\"ShipmentTracking\": %s}", requestJSON))
	queryURL := fmt.Sprintf("%v?data=%v", shipmentTrackingURL, query)

	resp, err := http.Get(queryURL)
	if err != nil {
		return response, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(cleanResponse(body), &response)
	return response, err
}
