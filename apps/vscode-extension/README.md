Extension responsibilities
- activate/deactivate with VS Code lifecycle
- start daemon process
- connect transport
- capture editor state
- capture selection / open file / diagnostics
- send mic audio chunks
- render transcript UI
- render pending diffs
- apply editor-side decorations
- expose commands / hotkeys
- maybe inline chat / webview panel

The extension should not:
- parse repo deeply
- decide edit strategy
- do heavy indexing
- own voice intent logic
- perform complex command orchestration