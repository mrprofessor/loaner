package service

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalcPMT(t *testing.T) {
	tests := []struct {
		name        string
		principal   string
		annualRate  string
		numPayments int32
		want        string
	}{
		{
			name:        "30-year mortgage",
			principal:   "100000",
			annualRate:  "5.5",
			numPayments: 360,
			want:        "567.79",
		},
		{
			name:        "one year loan",
			principal:   "10000",
			annualRate:  "10",
			numPayments: 12,
			want:        "879.16",
		},
		{
			name:        "zero interest",
			principal:   "12000",
			annualRate:  "0",
			numPayments: 12,
			want:        "1000.00",
		},
		{
			name:        "single payment",
			principal:   "5000",
			annualRate:  "6",
			numPayments: 1,
			want:        "5025.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := decimal.RequireFromString(tt.principal)
			r := decimal.RequireFromString(tt.annualRate)
			got := CalcPMT(p, r, tt.numPayments)
			want := decimal.RequireFromString(tt.want)
			if !got.Equal(want) {
				t.Errorf("CalcPMT(%s, %s, %d) = %s, want %s", tt.principal, tt.annualRate, tt.numPayments, got, tt.want)
			}
		})
	}
}
