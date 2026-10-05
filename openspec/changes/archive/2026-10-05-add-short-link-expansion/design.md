# Design

## Context

See proposal.md. The inspector already runs background work (ClearURLs refresh) with results delivered over a channel and a window invalidate. `Model.Refresh` re-derives analysis, host and site rule from `Analysis.URL`.

## Decisions

- **Shortener list.** `shortlinks.Hosts` is a curated set of exact hosts: `bit.ly`, `bitly.com`, `t.co`, `tinyurl.com`, `lnkd.in`, `ow.ly`, `buff.ly`, `is.gd`, `v.gd`, `rebrand.ly`, `cutt.ly`, `rb.gy`, `t.ly`, `shorturl.at`, `tiny.cc`, `s.id`, `amzn.to`, `amzn.eu`, `a.co`, `aka.ms`, `trib.al`, `dlvr.it`, `ift.tt`, `fb.me`, `spoti.fi`, `apple.co`, `tinyurl.is`, `bl.ink`, `short.io`, `qrco.de`. A leading `www.` is ignored. The list is small and conservative, so bopen never contacts an ordinary site.
- **Expander.** `shortlinks.Expander{Client, IsShortener}` uses an `http.Client` whose `CheckRedirect` returns `http.ErrUseLastResponse`, so every hop is decided by bopen, with no cookie jar. For each hop it sends `HEAD`, retries with `GET` on 405 or 501, and closes the body unread. It requires a 3xx status with a `Location`, which is resolved against the current URL. If the next host is not a shortener, that location is the result. The whole call runs under a 5-second context. `IsShortener` is injectable, so tests can use `httptest` hosts.
- **Preference.** `prefs.Config.ExpandShortLinks bool` (`expand_short_links`).
- **Inspector.** `Env.Expand func(ctx, url) (string, error)`; nil means disabled. When enabled and `shortlinks.IsShortener(host of the inspected link)`, the Link card shows an **Expand** button. Clicking it starts a goroutine and disables the button ("Expanding…"). The result arrives on a channel. On success, `Model.Replace(url)` validates and re-analyses the link, resets the toggles, sets `ExpandedFrom`, and refreshes the site rule and pre-selection. On failure, the error is shown as a notice. The Link card shows "Expanded from `<short link>`" once replaced.
- **Settings.** A "Short links" card with the toggle and the privacy sentence.

## Risks / Trade-offs

- **[Shortener returns HTML instead of a redirect for non-browser agents]** → Reported as "could not expand". The link stays usable.
- **[List drift]** → Unknown shorteners are simply not offered. The list is easy to extend.
