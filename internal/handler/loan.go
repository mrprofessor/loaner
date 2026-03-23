package handler

import (
	"context"

	loanv1 "github.com/mrprofessor/loaner/gen/loan/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LoanHandler struct {
	loanv1.UnimplementedLoanServiceServer
}

func New() *LoanHandler {
	return &LoanHandler{}
}

func (s *LoanHandler) CalculateRepayment(ctx context.Context, req *loanv1.CalculateRepaymentRequest) (*loanv1.CalculateRepaymentResponse, error) {
	return &loanv1.CalculateRepaymentResponse{
		MonthlyRepayment: "a millon dollah!!",
		CalculatedAt:     timestamppb.Now(),
	}, nil
}
