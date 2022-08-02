package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getData() (fact CatFact) {
	resp, err := http.Get("https://catfact.ninja/fact")

	if err == nil {
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			defer resp.Body.Close()
			return
		}

		unMarshalErr := json.Unmarshal(body, &fact)

		if unMarshalErr != nil {
			return CatFact{}
		}

		return
	}

	return
}

func main() {
	fmt.Println(getData().FactText)
}
