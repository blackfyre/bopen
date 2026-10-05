# Design

## Context

See proposal.md. Pre-selection lives in `app.Preselect`, and the window decision in `app.NeedWindow`. `ui.Model.Refresh` re-derives the browsers and the analysis from `Env`. Settings changes go through `prefs.UpdateConfig`.

## Decisions

- **Model.** `prefs.SiteRule{ID, Hosts, Browser, Direct}` is stored in `Config.Sites` (`[[sites]]`). It has `Validate`, `AddSiteRule` (ID `s-` plus 6 hex), `SetSiteRule` and `DeleteSiteRule`, mirroring the user rules.
- **Matching.** `app.MatchSite(cfg, visible, host)` returns the first valid rule with a matching pattern (`path.Match` on lower-case strings) whose browser is in the visible list. The host comes from `Analysis.Clean(Defaults())`, which is the default destination, so redirect wrappers match by their target. It is fixed when the link is analysed, so toggling suggestions in the inspector doesn't move the selection under the user.
- **Pre-selection.** `app.Preselect` gains a preferred identity, which wins when present in the list.
- **Window decision.** `app.Situation` gains `Direct bool`. `NeedWindow` returns false when `Direct` is set and nothing must be reported. The silent path opens `Result()`, which applies the defaults, so the direct case needs no separate code.
- **Inspector.** The Open-in card ends with either:
  - a muted line, "Chosen by your site rule for `<patterns>`", when a rule applies; or
  - a checkbox, "Always open `<host>` in this browser".

  `OpenSelected` saves the rule after a successful launch. A failed save is reported as a notice, but doesn't block opening.
- **Settings.** A "Site rules" card lists `patterns → browser` with an "Opens directly" chip, plus Edit, Delete and Add buttons. A `siteForm` overlay provides a host-patterns editor (comma-separated), a browser radio list of all discovered browsers (hidden ones marked) and a "Open directly (tracking is still removed)" checkbox. It shares the overlay, scrim and keyboard handling with the rule form.

## Risks / Trade-offs

- **[Direct opening hides suggestions]** → The user chooses it per rule, the label states that tracking is still removed, and affiliate tags stay as they are by default. Problems always show the window.
