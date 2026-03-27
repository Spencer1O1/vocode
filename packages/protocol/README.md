# `@vocode/protocol`

JSON Schema is the source of truth for wire types. Codegen produces Go and TypeScript types used by daemon and extension.

## Policy-free boundary (validators + constructor helpers)

Everything in this package that **constructs** or **validates** protocol values must stay **policy-free**:

- **Validators** (`go/validators.go`, `typescript/validators.ts`): express **schema / wire-contract** invariants only (required fields, allowed unions, mutually exclusive fields, enum membership). They must **not** encode daemon safety rules, command allowlists, path normalization strategy, retries, or environment-driven defaults.
- **Constructor helpers** (`go/*_constructors.go`): are thin **shape helpers** for building valid structs (e.g. `kind` + required fields). They must **not** apply business rules, pick defaults from env, or reject payloads for policy reasons.

**Policy** (what is allowed to run, how to interpret ambiguous input, retries, caps) belongs in **`apps/daemon`** (and extension-side UX/execution), not here.

If a helper starts needing policy, move it next to the code that owns that policy (typically `apps/daemon/internal/...`) and keep `packages/protocol` as the shared contract + mechanical validation only.
