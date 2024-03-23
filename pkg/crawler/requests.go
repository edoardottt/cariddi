/*
==========
Cariddi
==========

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see http://www.gnu.org/licenses/.

	@Repository:  https://github.com/edoardottt/cariddi

	@Author:      edoardottt, https://www.edoardoottavianelli.it

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package crawler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// GetRequest performs a GET request and return
// a string (the body of the response).
func GetRequest(target string) (string, error) {
	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// Convert the body to type string
	sb := string(body)

	return sb, nil
}

// PostRequest performs a POST request and return
// a string (the body of the response)
// the map in the input should contains the data fields and values
// in this way for example:
// { email: test@example.com, password: stupid_pwd }.
func PostRequest(target string, data map[string]string) (string, error) {
	postBody, _ := json.Marshal(data)
	responseBody := bytes.NewBuffer(postBody)

	resp, err := http.Post(target, "application/json", responseBody)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	sb := string(body)

	return sb, nil
}

// HeadRequest performs a HEAD request and return
// a string (the headers of the response).
func HeadRequest(target string) (string, error) {
	resp, err := http.Head(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	sb := string(body)

	return sb, nil
}
