# Contributing to URL Sluice

First off, thank you for considering contributing to URL Sluice! It's people like you that make URL Sluice such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* Use a clear and descriptive title
* Describe the exact steps which reproduce the problem
* Provide specific examples to demonstrate the steps
* Describe the behavior you observed after following the steps
* Explain which behavior you expected to see instead and why
* Include details about your configuration and environment

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

* A clear and descriptive title
* A detailed description of the proposed feature
* Any possible implementation details
* Why this enhancement would be useful to most URL Sluice users

### Pull Requests

* Fill in the required template
* Follow the Go coding style and conventions
* Include appropriate tests
* Update documentation for any new features
* Ensure all tests pass locally before submitting

## Development Process

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. If you've changed APIs, update the documentation
4. Ensure the test suite passes
5. Make sure your code follows the existing style
6. Issue that pull request!

## Local Development Setup

1. Install Go 1.21 or higher
2. Clone your fork of the repo
3. Run `make deps` to install dependencies
4. Run `make test` to run tests

## Styleguides

### Git Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line

### Go Styleguide

* Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
* Run `gofmt` before committing
* Document all exported functions and types
* Include tests for new code

## Additional Notes

### Issue Labels

* `bug`: Something isn't working
* `enhancement`: New feature or request
* `good first issue`: Good for newcomers
* `help wanted`: Extra attention is needed
* `question`: Further information is requested

Thank you for contributing to URL Sluice! 