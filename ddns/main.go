package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type GetResponse struct {
	Errors   []any `json:"errors"`
	Messages []any `json:"messages"`
	Success  bool  `json:"success"`
	Result   []struct {
		Content   string    `json:"content"`
		Name      string    `json:"name"`
		Proxied   bool      `json:"proxied"`
		Type      string    `json:"type"`
		Comment   string    `json:"comment"`
		CreatedOn time.Time `json:"created_on"`
		ID        string    `json:"id"`
		Locked    bool      `json:"locked"`
		Meta      struct {
			AutoAdded bool   `json:"auto_added"`
			Source    string `json:"source"`
		} `json:"meta"`
		ModifiedOn time.Time `json:"modified_on"`
		Proxiable  bool      `json:"proxiable"`
		Tags       []string  `json:"tags"`
		TTL        int       `json:"ttl"`
		ZoneID     string    `json:"zone_id"`
		ZoneName   string    `json:"zone_name"`
	} `json:"result"`
	ResultInfo struct {
		Count      int `json:"count"`
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
}

type PatchData struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

var EMAIL string
var API_KEY string
var ZONE_ID string

func main() {

	err := godotenv.Load(".env")

	if err != nil {
		fmt.Println(err)
		return
	}

	EMAIL = os.Getenv("EMAIL")
	API_KEY = os.Getenv("API_KEY")
	ZONE_ID = os.Getenv("ZONE_ID")

	currentip := getipv6()
	previp, patchlist := fetchdnsrecords()

	if currentip != previp {
		fmt.Println("Ip addr changed Updating IP")
		for _, patchitems := range patchlist {
			patchdnsrecords(patchitems.Id, currentip, patchitems.Name)
		}
	} else {
		fmt.Println("Current ip has not changed")
	}

}

func fetchdnsrecords() (string, []PatchData) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", ZONE_ID)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Auth-Email", EMAIL)
	req.Header.Add("X-Auth-Key", API_KEY)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	u := GetResponse{}
	json.Unmarshal([]byte(body), &u)

	retarr := []PatchData{}

	for _, value := range u.Result {

		if value.Type == "AAAA" {
			retarr = append(retarr, PatchData{Id: value.ID, Name: value.Name, Type: value.Type})
		}

	}

	return u.Result[0].Content, retarr

}

func patchdnsrecords(dnsrecordid string, newip string, name string) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", ZONE_ID, dnsrecordid)
	patchdata := map[string]interface{}{
		"content": newip,
		"name":    name,
		"type":    "AAAA",
	}

	payload, _ := json.Marshal(patchdata)

	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Auth-Email", EMAIL)
	req.Header.Add("X-Auth-Key", API_KEY)

	http.DefaultClient.Do(req)

}

func getipv6() string {
	res, _ := http.Get("https://api6.ipify.org")

	data, _ := io.ReadAll(res.Body)

	defer res.Body.Close()

	return string(data)
}
