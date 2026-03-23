CREATE TABLE IF NOT EXISTS loan_repayment (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    loan_amount       NUMERIC(20, 6) NOT NULL,
    annual_rate       NUMERIC(10, 6) NOT NULL,
    num_payments      INTEGER        NOT NULL,
    monthly_repayment NUMERIC(20, 6) NOT NULL,
    calculated_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
