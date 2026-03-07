
CREATE TABLE IF NOT EXISTS Users (

    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    hash_password TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP

);

CREATE TABLE IF NOT EXISTS Accounts (

    user_id UUID NOT NULL,
    account_id INTEGER NOT NULL,

    balance BIGINT NOT NULL DEFAULT 0,
    is_imported BOOLEAN NOT NULL,
    name_account VARCHAR(50) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_user_account
        FOREIGN KEY (user_id)
        REFERENCES Users(user_id)
        ON DELETE CASCADE,

    CONSTRAINT pk_account 
        PRIMARY KEY (user_id, account_id)
);

CREATE TABLE IF NOT EXISTS Category (

    category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), 
    user_id UUID NOT NULL,

    name_catagory VARCHAR(255) NOT NULL,
    is_income BOOLEAN NOT NULL,
    is_castome BOOLEAN NOT NULL DEFAULT FALSE,
    icon_url TEXT,

    CONSTRAINT fk_user_category
        FOREIGN KEY (user_id)
        REFERENCES Users(user_id)
        ON DELETE CASCADE,

    UNIQUE(name_catagory, is_income, user_id)
);

CREATE TABLE IF NOT EXISTS Transactions (

    transaction_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    user_id UUID NOT NULL,
    category_id UUID,

    name_transaction TEXT NOT NULL,
    is_income BOOLEAN NOT NULL,
    amount BIGINT CHECK (amount > 0) NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL,
    is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    is_imported BOOLEAN NOT NULL DEFAULT FALSE,
    comment TEXT,

    CONSTRAINT fk_user_transaction
        FOREIGN KEY (user_id)    
        REFERENCES Users(user_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_category_transaction
        FOREIGN KEY (category_id)
        REFERENCES Category(category_id)
        ON DELETE SET NULL,

    CONSTRAINT fk_account_transaction
        FOREIGN KEY (account_id)
        REFERENCES Accounts(account_id, account_id),

    CONSTRAINT pk_trasactions 
        PRIMARY KEY (user_id, transaction_id)
);

