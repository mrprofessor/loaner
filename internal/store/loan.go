package store

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

func (s *Store) InsertLoanCalculation(ctx context.Context, loanAmount, annualRate decimal.Decimal, numPayments int32, monthlyRepayment decimal.Decimal, calculatedAt time.Time) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO loan_repayment (loan_amount, annual_rate, num_payments, monthly_repayment, calculated_at)
         VALUES ($1, $2, $3, $4, $5)`,
		loanAmount, annualRate, numPayments, monthlyRepayment, calculatedAt,
	)
	return err
}
