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

	@Author:      edoardottt, https://edoardottt.com

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package slice_test

import (
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
			name:  "one element",
			slice: []string{"a"},
			want:  []string{"a"},
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
			want:  []string{"a", "b", "c", "d"},
		},
		{
			name:  "without duplicates",
			input: "a,b,c,d=234, ,",
			want:  []string{"a", "b", "c", "d=234"},
		},
		{
			name:  "spaces around elements",
			input: "forum, blog , forum,\tshop",
			want:  []string{"forum", "blog", "shop"},
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
