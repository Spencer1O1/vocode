# Architecture

Vocode is a voice-driven AI code editing system built around a local daemon and a thin editor client.

## Core Idea
- Magical UX, deterministic core: user experience should feel seamless, while the system remains structured, predictable, and debuggable.

## High-Level Architecture
- **VS Code Extension (TypeScript)**
  - UI (panels, status, diff viewer)
  - Voice capture
  - Command surface
  - RPC client
- **Daemon (Go)**
  - RPC server
  - Agent / planner
  - Edit engine
  - Indexing engine
  - Workspace model
  - Speech providers
- **Transport:** stdio (JSON-RPC) between extension and daemon.

## Component Responsibilities
### VS Code Extension
- Capture user input (voice, commands).
- Display output (diffs, transcripts, status).
- Maintain UI state.
- Send requests to daemon.
- Apply edits returned by daemon.

**Non-responsibilities**
- No business logic.
- No edit planning.
- No code intelligence.

### Daemon (Go)
- Interpret user intent.
- Build execution plans.
- Generate structured edit actions.
- Validate edits.
- Maintain workspace model.
- Perform indexing.
- Execute commands.
- Interface with speech providers.

## System Flow
1. User speaks.
2. Speech -> text (streaming).
3. Intent parsing.
4. Planning.
5. Edit actions.
6. Validation.
7. Diff generation.
8. Apply edits.
9. UI updates.

## Key Subsystems
### 1) Agent / Planner
- Location: `apps/daemon/internal/agent/`.
- Responsibilities: interpret natural language; ask clarifying questions; generate structured plans; orchestrate subsystems.

### 2) Edit Engine
- Location: `apps/daemon/internal/edits/`.
- Responsibilities: convert plans into edit actions; anchor edits to code structure; generate diffs; validate before applying.
- Critical rule: never blindly rewrite entire files unless explicitly intended.

### 3) Indexing Engine
- Location: `apps/daemon/internal/indexing/`.
- Responsibilities: fast file search (ripgrep); symbol extraction; file summaries; incremental updates via watchers.

### 4) Workspace Model
- Location: `apps/daemon/internal/workspace/`.
- Responsibilities: track files and structure; maintain snapshots; respect ignore rules; provide context to agent.

### 5) RPC Layer
- Locations: `apps/daemon/internal/rpc/`, `apps/vscode-extension/src/client/`.
- Transport: stdio (JSON-RPC).
- Responsibilities: request/response handling; streaming events; routing to handlers.

### 6) Speech System
- Locations: `apps/daemon/internal/speech/`, `apps/vscode-extension/src/voice/`.
- Responsibilities: streaming STT; provider abstraction (ElevenLabs, Whisper.cpp); partial transcript handling.

## Communication Model
- **Transport:** stdio-based JSON-RPC; extension spawns daemon; persistent connection.
- **Rules:** stdout carries protocol only; stderr is for logs only.
- **Example flow:** extension calls `startVoice()` -> RPC request; daemon streams transcript events; extension updates UI live.

## Directory Boundaries
- **Extension** (`apps/vscode-extension/src/`)
  - `commands/` - command entrypoints
  - `client/` - RPC client
  - `daemon/` - process spawning and paths
  - `ui/` - panels and views
  - `voice/` - microphone and audio
- **Daemon** (`apps/daemon/internal/`)
  - `agent/`
  - `edits/`
  - `indexing/`
  - `workspace/`
  - `rpc/`
  - `speech/`

## Runtime Model
- **Startup:** extension activates; daemon binary is resolved; daemon process is spawned; RPC connection is established.
- **Steady state:** daemon runs continuously; extension sends requests; daemon emits events.
- **Shutdown:** extension disposes daemon process; daemon exits cleanly.

## Design Principles
- Magical UX, deterministic core: UX should feel effortless; internals must be structured and predictable.
- Structured edits over text generation: edits are explicit operations; diffs are inspectable; results are reproducible.
- Daemon-first intelligence: all “smart” logic lives in the daemon; the extension stays simple.
- Local-first: no required cloud dependency; fast iteration; private by default.
- Incremental everything: streaming speech; incremental planning; progressive UI updates.

## Future Evolution
- AST-aware editing via tree-sitter.
- LSP integration for semantic understanding.
- Multi-file coordinated edits.
- Collaborative sessions.
- Plugin system for tools/providers.

## Mental Model
Think of Vocode as a real-time, voice-driven compiler for code changes where input = natural language, output = structured edits, execution = a deterministic pipeline.

## Summary
- Extension = interface.
- Daemon = brain.
- Protocol = glue.
- Everything should reinforce that separation.
