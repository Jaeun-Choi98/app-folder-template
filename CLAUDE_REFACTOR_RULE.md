# CLAUDE.md

This file guides Claude Code when refactoring a legacy application into a new, layered
implementation in this workspace (uwemgradt, uwemgr, sig, osvms, and any future project done the
same way).

## General behavioral guidelines (foundational — apply on top of everything else in this file)

These are general behavioral guidelines to reduce common LLM coding mistakes, meant to be merged
with the project-specific instructions that follow, not replace them. Where a rule below ever
conflicts with this file's own top-priority compatibility rule (next section) — e.g. "Simplicity
First" suggesting a cleaner approach where legacy behavior must be preserved byte-for-byte, or
"Surgical Changes" scoping a change more narrowly than a full legacy-parity port requires —
compatibility wins, per this file's own stated hierarchy ("above every other convention in this
file or introduced later"). Absent such a conflict, these apply generally.

**Tradeoff**: these guidelines bias toward caution over speed. For trivial tasks, use judgment.

### 1. Think Before Coding

Don't assume. Don't hide confusion. Surface tradeoffs.

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them — don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### 2. Simplicity First

Minimum code that solves the problem. Nothing speculative.

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### 3. Surgical Changes

Touch only what you must. Clean up only your own mess.

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it — don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: every changed line should trace directly to the user's request.

### 4. Goal-Driven Execution

Define success criteria. Loop until verified.

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant
clarification.

These guidelines are working if: fewer unnecessary changes in diffs, fewer rewrites due to
overcomplication, and clarifying questions come before implementation rather than after mistakes.

## Top priority: complete compatibility — the new project must be a drop-in replacement

The single highest-priority rule, above every other convention in this file or introduced later:
**the rewritten project must be usable as a complete substitute for the legacy project it replaces.**
Nothing else in a refactor is allowed to compromise this. If a structural/architectural rule
elsewhere ever conflicts with compatibility, compatibility wins.

Concretely, this means every external interface is ported **bit-for-bit / byte-for-byte identical**
to legacy, not "improved," "fixed," or "modernized" as a side effect of the rewrite:

- **HTTP**: URL path, HTTP method, request shape, response shape, status codes, and header behavior.
  "Shape" means the literal wire-visible JSON key names (the `json:"..."` tag value), not just the
  Go field identifier — Go-side field naming can follow this project's own convention, but the tag
  string a field serializes under must match legacy's key byte-for-byte, along with each field's
  type, presence/absence, and nesting. A frontend built against legacy parses responses by key name;
  a same-data-different-key response is indistinguishable from a missing field to that caller. This
  has already caused real bugs twice: one project shipped clean camelCase tags across many routes
  instead of legacy's actual wire keys, and a sibling project had the identical class of mismatch
  found and fixed separately. Treat both as standing cautionary examples alongside the bcrypt one
  below — when porting a response type, copy legacy's literal tag strings, don't rederive them from
  what "looks like" a reasonable name for the field.
- **SSE**: the same JSON-key rule above applies to the payload carried in an SSE event's `data`
  field, not just ordinary HTTP request/response bodies — an SSE push is still an external interface
  a frontend parses by key name, and it's easy to forget it's in scope because its DTOs typically
  live in their own `internal/event/`-style package, physically separate from
  `internal/transport/http-rest/response/`, so a sweep that only checks the latter silently skips
  it. Verify SSE payload keys against legacy explicitly, as their own pass — don't assume "we already
  checked all the response types" covered them.
- **TCP**: opcode/command values, byte offsets, field widths, endianness, frame/delimiter structure,
  parsing and serialization logic — anything that determines what goes on the wire in either
  direction.
- **UDP**: same as TCP — packet layout, field offsets/widths, endianness.
- Any other external interface a legacy process exposes (file formats it reads/writes, IPC/shared
  memory contracts if anything outside the process depends on their shape, CLI arguments, config
  file keys) follows the same rule.
- **Existing code flow and business rules** — the order of operations, validation logic, what gets
  persisted vs. skipped, error/edge-case handling, auth comparison logic and enforcement (or lack of
  it) — are preserved exactly, even when the legacy behavior looks wrong, insecure, or awkward from
  the outside.
