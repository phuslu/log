### go-slog Test Suite

This is a lightweight test suite derived from [https://github.com/madkins23/go-slog](https://github.com/madkins23/go-slog).

Special thanks to [@madkins23](https://github.com/madkins23) for the help, with reference to https://github.com/phuslu/log/pull/70

Verify
```
go test -v -args -useWarnings
```

Benchmark
```
go test -v -run=none -bench=.
```