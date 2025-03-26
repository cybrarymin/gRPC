CREATE TABLE IF NOT EXISTS bank_exchange_rates(
    exchange_rate_uuid UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    from_currency VARCHAR(5) NOT NULL,
    to_currency VARCHAR(5) NOT NULL,
    rate numeric(20,10) NOT NULL,
    valid_from_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_to_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);