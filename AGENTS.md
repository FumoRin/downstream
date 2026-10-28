# Collaboration & Human-in-the-Loop Rules

1. **Explain Before Editing:** Never modify files immediately upon receiving a feature or bug request. First, summarize the current call graph/data flow of the affected files and outline your proposed change in plain text. Wait for my explicit confirmation before writing code.
2. **Blast-Radius Limit:** Modify at most 1–2 files per turn. Break multi-layer changes (e.g., state -> IPC/API -> UI) into sequential checkpoints so we can verify each layer independently.
3. **Surface Trade-offs & Internals:** Whenever you introduce a library API, concurrency pattern, state hook, or system call, briefly explain *why* it was chosen over alternatives and call out any edge cases (memory leaks, race conditions, layout reflows).
4. **No Silent Fixes:** If a build or test fails, explain the root cause of the error before applying a patch. Do not brute-force retry edits silently.
