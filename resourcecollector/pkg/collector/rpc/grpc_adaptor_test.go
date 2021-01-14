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
