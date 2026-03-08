# Contributing

I love pull requests from everyone!

When contributing to this repository, please first discuss the change you wish
to make via [issue](https://github.com/dokadev/lazyprisma/issues) or any other
method with me before making a change.

If you've never written Go in your life, that's completely fine. Go is widely
considered an easy-to-learn language, and lazyprisma's codebase is modest in
size, so if you're looking for an open source project to gain dev experience,
you've come to the right place.

## All code changes happen through Pull Requests

Pull requests are the best way to propose changes to the codebase. I actively
welcome your pull requests.

Please do not raise pull requests from your fork's `main` branch: make a
feature branch instead. I'll sometimes push changes to your branch when reviewing a PR, and I often
can't do this if you use your main branch.

## Commit history

I value a clean and useful commit history, so please take some time to
organise your commits so that they make sense.

In particular:

- Refactorings and behaviour changes should be in separate commits. There are very few exceptions where this is not possible, but in practice they are very rare.
- Strive for minimal commits; every change that is independent from other changes should be in a commit of its own (with a good commit message that explains why the change is made).


## Go

### Formatting

I expect all code to be formatted with `gofmt` before it lands. Most editors
sort this out automatically on save, so just double-check yours does and you'll
be fine.

### Naming

I follow a few naming conventions throughout the codebase:

- Exported types and functions are `PascalCase`; unexported bits are `camelCase`.
- Interfaces get an `I` prefix (e.g. `IPopupHandler`, `IGuiCommon`).
- Config-style structs tend to have an `Opts` or `Config` suffix
  (e.g. `AsyncCommandOpts`, `AppConfig`).
- Sentinel errors start with `Err` (e.g. `ErrNotConnected`).

### Imports

I keep imports in two groups, separated by a blank line:

1. Standard library
2. External packages

### Error handling

- Stick with the plain `if err != nil` pattern. No fancy helper wrappers.
- Wrap errors with `fmt.Errorf("context: %w", err)` so the chain stays intact.
- If you've got a reusable error, pop it in as a sentinel with `errors.New` at
  package level.

### Comments

- I write comments in English, and I'd appreciate it if you did the same.
- Exported symbols should have a comment that starts with the symbol name
  (e.g. `// NewClient creates...`).

### Patterns in use

These patterns are already well established in the codebase. I'd rather you
followed them than introduced alternatives:

- **Trait composition**: embed `ScrollableTrait`, `TabbedTrait`, etc. into
  context types instead of duplicating behaviour.
- **Interface dispatch**: use type assertions on context interfaces for
  polymorphism (have a look at `keybinding.go`).
- **Callback fields**: pass behaviour via `func` fields in option structs
  rather than creating a new interface for every callback.
- **Atomic types**: use `atomic.Bool` / `atomic.Value` for cross-goroutine
  state. Don't reach for bare mutexes when a simple flag will do.

## Report bugs using GitHub Issues

I use GitHub issues to track public bugs. Report a bug by
[opening a new issue](https://github.com/dokadev/lazyprisma/issues/new). It's
that easy!


## Any contributions you make will be under the MIT Software License

In short, when you submit code changes, your submissions are understood to be
under the same [MIT License](http://choosealicense.com/licenses/mit/) that
covers the project. Feel free to drop me a line if that's a concern.

## Improvements

If you can think of any way to improve these docs, do let me know.
