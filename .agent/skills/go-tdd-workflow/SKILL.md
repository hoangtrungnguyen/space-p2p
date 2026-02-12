---
name: Go TDD Workflow
description: Strictly enforces Red-Green-Refactor development cycle for Go projects.
---

# TDD Workflow Decision Tree

This skill enforces a strict Test-Driven Development lifecycle. You must follow these steps for any feature implementation.

1.  **Red Phase (Test Generation)**
    *   **Trigger**: When tasked with implementing a new feature or fixing a bug.
    *   **Action**: Write a **FAILING** unit test in a `_test.go` file.
    *   **Tools**:
        *   Use `mockgen` to create mocks for interfaces (especially for LiveKit services).
        *   Use `testify/assert` and `testify/suite` for assertions.
    *   **Constraint**: DO NOT write implementation code yet.

2.  **Execution Gate**
    *   **Action**: Execute `scripts/test_watcher.sh`.
    *   **Logic**:
        *   IF test **PASSES** immediately: **STOP**. Test is invalid or feature exists. Rewrite test to fail.
        *   IF test **FAILS**: **PROCEED** to Green Phase.

3.  **Green Phase (Implementation)**
    *   **Action**: Write the **MINIMUM** amount of Go code required to make the test pass.
    *   **Loop**: Run `scripts/test_watcher.sh` repeatedly until it returns exit code 0.

4.  **Refactor Phase**
    *   **Action**: Optimize code, clean up syntax, improve performance.
    *   **Constraint**: Must pass `scripts/test_watcher.sh` after every change.

5.  **Coverage Gate**
    *   **Action**: Run `scripts/check_coverage.sh`.
    *   **Logic**:
        *   **80% - 100%**: **PASS**. Proceed to commit/deploy.
        *   **50% - 79%**: **WARNING**. Generate more integration tests.
        *   **< 50%**: **FAIL**. STOP. Generate table-driven unit tests for edge cases.
