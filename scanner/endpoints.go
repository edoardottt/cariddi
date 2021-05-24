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
*/

package scanner

type EndpointMatched struct {
	Parameters []string
	Url        string
}

//GetJuicyParameters
func GetJuicyParameters() []string {
	var juicyParameters = []string{
		"apikey",
		"api_key",
		"api-key",
		"key",
		"token",
		"secret",
		"secretid",
		"user-id",
		"user_id",
		"userid",
		"secret-id",
		"secret_id",
		"auth-id",
		"auth_id",
		"admin",
		"adminid",
		"admin-id",
		"admin_id",
		"uid",
		"exec",
		"cmd",
	}
	return juicyParameters
}

//RemovDuplicateEndpoints
func RemovDuplicateEndpoints(input []EndpointMatched) []EndpointMatched {
	keys := make(map[string]bool)
	list := []EndpointMatched{}
	for _, entry := range input {
		if _, value := keys[entry.Url]; !value {
			keys[entry.Url] = true
			list = append(list, entry)
		}
	}
	return list
}