- **Dispatch/trigger table completeness** — every opcode a legacy TCP/UDP handler switches on, every
  route legacy registers, and every condition under which legacy fires an SSE/event push must have a
  1:1-covering entry in the new project's own dispatch table, route table, or trigger set. This is a
  different failure mode from getting an individual entry's wire format right: a dispatch table with
  a missing case, a route-registration function that's defined but never called, or an SSE publish
  that legacy fires on some state change but the port doesn't, all compile clean and pass review —
  every entry that *is* there can be byte-perfect — and the gap only surfaces once that specific
  opcode/route/trigger condition actually occurs in production. Porting "the handlers" is not the
  same task as porting "the table that dispatches to them"; verify the latter explicitly, by diffing
  the full set of legacy cases/routes/triggers against the new project's, not by spot-checking a few.
  This already happened here: a `Route*Paths()` function that correctly implemented ~13 routes was
  simply never called from the router's setup, silently unregistering all of them — found and fixed
  by an explicit full-route diff against legacy, not by reading the individual route handlers, which
  all looked correct in isolation.

Why this is the top priority, not just one convention among many: another system, service, client,
or piece of hardware talking to this process was coded against the *exact* current shape of that
interface. It does not know or care that the internals were refactored. A URL path rename, a
reordered TCP field, or a "fixed" password comparison is invisible to a code reviewer reading only
the new codebase, but breaks every real caller the moment it ships — with no compile error and no
warning. This already happened once here: swapping a legacy plaintext password comparison for
`bcrypt.CompareHashAndPassword` against a column that still held plaintext values failed every login
outright, for every account, the instant the change shipped, because "fix the auth" was done
unprompted as part of a structural pass. Treat that as the standing cautionary example.

## What "refactoring" means here: transforming structure, not rewriting source code

This is the most important framing in this file, and everything else follows from it: the goal of
this project is to move existing, working logic into a new layered architecture — new file/package
boundaries, new struct types, a new process model — **not to rewrite what that logic does.** The
default action on any piece of legacy logic is to relocate it with its behavior and internal decision
-making intact, not to rewrite it into something different because the new version looks cleaner,
more idiomatic, or more efficient.

This applies *inside* a function's body, not just at the external boundaries described above. Do not
reorder steps because the new order seems more sensible, drop a branch that looks dead or redundant,
tighten a loose conditional, change a default, or otherwise alter what the code computes or decides
— even when nothing external would ever notice the difference. The "existing code flow and business
rules" bullet above is one instance of this same rule, called out at the boundary case (external
behavior) where it's most obviously required; this generalizes it to all of the logic, including the
parts no caller could ever observe changing.

What *is* fine, because it's structural rather than logical: splitting one legacy function's body
across the new architecture's layers, renaming identifiers to fit Go convention or the new type they
live on, adapting a call's signature to whatever the new layering requires, or swapping a legacy
mechanism for an equivalent one this project has already standardized on elsewhere (a module
primitive instead of hand-rolled plumbing, for instance) — as long as the net decision-making and
computed results are unchanged. When genuinely unsure whether a change is "just structure" or
actually changes behavior, treat it as the latter and ask before making it.

## Security-looking issues are not an exception — never fix them silently

If legacy code has an obvious security problem — plaintext password comparison, a "token" that's
just the password echoed back, no server-side session, no endpoint-level auth check, a predictable
ID used as a secret, anything that looks like a real vulnerability — **port it exactly as legacy has
it anyway.** Looking insecure is not grounds for an exception to the compatibility rule above. Do not
hash a previously-plaintext comparison, do not add a session/token check that didn't exist, do not
gate an endpoint that was open, do not change a response to withhold information legacy exposed —
not even "just this one obvious case" — while doing a compatibility-focused refactor.

This is intentional, not an oversight: the legacy system's callers (a frontend, another service) are
coded against — and in some cases actively depend on — the exact current behavior, including its
insecurity. Silently closing a hole changes what a legitimate call returns or whether it succeeds,
which is exactly the kind of interface change the top-priority rule forbids. It also risks a hard
failure rather than a quiet improvement: see the bcrypt-vs-plaintext-column incident above — the
"fix" didn't make login more secure, it made login fail for everyone, immediately, in production.

Fixing a real security issue is legitimate work, but it is always a **separate, explicitly-requested,
cross-team-coordinated phase** — never something to do opportunistically while already in the file
for structural reasons. When asked to do that phase, flag the coordination it needs (a data
migration for hashing, a frontend release that now expects a token, etc.) rather than shipping a
silent behavior change.

**Narrow scope note:** this rule protects *observable behavior for legitimate callers*, not literal
source text. An internal implementation change that a legitimate caller cannot tell happened —
parameterizing a SQL query that was previously built with string interpolation, for example — is not
covered by this rule, because it does not change what any real request/response pair looks like; it
only removes an injection path for adversarial input. Ordinary hygiene like that is fine. If there's
any doubt whether a change could alter observable behavior for a real caller, treat it as covered by
this rule and ask first.

