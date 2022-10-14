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

package slice_test

import (
	"net/http"
	"reflect"
	"testing"

	sliceUtils "github.com/edoardottt/cariddi/internal/slice"
)

func TestRemoveDuplicateValues(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		want  []string
	}{
		{
			name:  "empty slice",
			slice: []string{},
			want:  []string{},
		},
		{
			name:  "nil slice",
			slice: nil,
			want:  []string{},
		},
		{
			name:  "withous duplicates",
			slice: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "has duplicates",
			slice: []string{"a", "b", "c", "e", "c", "a"},
			want:  []string{"a", "b", "c", "e"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sliceUtils.RemoveDuplicateValues(tt.slice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveDuplicateValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckInputArray(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty slice",
			input: "",
			want:  []string{},
		},
		{
			name:  "empty strings",
			input: ",,,,",
			want:  []string{},
		},
		{
			name:  "with duplicates",
			input: "a,b,a,,c,,d,b, ,  ,, ",
			want:  []string{"a", "b", "c", "d", " ", "  "},
		},
		{
			name:  "without duplicates",
			input: "a,b,c,d=234, ,",
			want:  []string{"a", "b", "c", "d=234", " "},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sliceUtils.CheckInputArray(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckInputArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckCookies(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []*http.Cookie
	}{
		{
			name:  "empty input",
			input: "",
			want:  []*http.Cookie{},
		},
		{
			name:  "zero pairs",
			input: "asdd311ue2",
			want:  []*http.Cookie{},
		},
		{
			name:  "one pair",
			input: "name:some_value123",
			want: []*http.Cookie{
				{
					Name:  "name",
					Value: "some_value123",
				},
			},
		},
		{
			name:  "several paris",
			input: "name1:some_value@1;name_2:some$%_value@",
			want: []*http.Cookie{
				{
					Name:  "name1",
					Value: "some_value@1",
				},
				{
					Name:  "name_2",
					Value: "some$%_value@",
				},
			},
		},
		{
			name:  "some pairs are not valid",
			input: "name1:value:_1;name;2:value2;name_3:value_3",
			want: []*http.Cookie{
				{
					Name:  "2",
					Value: "value2",
				},
				{
					Name:  "name_3",
					Value: "value_3",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sliceUtils.CheckCookies(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckCookies() = %v, want %v", got, tt.want)
			}
		})
	}
}
