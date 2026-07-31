# Design: Run Suite & Fix Broken Tests

## Technical Approach

**Source-is-truth** (decision #178, binding). All 20 broken tests are updated to reflect current source behavior. The ONLY source change is the cosmetic `perfsprint` lint fix in `internal/logs/message.go:58` (`fmt.Sprintf("%d", x)` → `strconv.Itoa(x)`). No source-bug fixes, no interface changes, no mock regeneration.

Rationale: the user explicitly chose source as truth, overriding the proposal's hybrid strategy. Bugs like `PrevField` calling `nextField()`, `paneBg` using `BorderForeground`, `Resize` using `height-6`, and `Style()` returning the same style for all LogTypes are current behavior and tests must align.

## Architecture Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Source code changes | Only `strconv.Itoa` lint fix in `message.go` | Binding decision #178 — no source-bug fixes |
| Mocks / interfaces | None regenerated | No ports changed in last 5 commits (verified via `codegraph_explore` of `internal/ui/ports`) |
| Test helper for ANSI | Add `stripANSI` helper in test files (or inline regex) | Required for asserting `Details` string content in `processor_test.go` |
| `Margin(1,2)` layout | Source stays, tests expect header on line 1 | Intentional layout per `internal/ui/app_view.go:24` |
| Per-file commits | One commit per test file + one for lint fix | Clean rollback boundary; `git revert` per file if needed |

## Data Flow (no production change)

The test update does not alter runtime data flow. The only runtime change is `message.go:58` (string format → `strconv.Itoa`) — identical output, no observable behavior shift.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/logs/message.go` | Modify | `fmt.Sprintf("%d", res.StatusCode)` → `strconv.Itoa(res.StatusCode)`; swap `fmt` import for `strconv` |
| `internal/processor/processor_test.go` | Modify | 4 assertion updates (see below) |
| `internal/ui/app_view_test.go` | Modify | 3 header-line assertions: `firstContent == 0` → `1` (because of `Margin(1,2)`); `lines[0]` non-empty → `lines[1]` |
| `internal/ui/views/settings_test.go` | Modify | 8 test updates (see below) |
| `internal/ui/components/detailed_list_test.go` | Modify | 3 windowing/line-count assertion updates |
| `internal/ui/components/log_message_renderer_test.go` | Modify | 3 assertion updates |

## Per-File Test Plan

### `processor_test.go` (4 changes)

- **L171** `assert.Equal(t, 200, m.BadgeIcon)` → `assert.Equal(t, "200", m.BadgeIcon)` — `BadgeIcon` is now `string`.
- **L237** `assert.Equal(t, 404, m.BadgeIcon)` → `assert.Equal(t, "404", m.BadgeIcon)`.
- **L238** `assert.Equal(t, "POST https://example.com/missing", m.Text)` → `assert.Equal(t, "GET https://example.com/missing", m.Text)` — mock now returns `Method: "GET"`; source `NewHTTPMessage` uses `res.Method + " " + res.URL`.
- **L239** `assert.Equal(t, []byte(`{"error":"not found"}`), m.Details)` → `assert.Contains(t, m.Details, "not found")` — `Details` is now a `string` produced by `pretty.Color(pretty.Pretty(...), nil)`. Add ANSI-strip helper or assert substring; the test contract is the body payload is preserved, not byte-exact.

### `app_view_test.go` (3 changes)

- **L36-37** `assert.NotEqual(t, "", strings.TrimSpace(lines[0]), ...)` → index `lines[1]`.
- **L83, L186** `assert.Equal(t, 0, firstContent, ...)` → `assert.Equal(t, 1, firstContent, ...)`. Update message strings to mention `Margin(1,2)` instead of "line 0".

### `settings_test.go` (8 changes)

- **L101 `TestSettingsView_SliderFocusedOnConstruction_PlusKeyIncrements`** — default `focusPane == paneList` swallows `+`/`-` keys into `profileList.Update`. Fix: Tab once (toggle to `paneForm`) before pressing `-`/`+`.
- **L196 `TestSettingsView_SliderAcceptsKittyProtocolPlusKey`** — same fix: Tab to `paneForm` before the `-`/`+` sequence.
- **L340 `TestSettingsView_ShiftTabInFormCyclesBackward`** — source `PrevField` calls `nextField()`. From `urlField`, `nextField()` lands on `methodField` (not `sliderField`). Fix: assert `methodField` and add `// NOTE: source binds Shift+Tab to nextField()` comment.
- **L358 `TestSettingsView_ShiftTabWrapsFromCsvToHeaders`** — source: `csvFieldsField → nextField() → sliderField`; second press: `sliderField → urlField`. Update both assertions.
- **L462 `TestSettingsView_Resize_UpdatesAllSizesInOnePass`** — `v.viewport.Height() == 14` (=20−6) and `v.profileList.Height() == 14` (not 18 / 20).
- **L501 `TestSettingsView_FocusedPaneBg_AppliedToList`** — `paneBg` uses `BorderForeground` (escape `\x1b[38;` not `\x1b[48;`). Add a `borderForegroundEscapeSeq(c)` helper mirroring the existing `backgroundEscapeSeq`; rename test or update the helper.
- **L519 `TestSettingsView_FocusedPaneBg_AppliedToForm`** — same.
- **L1075 `TestSettingsView_Resize_AccountsForOwnMargin`** — assert `height-6 == 14`; update regression comment to reference `height-6` rather than `height-2`.

### `detailed_list_test.go` (3 changes)

- **L283 `height clips to window`** — `PaddingBottom(1)` adds 1 line per collapsed item, so `strings.Split(out, "\n")` returns 6 lines for 3 items. Change `assert.Len(..., 3)` to `assert.Len(..., 6)` and keep `assert.Contains(i0/i1/i2)`.
- **L298 `cursor at bottom scrolls window`** — same: 3 items × 2 lines = 6.
- **L314 `autoScroll pins to tail`** — same.
- **L341 `expanded item uses extra lines`** — already asserts `NotContains "b"` and `Contains "line1"/"line2"` (no exact line count), but assert `Contains "a"` not `Contains "detail-of:a"` (the title uses the item string). Wait — current assertion is `Contains(out, "a")` which is a substring of "detail-of:a". Should still pass. **No change required** for this subtest; just confirm.

### `log_message_renderer_test.go` (3 changes)

- **L33 `TestLogMessageRenderer_Title_FormatsHTTPRequestLine`** — source includes method. Change `assert.NotContains(t, got, "POST")` to `assert.Contains(t, got, "POST")`. Update test comment.
- **L41 `TestLogMessageRenderer_Title_FreeFormIncludesIconAndKind`** — source `NewGeneralMessage` ignores `kind` arg (`Text: title` only). Drop `assert.Contains(t, got, "Cancelation")`; keep `💀` and `"read 5 lines"`. Add `// NOTE: source ignores kind arg` comment.
- **L95 `TestLogMessageRenderer_Style_ColorsByLogType`** — source `Style()` returns `logRowStyle` for all LogTypes. Flip `assert.NotEqual` pairwise to `assert.Equal`; update test name to `TestLogMessageRenderer_Style_AllLogTypesShareBaseStyle` and add comment.

## Interfaces / Contracts

**No new interfaces, no struct changes, no mock regeneration.** `make mocks` is NOT needed (`make mocks` output unchanged; verified `internal/ui/ports` unchanged in last 5 commits).

## Testing Strategy

| Layer | What | Approach |
|---|---|---|
| Unit | 20 assertion updates | Direct `assert.Equal/Contains` rewrites per per-file plan |
| Helpers | ANSI-strip regex `ansi := regexp.MustCompile(`\x1b\[[0-9;]+m`)` | Local helper inside `processor_test.go`; small enough to inline |

## Threat Matrix

**N/A** — no routing, shell, subprocess, VCS/PR automation, executable-file classification, or process-integration boundary. Pure test-update + one cosmetic source line.

## Migration / Rollout

No migration. Apply as one PR (per `delivery_strategy = none`): 5 test files + 1 lint fix. Recommended commit sequence:

1. `chore(lint): use strconv.Itoa in logs/message.go` (cosmetic source change)
2. `test(processor): align assertions with current source types`
3. `test(ui): align app_view_test with Margin(1,2) header layout`
4. `test(settings): align tests with two-pane source behavior`
5. `test(detailed_list): align windowing tests with PaddingBottom(1)`
6. `test(log_message_renderer): align with current Style/Title/Title contracts`

Each commit independently reverts cleanly. The lint fix is a single-line change and stays isolated.

## Open Questions

- **None blocking.** Decision #178 already resolved all open questions from the proposal (Source is truth; tests reflect current behavior).

## Verification Strategy

1. `make test` — expect 229/0 passing (was 209/20 failing).
2. `make lint` — expect 0 issues (was 1 perfsprint).
3. Spot-check: rerun the 8 settings tests + 4 processor tests + 3 app_view tests + 3 detailed_list tests + 3 renderer tests in isolation (`go test -run TestName ./...`) to confirm no flakiness.
4. Confirm 209 previously-passing tests still pass (no regression introduced by the test edits).
5. `make build` to ensure the cosmetic `strconv.Itoa` change compiles cleanly (the `fmt` import is now unused — remove it).
