CREATE TABLE IF NOT EXISTS bank_transactions(
    transaction_uuid UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    account_uuid UUID REFERENCES bank_accounts (account_uuid) ON DELETE SET NULL,
    transaction_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    amount NUMERIC(15,2) NOT NULL,
    transaction_type VARCHAR(25) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);