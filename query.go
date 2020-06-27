package netbox

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	clog "github.com/coredns/coredns/plugin/pkg/log"
)

type Record struct {
	Address  string `json:"address"`
	HostName string `json:"dns_name,omitempty"`
}

type RecordsList struct {
	Records []Record `json:"results"`
}

func query(url, token, dns_name string) string {

	records := RecordsList{}
	client := &http.Client{}
	var resp *http.Response

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))

	for i := 1; i <= 10; i++ {
		resp, err = client.Do(req)

		if err != nil {
			clog.Fatalf("HTTP Error %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			break
		}

		time.Sleep(1 * time.Second)
	}
	// TODO: check that we got status code 200
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		clog.Fatalf("Error reading body %v", err)
	}

	jsonAns := string(body)
	err = json.Unmarshal([]byte(jsonAns), &records)
	if err != nil {
		clog.Fatalf("could not unmarshal response %v", err)
	}

	return strings.Split(records.Records[0].Address, "/")[0]
}