Rules that follow from the top-priority rule generally:

- Never change an interface's observable behavior — including a call that legacy handles insecurely,
  inefficiently, or in a way that looks like a bug — as a side effect of "porting" or "refactoring"
  it. If something looks like a real bug or a real security hole, say so explicitly and ask; do not
  silently fix it while moving the code.
- Hardening or cleanup that would change an external contract (hashing at rest, real auth tokens,
  endpoint enforcement, renaming a wire field, fixing a framing bug, changing a URL) is a separate,
  explicitly-requested phase — never bundled into a compatibility-preserving refactor.
- Route/endpoint parity is verified by diffing the actual registration call sites in the legacy
  source against the new project's registered routes — not by judging from the outside which
  handlers "look like" real features versus debug/test scaffolding. A live-but-trivial-looking
  handler (a hardcoded test response, a manual protocol-poke endpoint) is still part of the contract
  if legacy actually registers it on a reachable route. A copy-pasted-but-never-wired-up type is not,
  no matter how official it looks. The only reliable signal is registration status in the legacy
  source, checked directly — not memory of "what seemed to matter" from reading the file once.
- A structural rewrite never adds a new configuration key, section, or file that the legacy
  deployment didn't already have. The env/config surface is as much an external contract as the API
  — existing deployment tooling and runbooks are written against the legacy key set.
- When in doubt about whether a change affects an external interface, assume it does and preserve
  the legacy behavior; ask before deviating.

## Structural standard: `app-folder-template`, with `rpi` as a secondary reference

