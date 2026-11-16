# Unreleased

## Added

- Named Warpgate profiles with a `p` picker, in-app creation and renaming, and a `--profile` startup override.
- Separate cached targets, recent targets, and transfer drafts for each profile, with the active name shown in the header.
- Dated backups and restore instructions when migrating legacy data or renaming profiles. Original files stay in place.

## Fixed

- Refresh results from an earlier connection cannot replace the active profile's targets.
- Switching profiles pauses target actions and restores auto-detected SSH ports from usable cache entries.
- Profile renames work offline and preserve the current snapshot and local state.
- Config saves run one at a time, reject stale completions, and finish before a queued quit request exits.
- Profile picker actions remain available with custom navigation keys. `last_profile` is reserved regardless of case.
- Profiles with no SSH host derive it from the server URL.
- Quoted browser command templates keep the approval URL in one argument.

## Compatibility

Older top-level `server` and `ssh` configs load as a profile named `default` and use the new shape on their next save. Startup selects `--profile`, then a valid last-used profile, then `default_profile`. See [Configuration](../../reference/configuration/) for backup locations and recovery steps.
