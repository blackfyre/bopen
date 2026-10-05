# Tasks

## 1. Command

- [x] 1.1 Add `clean` to argument parsing and usage, and implement `runClean` (arguments or stdin, `--explain`, invalid lines echoed, exit status). Verify with unit tests for each spec scenario and a dispatch test.
- [x] 1.2 Keep valid redirected standard handles in the Windows `attachConsole`. Verify `GOOS=windows go vet`, and add a Windows CI test running the built `bopen.exe clean` through a pipe and checking its output.

## 2. Inspector

- [x] 2.1 Add the Copy button and `C` shortcut with "Copied" feedback. Verify with a UI test that `C` issues a clipboard write of the result (captured through the router), and with a render.

## 3. Verification

- [x] 3.1 Document `bopen clean` and Copy in the README. Run all gates, push, and confirm both CI jobs pass.
