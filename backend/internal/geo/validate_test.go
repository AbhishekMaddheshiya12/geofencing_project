package geo

import "testing"

func TestValidatePolygon(t *testing.T) {
	cases := []struct {
		name    string
		coords  [][]float64
		wantErr bool
	}{
		{
			name: "valid square",
			coords: [][]float64{
				{37.7749, -122.4194},
				{37.7849, -122.4194},
				{37.7849, -122.4094},
				{37.7749, -122.4094},
				{37.7749, -122.4194},
			},
		},
		{
			name: "too few points",
			coords: [][]float64{
				{37.7749, -122.4194},
				{37.7849, -122.4194},
				{37.7749, -122.4194},
			},
			wantErr: true,
		},
		{
			name: "not closed",
			coords: [][]float64{
				{37.7749, -122.4194},
				{37.7849, -122.4194},
				{37.7849, -122.4094},
				{37.7749, -122.4094},
				{37.7749, -122.4195},
			},
			wantErr: true,
		},
		{
			name: "lat out of range",
			coords: [][]float64{
				{91, -122.4194},
				{37.7849, -122.4194},
				{37.7849, -122.4094},
				{91, -122.4194},
			},
			wantErr: true,
		},
		{
			name: "lng out of range",
			coords: [][]float64{
				{37.7749, -181},
				{37.7849, -122.4194},
				{37.7849, -122.4094},
				{37.7749, -181},
			},
			wantErr: true,
		},
		{
			name: "missing element",
			coords: [][]float64{
				{37.7749},
				{37.7849, -122.4194},
				{37.7849, -122.4094},
				{37.7749, -122.4194},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePolygon(tc.coords)
			if (err != nil) != tc.wantErr {
				t.Fatalf("got err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestToWKT_FlipsLatLng(t *testing.T) {
	coords := [][]float64{
		{37.0, -122.0},
		{38.0, -122.0},
		{38.0, -121.0},
		{37.0, -121.0},
		{37.0, -122.0},
	}
	wkt := ToWKT(coords)
	expected := "POLYGON((-122.00000000 37.00000000, -122.00000000 38.00000000, -121.00000000 38.00000000, -121.00000000 37.00000000, -122.00000000 37.00000000))"
	if wkt != expected {
		t.Fatalf("wkt mismatch:\nwant %s\ngot  %s", expected, wkt)
	}
}
