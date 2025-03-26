CREATE TABLE IF NOT EXISTS bank_transfers(
    transfer_uuid UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    from_account_uuid UUID REFERENCES bank_accounts (account_uuid) ON DELETE SET NULL,
    to_account_uuid UUID REFERENCES bank_accounts (account_uuid) ON DELETE SET NULL,
    currency VARCHAR(20) NOT NULL,
    amount NUMERIC(15,2) NOT NULL,
    transfer_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    transfer_succeed BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);