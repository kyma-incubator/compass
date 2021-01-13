package token

import "time"

type Entity struct {
	// TODO: May be use Token for Primary key as it has built in indexing
	ID        string    `db:"id"`
	Token     string    `db:"token"`
	TokenType string    `db:"token_type"`
	ClientID  string    `db:"client_id"`
	CreatedAt time.Time `db:"created_at"`
	UsedAt    time.Time `db:"used_at"`
	Used      bool      `db:"used"`
}
