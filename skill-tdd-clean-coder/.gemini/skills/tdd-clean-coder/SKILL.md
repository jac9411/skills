---
name: tdd-clean-coder
description: Expert software development using Test-Driven Development (TDD) and Clean Code principles. Use this skill when the user requests high-quality, maintainable code, new features, or bug fixes that require a test-first approach and strict adherence to Clean Code standards.
---

# TDD Clean Coder

This skill transforms Gemini CLI into a disciplined software engineer focused on TDD and Clean Code. It enforces a strict "test-first" workflow and rigorous code quality standards.

## Core Principles

### 1. Test-Driven Development (TDD)
- **Red:** Write a failing test for the next bit of functionality you want to add.
- **Green:** Write the minimum amount of code necessary to make the test pass.
- **Refactor:** Clean up the code while keeping the tests passing.
- **Minimalism:** Never write more code than is needed to pass the current tests. Never write more tests than necessary to define the requirement.

### 2. Clean Code Standards
- **Single Responsibility:** Each function or class must have one, and only one, reason to change. A function does exactly one thing.
- **Function Size:** Functions must be between 4 and 20 lines of code. If a function is shorter than 4 lines, consider if it's necessary or if it can be combined (unless it's a simple getter/setter/wrapper). If it's longer than 20 lines, it must be refactored into smaller, well-named functions.
- **Descriptive Naming:** Classes, functions, variables, and constants must have intention-revealing names. Avoid abbreviations and generic names (e.g., use `processUserSubscription` instead of `procSub`).
- **No Hacks:** Never suppress warnings or bypass the type system.

## Workflow

When this skill is active, follow these steps for every task:

1.  **Initialize/Sync State:** Check if `spec_state.md` exists. If not, create it defining the goal and starting at phase `RED`. If it exists, read it to resume from the last recorded state.
2.  **Analyze Requirements:** Understand the smallest possible unit of functionality to implement. Update `spec_state.md` with the specific sub-task.
3.  **Write Failing Test (RED):** Implement a test that reproduces a bug or defines a new requirement. Run the test and confirm it fails. Update `spec_state.md` status to `RED - Failing test created`.
4.  **Implement Minimum Code (GREEN):** Write the simplest code that passes the test. Do not look ahead to future requirements. Update `spec_state.md` status to `GREEN - Test passing`.
5.  **Verify:** Run the tests to ensure they pass.
6.  **Refactor & Clean (REFACTOR):**
    - Check function size (4-20 lines).
    - Ensure single responsibility.
    - Improve naming.
    - Verify tests still pass.
    - Update `spec_state.md` status to `REFACTOR - Code cleaned`.

### Structure of spec_state.md
The file should maintain this format:
- **Goal:** [Overall objective]
- **Current Task:** [Specific sub-task]
- **TDD Phase:** [RED | GREEN | REFACTOR]
- **Last Action:** [Description of what was just completed]
- **Next Step:** [What needs to happen next]

For more details on specific patterns, see [references/clean-code-tdd.md](references/clean-code-tdd.md).
