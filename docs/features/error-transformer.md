# Error transformer

`WithErrorTransformer(fn func(error) error)` pipes every error returned by the repository through your function. The typical use case is to stop leaking `gerpo.ErrNotFound` outwards and map it to a domain error instead.

## Replacing ErrNotFound

```go
var ErrUserNotFound = errors.New("user not found")

repo, _ := gerpo.New[User]().
    DB(adapter).
    Table("users").
    Columns(/* … */).
    WithErrorTransformer(func(err error) error {
        if errors.Is(err, gerpo.ErrNotFound) {
            return ErrUserNotFound
        }
        return err
    }).
    Build()
```

Now the hexagonal layer knows nothing about gerpo.

## What flows through the transformer

- `GetFirst` — no rows ⇒ `gerpo.ErrNotFound`.
- `Update`, `Delete` — `rowsAffected == 0` ⇒ `gerpo.ErrNotFound`.
- Any DB error (FK, unique, syntax, network).
- `gerpo.ErrApplyQuery`, `gerpo.ErrApplyPersistentQuery` — when WHERE/ORDER/etc. could not be assembled.

## What does **not**

- The happy path — the transformer isn't invoked when `err == nil`.
- Logic errors raised before any DB call (e.g. an empty `Build()` state) — they come out of `gerpo.New[T]().…Build()`, which is outside the transformer.

## Passing back the wrapped error

```go
.WithErrorTransformer(func(err error) error {
    switch {
    case errors.Is(err, gerpo.ErrNotFound):
        return ErrUserNotFound
    default:
        // wrap with domain context, keep the original for logs
        return fmt.Errorf("user repo: %w", err)
    }
})
```

## Behavior with tests

`errors.Is` is transitive: `errors.Is(err, ErrUserNotFound)` → true. But `errors.Is(err, gerpo.ErrNotFound)` → **false**, because the wrapping chain starts from the replacement. That's by design: the transformer is the boundary between infrastructure and domain.
