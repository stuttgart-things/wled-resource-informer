/*
Copyright © 2023 Patrick Hermann patrick.hermann@sva.de
*/

package wled

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func ControllWled(wledUrl string) {

	// wledUrl := "http://wled-87552c.local/json/state"
	fmt.Println("WLED URL:", wledUrl)

	var jsonData = []byte(`{
	"on": "t",
	"v": true
		}`)

	request, error := http.NewRequest("POST", wledUrl, bytes.NewBuffer(jsonData))
	if error != nil {
		panic(error)
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error2 := client.Do(request)
	if error2 != nil {
		panic(error)
	}

	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))
}
