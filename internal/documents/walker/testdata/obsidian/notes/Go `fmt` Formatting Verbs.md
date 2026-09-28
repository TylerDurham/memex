---
aliases: [Go `fmt` Formatting Verbs]
linter-yaml-title-alias: Go `fmt` Formatting Verbs
date created: Saturday, September 12th 2026, 12:05:52 am
date modified: Saturday, September 12th 2026, 12:06:17 am
---

# Go `fmt` Formatting Verbs

|Verb|Description|
|---|---|
|`%v`|Default format for the value|
|`%+v`|Adds field names when formatting structs|
|`%#v`|Go-syntax representation of the value|
|`%T`|Type of the value|
|`%%`|A literal percent sign (escapes `%`)|
|`%t`|Boolean — `true` or `false`|
|`%d`|Base-10 integer|
|`%b`|Base-2 (binary) integer|
|`%o`|Base-8 (octal) integer|
|`%O`|Base-8 with `0o` prefix|
|`%x`|Base-16 (hex), lowercase|
|`%X`|Base-16 (hex), uppercase|
|`%c`|Character represented by the corresponding rune|
|`%q`|Single-quoted rune, safely escaped|
|`%U`|Unicode format, e.g. `U+1234`|
|`%f`|Decimal point, no exponent (e.g. `123.456`)|
|`%F`|Synonym for `%f`|
|`%e`|Scientific notation, lowercase `e` (e.g. `1.234456e+02`)|
|`%E`|Scientific notation, uppercase `E`|
|`%g`|`%e` for large exponents, `%f` otherwise — most compact, useful representation|
|`%G`|Same as `%g`, uses `%E` for large exponents|
|`%s`|Plain string (or `[]byte`) value|
|`%q`|Double-quoted string, safely escaped|
|`%x`|Hex dump of string/bytes, lowercase, two chars per byte|
|`%X`|Hex dump of string/bytes, uppercase|
|`%p`|Pointer address, base-16, with leading `0x`|
|`%w`|Wraps an error (only valid with `fmt.Errorf`) — supports `errors.Unwrap`/`errors.Is`/`errors.As`|

**Width/precision & flags** (used alongside the verbs above):

|Syntax|Description|
|---|---|
|`%6d`|Pad to minimum width of 6|
|`%-6d`|Pad to width 6, left-justified|
|`%06d`|Pad with zeros to width 6|
|`%.2f`|2 digits of precision after the decimal|
|`%8.2f`|Width 8, precision 2|
|`%+d`|Always show sign (`+` or `-`)|
|`% d`|Space instead of `+` for positive numbers|
|`%#x`|Add `0x` prefix (or `0` for octal)|
|`%*d`|Width supplied as an argument, not a literal|

A couple of usage notes worth knowing:

- `%v` combined with a type's `String() string` method (the `Stringer` interface) will use that method automatically for custom formatting.
- `%w` is special-cased — it only works inside `fmt.Errorf` and is what enables Go's error-wrapping chain (`errors.Is`, `errors.As`, `errors.Unwrap`).
