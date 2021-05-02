package crawler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

//GetRequest performs a GET request and return
//a string (the body of the response)
func GetRequest(target string) (string, error) {
	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	//Convert the body to type string
	sb := string(body)
	return sb, nil
}

//PostRequest performs a POST request and return
//a string (the body of the response)
//the map in the input should contains the data fields and values
//in this way for example:
//{ email: test@example.com, password: stupid_pwd }
func PostRequest(target string, data map[string]string) (string, error) {
	postBody, _ := json.Marshal(data)
	responseBody := bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request
	resp, err := http.Post(target, "application/json", responseBody)
	//Handle Error
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sb := string(body)
	return sb, nil
}

//HeadRequest performs a HEAD request and return
//a string (the body of the response)
func HeadRequest(target string) (string, error) {
	resp, err := http.Head(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sb := string(body)
	return sb, nil
}
