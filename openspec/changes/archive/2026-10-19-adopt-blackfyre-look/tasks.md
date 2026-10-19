# Tasks

## 1. Look

- [x] 1.1 Add `internal/ui/fonts` (Archivo SemiBold, JetBrains Mono Regular, OFL texts) and use it for the shaper; replace "Go Mono" with JetBrains Mono.
- [x] 1.2 Replace the palette with the Blackfyre tokens and button colours, drop the system accent, keep high contrast. Verify with the legibility test over light, dark and both high-contrast variants, including every button kind.
- [x] 1.3 Restyle the shared widgets: tracked section titles, chips and button labels, tactile buttons. Verify the existing UI tests pass and check offscreen renders in light and dark.

## 2. Verification

- [x] 2.1 Update the README's appearance paragraph, then run `gofmt -l .` (empty), `go vet ./...` (Linux and Windows), `go test ./...` and the Windows GUI build.
