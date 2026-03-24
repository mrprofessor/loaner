package handler

import (
	"context"
	"time"

	loanv1 "github.com/mrprofessor/loaner/gen/loan/v1"
	"github.com/mrprofessor/loaner/internal/service"
	"github.com/mrprofessor/loaner/internal/store"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LoanHandler struct {
	loanv1.UnimplementedLoanServiceServer
	store *store.Store
}

func New(s *store.Store) *LoanHandler {
	return &LoanHandler{store: s}
}

func (s *LoanHandler) CalculateRepayment(ctx context.Context, req *loanv1.CalculateRepaymentRequest) (*loanv1.CalculateRepaymentResponse, error) {
	loanAmount, err := decimal.NewFromString(req.LoanAmount)
	if err != nil || loanAmount.LessThanOrEqual(decimal.Zero) {
		return nil, status.Error(codes.InvalidArgument, "loan_amount must be a positive number")
	}

	annualRate, err := decimal.NewFromString(req.AnnualInterestRate)
	if err != nil || annualRate.LessThan(decimal.Zero) {
		return nil, status.Error(codes.InvalidArgument, "annual_interest_rate must be a non-negative number")
	}

	if req.NumPayments <= 0 {
		return nil, status.Error(codes.InvalidArgument, "num_payments must be a positive integer")
	}

	monthlyPayment := service.CalcPMT(loanAmount, annualRate, req.NumPayments)
	now := time.Now().UTC()

	if err := s.store.InsertLoanCalculation(ctx, loanAmount, annualRate, req.NumPayments, monthlyPayment, now); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save calculation: %v", err)
	}

	return &loanv1.CalculateRepaymentResponse{
		MonthlyRepayment: monthlyPayment.StringFixed(2),
		CalculatedAt:     timestamppb.New(now),
	}, nil
}
