---
tags: [go]
---

# Error Handling

Wrap errors with `%w` so callers can use `errors.Is` and `errors.As`.

## Sentinel errors

```go
var ErrNotFound = errors.New("not found")
```

## Related

- [[Interfaces]]
- [[memex#Design]] — a link to a heading
