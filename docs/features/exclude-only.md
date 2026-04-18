# Exclude & Only

`Exclude` and `Only` narrow the set of columns that participate in a specific operation. Fields are referenced by pointer, same as when configuring columns.

## GetFirst / GetList

Affect **SELECT**: excluded fields are never fetched and stay at their zero value.

```go
// SELECT id, name FROM users WHERE …
u, _ := repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
    h.Where().Field(&m.ID).EQ(id)
    h.Only(&m.ID, &m.Name)
})

// Opposite: fetch everything except the password
u, _ := repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
    h.Where().Field(&m.ID).EQ(id)
    h.Exclude(&m.PasswordHash)
})
```

!!! tip
    Excluded columns keep working in WHERE — this is metadata of the operation, not of the schema.

## Insert

Affects **INSERT**: excluded fields don't appear in the statement — the database uses its `DEFAULT`.

```go
repo.Insert(ctx, u, func(m *User, h query.InsertHelper[User]) {
    h.Exclude(&m.ID, &m.CreatedAt) // ID and CreatedAt are filled by the DB
})
```

## Update

Affects **SET**: only the selected subset is touched.

```go
// UPDATE users SET name = ? WHERE id = ?
repo.Update(ctx, u, func(m *User, h query.UpdateHelper[User]) {
    h.Where().Field(&m.ID).EQ(u.ID)
    h.Only(&m.Name)
})
```

`Exclude` works symmetrically — all fields except the listed ones are updated.

## Interaction with `With*Protection`

Protected columns are **always** excluded from the matching operation, regardless of helpers. For example, `WithUpdateProtection()` on `ID` means `ID` never enters the SET — even if you explicitly pass it into `Only`. That's a schema property; `Only`/`Exclude` is a query property.
