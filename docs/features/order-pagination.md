# Ordering & pagination

## Sorting

```go
h.OrderBy().Field(&m.CreatedAt).DESC()
h.OrderBy().Field(&m.Name).ASC()
```

Multiple calls stack into a single `ORDER BY`, comma-separated:

```go
// ORDER BY priority DESC, created_at ASC
h.OrderBy().Field(&m.Priority).DESC()
h.OrderBy().Field(&m.CreatedAt).ASC()
```

Available in `GetFirst` and `GetList`. `Count`/`Update`/`Delete` don't benefit from sorting, so it's absent in their helpers.

## Pagination (GetList only)

```go
h.Page(1).Size(20)
```

- `Page(n)` — page number, **1-indexed**.
- `Size(n)` — page size (LIMIT).
- OFFSET is computed as `(page - 1) * size`.

!!! warning "Page without Size"
    Calling `Page(...)` without `Size(...)` returns an error on Apply: *"incorrect pagination: size is required then page is set"*.

### Limit only, no offset

Just `Size`:

```go
h.Size(10) // LIMIT 10 OFFSET 0
```

### Page past the end

If OFFSET goes beyond the data, `GetList` returns an empty slice — no error.

## Combining with filters

Call order doesn't matter: WHERE, ORDER BY, and LIMIT/OFFSET are assembled independently.

```go
users, _ := repo.GetList(ctx, func(m *User, h query.GetListHelper[User]) {
    h.Where().Field(&m.Age).GTE(18)
    h.OrderBy().Field(&m.CreatedAt).DESC()
    h.Page(3).Size(25)
})
```
