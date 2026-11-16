# Documentation capture inputs

This private directory contains the fixture config, target catalog, saved state, theme, and [asset manifest](manifest.md) used by `cmd/wgt-docs`.

See the [capture guide](../04-development/03-docs-capture-pipeline.md) for tool requirements, regeneration commands, and timing validation. Keep scenario names and output filenames in `internal/ui/doc_capture.go`; do not duplicate the registry here.

The `_` prefix keeps these inputs out of the published site.
