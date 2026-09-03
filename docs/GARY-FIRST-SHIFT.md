# GARY First Shift: Telltail 0.3.0

## Overview
Telltail 0.3.0 provides robust, observable execution and evaluation for AI worker shifts, supporting both local runs and Google Cloud Batch remote execution.

## Yardmaster Waves & Execution Lifecycle
1. **Wave 1 (Initialization & Research):** Workspace setup, toolchain verification, and dependency checking.
2. **Wave 2 (Execution & Trajectory Recording):** Immutable append-only JSONL event logging with SHA-256 hash chains (`prev_hash`, `hash`, `seq`).
3. **Wave 3 (Detection & Profiling):** Deterministic evaluation of worker behavior, identifying bottlenecks, repetition, tool misuse, and boundary violations.
4. **Wave 4 (Mirror Shift & Scoring):** Workspace initialization (`mirror init`) and precision/recall evaluation (`mirror score`).

## Bottlenecks and Resilience
- **Command execution phases** typically represent the primary duration bottleneck during test suites and compilation.
- **Fail-closed verification** ensures any trajectory tampering or sequence break immediately invalidates the run.
- **Cross-platform compilation** guarantees parity across Linux and Windows/amd64 targets.
