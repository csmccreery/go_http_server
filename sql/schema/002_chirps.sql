-- +goose Up
CREATE TABLE chirps (
       id UUID UNIQUE NOT NULL,
       created_at TIMESTAMP NOT NULL,
       updated_at TIMESTAMP NOT NULL,
       body TEXT NOT NULL,
       user_id uuid NOT NULL REFERENCES users(id)
);

-- +goose Down
DROP TABLE chirps;
