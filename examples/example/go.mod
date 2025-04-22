module example

go 1.23.0

toolchain go1.23.1

require (
	github.com/insei/gerpo v1.0.0
	github.com/lib/pq v1.10.9
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/insei/fmap/v3 v3.1.2 // indirect
)

replace github.com/insei/gerpo v1.0.0 => ../../
