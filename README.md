# kimoj-core

The shared kernel of [kimoj](https://github.com/solomanpan/kimoj) (the Go twin of the Python `hiaok-core`).

* `exception` — `DomainError` (a business failure, a value) vs `InfrastructureError` (a technical failure, a panic); kinds map to 400 / 401 / 403 / 404 / 409.
* `factory` — the error catalog: define every business error once (`factory.Conflict("system_role", "system role '{code}' cannot be deleted")`).
* `fp` — `Result[A] = Either[DomainError, A]` and `Option`, on top of [IBM/fp-go](https://github.com/IBM/fp-go).
* `facade` — contracts only: transaction scopes (`InMaster` / `InSlave`), event bus, cache, password hasher, id generator.
* `rctx` — request context: request id, operator, locale.

```bash
go get github.com/solomanpan/kimoj-core@v0.1.0
go test ./...
```

MIT © 2026 PAN JIE
