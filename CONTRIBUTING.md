# Contributing

We would love to see the ideas you want to bring in to improve this project.
Before you get started, make sure to read the guidelines below.

## Issues

If you have an idea how to improve this project, or if you find a bug, create an issue to let us know.
Please format your issue titles according to the [conventional commits guidelines](https://www.conventionalcommits.org/en/v1.0.0/).

## Code Contributions

### Committing

Please use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) for your commits.

##### Types
We use the following types:

- **build**: Changes that affect the build system or external dependencies
- **ci**: changes to our CI configuration files and scripts
- **docs**: changes to the documentation
- **feat**: a new feature
- **fix**: a bug fix
- **perf**: an improvement to performance
- **refactor**: a code change that neither fixes a bug nor adds a feature
- **style**: a change that does not affect the meaning of the code
- **test**: a change to an existing test, or a new test

### Fixing a Bug

If you're fixing a bug, if possible, add a test case for that bug to ensure it's gone for good.

### Code Style

Make sure all code passes the golangci-lint checks.
If necessary, add a `//nolint:{{name_of_linter}}` directive to the line or block to silence false positives or exceptions.

### Testing

If possible and appropriate you should fully test the code you submit.
Each function should have a single test, 
which either tests directly or is split into subtests, preferably table-driven.

Usually, a there is either a single table named `cases`, or multiple tables, 
e.g. `successCases` and `failureCases`.

Each table should be a slice of an anonymous struct, usually with an `except` field.
