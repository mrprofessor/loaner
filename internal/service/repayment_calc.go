package service

import "github.com/shopspring/decimal"

func CalcPMT(principal, annualRate decimal.Decimal, numPayments int32) decimal.Decimal {
	// Convert percentage to decimal (e.g. 5.5 -> 0.055)
	annualRate = annualRate.Div(decimal.NewFromInt(100))

	// if rate is zero return principal/numPayments
	if annualRate.IsZero() {
		return principal.Div(decimal.NewFromInt32(numPayments)).Round(2)
	}

	// Monthly rate: annualRate/12
	r := annualRate.Div(decimal.NewFromInt32(12))

	// (1 + r)^n
	pow := r.Add(decimal.NewFromInt32(1)).Pow(decimal.NewFromInt32(numPayments))

	// Numerator: principal * r * (1 + r)^n
	numerator := principal.Mul(r).Mul(pow)

	// Denominator: (1+r)^n - 1
	denominator := pow.Sub(decimal.NewFromInt32(1))

	// return:  P × r(1+r)^n / ((1+r)^n - 1)
	return numerator.Div(denominator).Round(2)
}
