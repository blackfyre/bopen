# Design

## Context

This builds on `add-link-inspector-core` (rule schema, span-based analysis) and `add-settings-view` (rule IDs, read-modify-write saves, view switching). See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- One rule model and one matcher for built-in and user rules.
- Creating a rule in context, in a few seconds, without leaving the inspector.

**Non-Goals:**
- User-defined redirect wrappers. They need path and target-parameter configuration and validation that isn't justified yet.
- Regex patterns in user rules; only exact names and a trailing `*`.
- Import and export of rule sets.

## Decisions

### Rule sources as an ordered list

`internal/clean` takes an ordered slice of rule sources (`user`, then `builtin`; `clearurls` is appended later). Analysis collects all matches per span, then keeps the match from the earliest source. Disabled built-in IDs are filtered out before matching. This gives precedence one obvious place to live, and lets `add-clearurls-rules` add a source without touching the matcher.

### Same schema, separate location

User rules reuse the built-in rule struct and are stored under `[rules.user]` in `config.toml`, next to `rules.disabled`:

```toml
[rules]
disabled = ["utm"]

[[rules.user]]
id     = "u-3f9a1c"
kind   = "tracking"
param  = "ref"
hosts  = ["news.example.com"]
reason = "Referrer tracking on this news site."
```

IDs are a `u-` prefix plus 6 random hex characters. That's enough to avoid collisions in a hand-sized list, and the prefix keeps them visibly distinct from built-in IDs.

### Context menus via gio-x

`gioui.org/x/component` provides `ContextArea` and `MenuState`. Interactive rich-text spans only report primary clicks and hover, so the rendered link is wrapped in one `ContextArea`. Each parameter segment (flagged or not, using the offsets the analyser records for every query segment) is an interactive span. Its hover state tells which parameter a right-click targets. Right-clicking anywhere else does nothing. Each suggestion row has its own `ContextArea`.

### Rule form as an overlay

The form is a modal overlay inside the inspector, not a navigation to settings, so the link stays visible while the user writes the reason. The settings view reuses the same form widget for add and edit.

### "This host" scope

"This host" stores the exact host of the URL that contains the parameter: for a parameter inside an unwrapped redirect target, that's the target's host. Wildcard host editing (for example `*.example.com`) is available in the settings form only, to keep the inline form short.

## Risks / Trade-offs

- **[Over-broad user rules]** "Any host" on a common name such as `id` could break many sites. → Suggestions remain toggleable, the form shows the scope prominently, and "this host" is the default.
- **[`gioui.org/x` stability]** It is less stable than core Gio. → It is already pinned to the matching Gio release, and the context-menu surface used is small enough to replace if needed.