The folder layout, layering, and dependency-injection shape for every refactor target in this
workspace follows `./app-folder-template/gin` (relative to this workspace's root). It's a
from-scratch template, not a legacy port, so it's free of any
compatibility constraint — treat its `internal/` package layout
(`app`/`config`/`container`/`db`/`event`/`healthcheck`/`infra`/`logger`/`service`/`transport`/`worker`,
with `transport/http-rest` and `transport/tcp/{client,server}` each split further into
`controller`/`parser`/`serializer`/`middleware`/`request`/`response`) as the default shape for a new
project's package tree, and its `container.go`/`app.go` wiring as the default shape for the DI root.

`./app-folder-template/rpi` is a second, real (non-template) project built on the same
foundation — useful when `app-folder-template` doesn't show a pattern for something a specific
refactor needs (e.g. its SPA-serving middleware, or its serial-port transport package). When
borrowing from `rpi`, borrow structure, not business rules — `rpi`'s own domain-specific behavior
(its extension whitelist, its redirect rules, etc.) is `rpi`'s, not a convention to copy into an
unrelated project; only the shape of how it split responsibilities into files/packages travels.

## Use the `github.com/Jaeun-Choi98/modules` packages instead of hand-rolling their equivalents

Both `app-folder-template` and `rpi` build on a shared internal module,
`github.com/Jaeun-Choi98/modules`, that already solves several cross-cutting problems this kind of
refactor runs into repeatedly. Depend on the relevant submodule and use its primitives rather than
writing a bespoke version of the same thing:

- **`modules/sse`** — `sse.SessionManager[K, V]` / `sse.NewSSEClient[K, V](...)` give
  session-per-user SSE with a client registry, broadcast-to-session, and broadcast-to-all already
  built. Use it for any SSE endpoint instead of hand-rolling a client map + broadcast loop — see
  `./app-folder-template/gin/internal/transport/http-rest/controller/sse_handle.go` for the
  reference wiring (`HandleSSEConnect`, `SendSSEMessageAll`, `SendSSEMessageToUser`).
- **`modules/eventbus`** — `eventbus.Publish`/`eventbus.Subscribe[T]`/`eventbus.Request[T]` on a
  shared bus (`event.GetEventBus()`). Two uses, and they take different primitives: `Publish` for
  fan-out where nobody waits — "something changed" reaching connected SSE clients is the standard
  case, and a service must never be wired into the SSE layer with a callback or channel of its own —
  and `Request` for outbound sends whose result the caller reads. The section below
  ("The service layer never holds a communication object") covers the second use in full; it is the
  pattern every project in this workspace uses to keep `service → transport` from existing.
- **`modules/tcpnet`** (`basic/client.ClientBase`, `basic/server.ServerBase`,
  `advanced/parser.Parser`, `advanced/serializer.BinaryWriter`, `advanced/handler.Manager[T]`,
  `advanced/model`) — connection lifecycle, framing, dispatch-by-opcode, and request/reply
  correlation for any custom TCP protocol. Don't hand-roll reconnect/backoff logic, a manual `switch
  opcode` dispatch, or a hand-rolled ack-channel — these are exactly what this module is for.
- **`modules/utils`** — shared small helpers (e.g. `utils.LocalKorea` for timezone-correct
  `time.Now()` calls) — check here before writing a one-off equivalent.
- Prefer these over ad hoc equivalents even when the ad hoc version would be shorter for one call
  site — the point of using a shared module is consistency across every project in this workspace,
  not minimizing line count in any single one.

## The service layer never holds a communication object — publish to the bus instead

A service that sends anything outward (a TCP frame to a controller, a UDP packet to a peer, an HTTP
call to another server) must not hold the object that does it. No `*conn.Conn`, no `*client.Manager`,
no `*apiclient.Client` as a struct field, and no interface standing in for one either when the bus
will do. The service publishes a typed event on `event.GetEventBus()`; the transport package
subscribes and performs the send.

Two things go wrong without this, and both have shown up in this workspace:

- **The dependency points the wrong way.** `service → transport` contradicts the layering every
  project's `service.go` states in its own header comment. The inbound direction is usually already
  clean (transport → a service interface); it's the outbound direction that quietly leaks.
- **It creates a cycle, which forces two-phase wiring.** `conn → controller → service → conn` cannot
  be built in one pass, so the container ends up constructing the conn as an empty shell and
  deferring the real wiring to `app.Start()`. Removing the return arrow removes the cycle, and the
  container goes back to being an ordinary inside-out constructor.

### Use `eventbus.Request`, not `Publish`

This is the part that decides whether the refactor is behavior-preserving. These call sites are
synchronous today: the service reads the device's result code and branches on it, or needs the
response body. `Publish` is fire-and-forget and would silently change that control flow; `Request`
runs the subscriber and hands back its return value and its error, so the caller's flow is untouched.
Reserve `Publish` for genuine fire-and-forget fan-out — SSE broadcast is the standard case.

Four details that are easy to get wrong, each of which has bitten a real port:

1. **Zero subscribers is silent success.** `eventbus.Request` with nobody subscribed returns an empty
   reply slice *and* an empty error slice. Left alone, a frame that went nowhere reads as a
   successful send. Every `event` package must convert that to an explicit error — this is the only
   new failure mode the pattern introduces, so name it plainly (`"no subscriber"`).
2. **Sentinel errors must survive the crossing.** Where a service compares `err == ErrNotConnected`
   to tell "couldn't send" from "sent but no reply" — and that distinction is often load-bearing
   legacy behavior, not cosmetics — define the sentinel in the `event` package and have the transport
   alias it (`var ErrNotConnected = event.ErrNotConnected`). Return subscriber errors unwrapped so
   `==` still holds.
3. **Cross the boundary with neutral types.** Put the payload and reply types in `event`, not the
   transport's own frame/DTO types. Where the service only reads two fields off a nine-field
   `ReqMsg`, the neutral type is those two fields and the service stops knowing the frame layout;
   where it genuinely reads all of them, mirroring the shape is fine — the point is breaking the
   import, not shrinking the struct.
4. **Correlation belongs to transport.** Session-key pools, reply channels, and the wait-for-ack
   timeout are frame bookkeeping. If they live in the service, move them down with the send; the
   service should be left calling one function and reading a result.

### Wiring and ordering

Subscriptions are established in the transport component's `Start()` (or a `SubscribeCommands()` it
returns an unsubscribe func from), and **the subscriber must start before any publisher**. A worker
that publishes on a ticker, or a link whose first inbound packet triggers an outbound call, will
otherwise hit `no subscriber` on startup. On shutdown the order reverses: stop the publishers, then
unsubscribe, then close the bus — so nothing is left publishing into a closed bus.

Keep the general pattern here; a project's own `CLAUDE.md` is where its specific topic names and
event types get written down once they exist.

## Every function gets a comment describing what it does

While relocating a piece of legacy logic into the new structure, add a short comment directly above
the function stating its role — exported or not, one-liner or not. This applies uniformly, not just
to the functions that seem to need explaining. The bar is: someone scanning only the comments in a
file, top to bottom, without reading any bodies, should get an accurate description of what that file
does.

For a genuinely new function that has no legacy counterpart (scaffolding, DI wiring, a new module
primitive's adapter), a plain comment stating what it does is enough. For a function that ports
legacy behavior — which is most of them in this project — say what legacy function or file it
replaces, since that link is otherwise only implicit in the fact that the logic happens to match:
e.g. `sendParams replaces the 17-byte fixed layout that clientif.go's makeParamMsg used to build.`
State the role/contract, not a line-by-line narration of the body — describe what the function does
and, where relevant, why it's shaped the way it is, not a restatement of each line inside it.

This comment is also where the "structural transformation, not source rewrite" principle above
becomes checkable after the fact: if a function's comment says what legacy behavior it replaces, a
later reviewer (human or Claude) can go re-read that legacy function and confirm the port didn't
quietly change anything — without that pointer, there's no way to verify parity except re-deriving
it from memory.

## Don't keep a private function that doesn't earn its name — inline it

A private helper is worth existing when its *name* tells the reader something the call site couldn't.
When it doesn't, it costs a jump: the reader leaves the flow they were following, reads a few lines
elsewhere, and comes back. Fold that body into its caller instead.

**Inline it** when all of these hold:

- it has exactly one call site,
- it's short — a handful of statements, no independent control flow worth naming,
- and its name only restates what the code plainly says (`buildX` that builds an `X`, `setY` that
  assigns `Y`, a wrapper that forwards its arguments to one other call and does nothing else).

**Keep it as a function** when any of these hold — these are the reasons a name earns its keep, and
they outrank the bullets above:

- it's called from more than one place (inlining would duplicate logic, which is worse),
- the name states a rule the body doesn't make obvious on its own (`isReplayable`, `genLRC`,
  `elapsed`) — naming a non-obvious *why* is exactly what a helper is for,
