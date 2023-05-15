/*
Copyright © 2023 Patrick Hermann patrick.hermann@sva.de
*/

package wled

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"
)

type WledStatus struct {
	Brightness int
	Segment    int
	Color      string
	Fx         int
}

var bodyData = `{
	"bri":{{ .Brightness }},
	"seg":[{"id":{{ .Segment }},"col":[{{ .Color }}],"fx":{{ .Fx }}}]}
		}`

func ControllWled(wledUrl string, updatedStatus WledStatus) {

	wledUrl = "http://" + wledUrl + "/json/state"
	fmt.Println("WLED URL:", wledUrl)

	var jsonData = []byte(renderBodyData(updatedStatus))

	fmt.Println(renderBodyData(updatedStatus))

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

func renderBodyData(updatedStatus WledStatus) string {

	tmpl, err := template.New("wledstatus").Parse(bodyData)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, updatedStatus)

	if err != nil {
		fmt.Println(err)
	}

	return buf.String()

}
