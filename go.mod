module github.com/mr-flannery/go-recipe-book

go 1.23.0

toolchain go1.24.4

require github.com/lib/pq v1.10.9 // PostgreSQL driver

require (
	github.com/golang-migrate/migrate/v4 v4.18.3
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
)
