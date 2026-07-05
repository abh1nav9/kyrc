module github.com/abh1nav9/kyrc/server

go 1.24

// The leaderboard API imports the shared protocol/engine/identity packages
// from the CLI module via a local replace, so client and server verify the
// exact same signed bytes. The CLI itself never imports this module or pgx.
require (
	github.com/abh1nav9/kyrc v0.0.0
	github.com/jackc/pgx/v5 v5.6.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/abh1nav9/kyrc => ../
