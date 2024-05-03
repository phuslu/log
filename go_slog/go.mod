module go_slog

go 1.21.5

toolchain go1.22.2

require (
	github.com/madkins23/go-slog v0.9.8-beta-8
	github.com/phuslu/log v1.0.93
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/gertd/go-pluralize v0.2.1 // indirect
	github.com/gomarkdown/markdown v0.0.0-20240419095408-642f0ee99ae2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/phuslu/log => ../
