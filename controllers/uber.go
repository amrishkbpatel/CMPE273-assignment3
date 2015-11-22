package uber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Estimates struct {
	Prices []Estimate `json:"prices"`
}

type Estimate struct {
	ProductId       string  `json:"product_id"`
	Estimate        string  `json:"estimate"`
	LowEstimate     int     `json:"low_estimate"`
	HighEstimate    int     `json:"high_estimate"`
	SurgeMultiplier float64 `json:"surge_multiplier"`
	Duration        int     `json:"duration"`
	Distance        float64 `json:"distance"`
	CurrencyCode    string  `json:"currency_code"`
	DisplayName     string  `json:"display_name"`
}

type UberOutput struct {
	Cost     int
	Duration int
	Distance float64
}

type ETA struct {
	Request_id      string  `json:"request_id"`
	Status          string  `json:"status"`
	Location        string  `json:"location"`
	ETA             int     `json:"eta"`
	SurgeMultiplier float64 `json:"surge_multiplier"`
}

func GetUberPrice(startLat, startLon, endLat, endLon string) UberOutput {
	client := &http.Client{}
	reqURL := fmt.Sprintf("https://sandbox-api.uber.com/v1/estimates/price?start_latitude=%s&start_longitude=%s&end_latitude=%s&end_longitude=%s&server_token=apkrMKlosrttkv_02okdjmHYlZrpmf33DlpdbiPP", startLat, startLon, endLat, endLon)
	fmt.Println("URL formed: " + reqURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error in sending req ", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error ", err)
	}

	var res Estimates
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println("error: ", err)
	}

	var uberOutput UberOutput
	uberOutput.Cost = res.Prices[0].LowEstimate
	uberOutput.Duration = res.Prices[0].Duration
	uberOutput.Distance = res.Prices[0].Distance

	return uberOutput

}

func GetUberEta(startLat, startLon, endLat, endLon string) int {

	var jsonStr = []byte(`{"start_latitude":"` + startLat + `","start_longitude":"` + startLon + `","end_latitude":"` + endLat + `","end_longitude":"` + endLon + `","product_id":"04a497f5-380d-47f2-bf1b-ad4cfdcb51f2"}`)
	reqURL := "https://sandbox-api.uber.com/v1/requests"
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "Bearer eypamskfjsdfnNKNLK9090NKNLNHdndhdmlwimss8s94n587mg06m00m3omlmujbuuppoaQOo055bddW8WoP4gTBc8gndgfgfgknmgkldfngdl-gmdlgfmdlgnjdlgmjdlgjndlkmcmvdlgmdlg9DGDFGF45ryryryFGfgfgfg9g9f9oE4Rtd4fgfg8J7urrrFFHHDfdfgfg5656gfgvpla987PaRGyJr7I9E3opdfjepoeyrrtrtg7T5I83333FsHtyLoPyfYgTgrereryHFBFASYHJMntfdeffhsInJlcXVlc3QiXSwic3ViIjoiNzcyMjZiYWMtMzJiMC00YzMzLWEwNWYtYWI0ODBjNTUyOTg3IiwsInJlcXVlc3QiXSwic3ViIjoiNzcyMjZiYWMtMzJiMC00YzMzLWEwNWYtYWI0ODBjNTUyOTg3Iiw2EN8bcx0eo3K-B51Wk29amm7-u8FV0l5mdQm1s5i7RNuI9t9pVh7t0To28OnvOsRCj5nIIz5t7ggZgPoed9mmVU7rndXDecjK4APBZArygEpHr7QDBXgxMkjnIlu_Nxftb8BDtcmVxdWVzdCJdLCJzdWIiOiJjZjg5YjZlMi04NTBkLTRhZjktYmU3MC05OTJmMDcyOTU0Y2YiLCJpc3MiOiJ1YmVyLXVzMSIsImp0aSI6ImQ4OGRiYTJhLTE0OWUtNDc5MC04ODJkLTMxYjIzZGM2NWZlYiIsImV4-Q")

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error: ", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error: ", err)
	}

	var res ETA
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println("error: ", err)
	}
	eta := res.ETA
	return eta

}
