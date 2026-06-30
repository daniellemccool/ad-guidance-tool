# Releasing `adg`

`adg` is the CLI; `write-adr` (in `tools/adr-plugin/`) is the Claude Code plugin that drives it.
They ship as **one version-locked bundle**, and a single git tag drives everything.

## The version-lock invariant

One tag `vX.Y.Z` simultaneously fixes:

1. the GitHub Release assets — `adg_<os>_<arch>[.exe]` + `checksums.txt` (built by `.goreleaser.yaml`);
2. the CLI's `adg --version` (the tag, minus the `v`, injected via `-ldflags`);
3. `tools/adr-plugin/.claude-plugin/plugin.json` `version` (**must** equal the tag minus `v`);
4. the d3i-skills marketplace `git-subdir` `ref`.

The plugin's `bin/adg` wrapper reads its version out of `plugin.json` and downloads
`adg vX.Y.Z` from the Release — so **`plugin.json` version == release tag minus `v`** is a hard
requirement, not a convention. Mismatch ⇒ the wrapper's first download 404s.

## Cutting a release

1. Bump `tools/adr-plugin/.claude-plugin/plugin.json` `version` → `X.Y.Z`; commit on `main`.
2. `git tag vX.Y.Z && git push origin vX.Y.Z`
3. The `release` workflow runs goreleaser and publishes the GitHub Release with the six
   `adg_<os>_<arch>` assets + `checksums.txt`.
4. Verify: the Release page lists all assets, and a downloaded asset prints `X.Y.Z`:
   `./adg_linux_amd64 --version`.
5. Open a one-line PR in the **d3i-skills** repo bumping the `write-adr` entry's `ref` to `vX.Y.Z`
   (this is the team's adoption event — see the marketplace entry below).

> Do not bump the d3i-skills `ref` until step 3's assets are live, or installs will 404 on first
> `adg` call.

## How consumers receive it

- **Plugin skills:** the plugin's `bin/adg` is auto-added to PATH; on first call it lazily downloads
  the matching `adg vX.Y.Z` into `${CLAUDE_PLUGIN_DATA}` (cached, persists across updates).
- **Governed-repo hooks** (PreToolUse/Stop, the git pre-commit hook, the `adr` wrapper): these run
  outside the plugin's PATH and need a system `adg` — install with `install.sh` (see README).

## d3i-skills marketplace entry

```json
{ "name": "write-adr", "source": "git-subdir",
  "url": "daniellemccool/ad-guidance-tool", "path": "tools/adr-plugin", "ref": "vX.Y.Z" }
```

## Local verification (no real release)

- Version wiring: `go build -ldflags "-X adg/cmd.version=v9.9.9-test" -o /tmp/adg . && /tmp/adg --version`
- Release config: `goreleaser check` then `goreleaser release --snapshot --clean` (inspect `dist/`).
