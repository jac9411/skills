# Clean Code & TDD Detailed Guidelines

## The Three Laws of TDD

1. You are not allowed to write any production code unless it is to make a failing unit test pass.
2. You are not allowed to write any more of a unit test than is sufficient to fail; and compilation failures are failures.
3. You are not allowed to write any more production code than is sufficient to pass the one failing unit test.

## Function Size and Structure

### The 4-20 Rule
- **4 lines:** The minimum size for a function to be meaningful, excluding trivial wrappers.
- **20 lines:** The maximum size. If a function exceeds this, it is doing too much. Split it.
- **Indentation:** Limit to one or two levels of nesting. Use guard clauses to reduce nesting.

## Naming Conventions

- **Variables:** Nouns (e.g., `userProfile`, `retryCount`).
- **Functions/Methods:** Verbs or verb phrases (e.g., `calculateTotal`, `isUserAuthorized`).
- **Booleans:** Start with `is`, `has`, `can`, or `should` (e.g., `isValid`, `hasPermission`).
- **Classes:** Nouns, avoid generic suffixes like `Manager`, `Processor`, `Helper`.

## Single Responsibility Principle (SRP)

A function should do one thing, do it well, and do it only. If a function contains steps that could be described with a "then" (e.g., "do A then do B"), it should probably be split.

## Refactoring Patterns

- **Extract Method:** For code blocks that do something specific.
- **Inline Method:** For methods that are not providing enough value (e.g., < 4 lines without added clarity).
- **Rename Variable/Method:** To improve descriptive power.
- **Replace Conditional with Polymorphism:** For complex `switch` or `if-else` chains.
