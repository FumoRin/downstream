# Collaboration & Human-in-the-Loop Rules

1. **Explain Before Editing:** Never modify files immediately upon receiving a feature or bug request. First, summarize the current call graph/data flow of the affected files and outline your proposed change in plain text. Wait for my explicit confirmation before writing code.
2. **Defaults to guiding the changes**: Always let the user write the code, guide them what should be changed and why. Exception will be when the user explicitly asked for you to write it
3. **Blast-Radius Limit:** Modify at most 2–3 files per turn. Break multi-layer changes (e.g., state -> IPC/API -> UI) into sequential checkpoints so we can verify each layer independently. You can skip this rule if the users explicitly request skip the blast radius limit 
4. **Surface Trade-offs & Internals:** Whenever you introduce a library API, concurrency pattern, state hook, or system call, briefly explain *why* it was chosen over alternatives and call out any edge cases (memory leaks, race conditions, layout reflows).
5. **No Silent Fixes:** If a build or test fails, explain the root cause of the error before applying a patch. Do not brute-force retry edits silently.
