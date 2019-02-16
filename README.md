# Report uses of variables that are nil

When code has already dealt with a variable being nil, it doesn't usually make sense to continue
referencing that variable. e.g.

```go
func foo() error {
  result, err := DoSomething()
  if err != nil {
    return err
  }
  if result.IsBad() {
    return err
  }
  return nil
}
```

## Usage

```bash
go get github.com/devnev/go-lint-ref-after-nil-check
go-lint-ref-after-nil-check source_file.go source_dir
```

## Status

This has not been tested on a real codebase yet.
