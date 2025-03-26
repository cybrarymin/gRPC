CREATE TABLE IF NOT EXISTS bank_accounts(
    account_uuid UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    account_number VARCHAR(20) UNIQUE NOT NULL,
    account_name VARCHAR(100) NOT NULL,
    currency VARCHAR(5) NOT NULL,
    current_balance NUMERIC(15,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);