- it's the documented port boundary for a specific legacy function, so its comment is what makes the
  parity claim checkable later (the section above),
- it's `defer`red, passed as a value, used as a closure/adapter, or otherwise needs to be a function,
- or inlining it would make the caller long enough to stop reading as one idea. Trading three tidy
  20-line functions for one 60-line one is a loss, not a win — this rule removes indirection that
  buys nothing, it doesn't reward long functions.

This is a readability rule, so it is subject to everything above it: **inlining must be provably
behavior-preserving.** Watch the traps that make it not — a `return` inside the helper only exits the
helper, `defer` inside it fires at the helper's end rather than the caller's, and named
results/shadowed variables can change what a bare `return` yields. If folding the body in would
require rearranging control flow to stay equivalent, leave the function alone; the indirection is
cheaper than the risk.

Apply it while writing new code and when passing through existing code — not as a standalone sweep
across files nobody asked about.

## Long-running goroutines: the component owns its own, and never dies of a panic

Two rules that go together, both about goroutines that live for the whole process — monitors, GC
workers, device rx loops, anything shaped like `for { select { case <-ctx.Done(): ...; case <-tick: ... } }`.

**The component that runs the loop starts it.** A `Start()` method spawns its own goroutine and
returns; the caller writes `c.Monitor.Start()`, never `go c.Monitor.Start()`. The owner is then the
only place that knows the loop exists, so it can pair `wg.Add(1)` with the spawn, expose a
`Shutdown()` that cancels and waits, and let a reader see the whole lifecycle in one file. Leaving
the `go` to the caller splits that lifecycle across two packages and invites a specific bug this
workspace has already shipped: with `wg.Add(1)` *inside* the goroutine, a `Shutdown()` that lands
before the scheduler runs it calls `wg.Wait()` on a zero counter and returns while the loop is still
running. The same reasoning rules out a second owner for work something else already supervises — a
`StartDBHeartbeat()` on a DB handler is dead weight once `healthcheck.SystemMonitoring` pings that
same handler on a ticker, and two tickers reconnecting one pool is worse than one.

**A panic in the loop must not take the process down.** These goroutines run unattended for months
and touch parsed device input, map lookups, and type assertions — exactly where a nil deref or an
out-of-range index hides. There is no caller up the stack to recover: an unrecovered panic in any
goroutine kills the whole process, so one malformed packet takes down every other link with it.
Wrap the loop body:

```go
func (w *Worker) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			if w.runLoop() { // false = ctx가 끝나 정상 종료
				return
			}
		}
	}()
}

// runLoop는 패닉이 나면 로그를 남기고 true를 돌려 바깥 루프가 다시 돌게 한다.
func (w *Worker) runLoop() (restart bool) {
	defer func() {
		if r := recover(); r != nil {
			logger.Infof("[Worker] panic recovered, restarting: %v\n%s", r, debug.Stack())
			restart = true
		}
	}()
	...
}
```

Restart rather than return: a supervisor that silently stops supervising is the failure this rule
exists to prevent, and the loop's state is rebuilt from the ticker on the next pass anyway. Always
log the value *and* `debug.Stack()` — a recovered panic with no stack is a bug you cannot find later.
Do not spread `recover()` into ordinary request handlers to paper over bugs there; this is about
keeping one supervised loop's crash from becoming a process-wide outage.
