---
tags: [go, reference]
source: "Effective Go"
---

# Interfaces

Interfaces in Go are satisfied implicitly.

## Method sets

### Pointer receivers

A method with a pointer receiver belongs only to the pointer type.

#### Example

```go
# This line starts with a hash but is code, not a heading.
type Strategy interface {
	Name() string
}

// *T implements Strategy; T does not.
func (t *T) Name() string { return "t" }
```

### Value receivers

A method with a value receiver belongs to both T and *T.

## Accept interfaces, return structs

Define interfaces where they are used, not where they are implemented.
