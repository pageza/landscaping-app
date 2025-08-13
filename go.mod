module github.com/pageza/landscaping-app

go 1.18

require (
	github.com/gorilla/mux v1.8.0
	github.com/lib/pq v1.10.9
	github.com/google/uuid v1.3.0
	github.com/golang-migrate/migrate/v4 v4.16.2
	github.com/pageza/go-storage v0.1.0
	github.com/pageza/go-llm v0.1.0
	github.com/pageza/go-payments v0.1.0
	github.com/pageza/go-comms v0.1.0
)

// Local package replacements until they are published
replace github.com/pageza/go-storage => ../go-storage
replace github.com/pageza/go-llm => ../go-llm
replace github.com/pageza/go-payments => ../go-payments
replace github.com/pageza/go-comms => ../go-comms
