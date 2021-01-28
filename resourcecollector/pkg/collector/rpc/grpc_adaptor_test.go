/*
Copyright 2020 Authors of Arktos.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rpc

import "testing"

func TestParseRegionAZ(t *testing.T) {
	type args struct {
		regionAZ string
	}
	tests := []struct {
		name       string
		args       args
		wantRegion string
		wantAz     string
		wantErr    bool
	}{
		{"test1", args{regionAZ: "region2|az2"}, "region2", "az2", false},
		{"test2", args{regionAZ: "region2|az2|"}, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRegion, gotAz, err := ParseRegionAZ(tt.args.regionAZ)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRegionAZ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRegion != tt.wantRegion {
				t.Errorf("ParseRegionAZ() gotRegion = %v, want %v", gotRegion, tt.wantRegion)
			}
			if gotAz != tt.wantAz {
				t.Errorf("ParseRegionAZ() gotAz = %v, want %v", gotAz, tt.wantAz)
			}
		})
	}
}
