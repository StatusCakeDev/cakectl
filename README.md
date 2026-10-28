# cakectl

[![Support][support-badge]][support]

[support-badge]: https://img.shields.io/badge/support-cakectl-0B8FE6.svg?style=flat
[support]: https://developers.statuscake.com/guides/cli/introduction

- Website: [statuscake.com](https://www.statuscake.com)
- Support: [developers.statuscake.com](https://developers.statuscake.com/guides/cli/introduction)

A command line tool for [StatusCake](https://www.statuscake.com). Create,
inspect and manage monitoring from the terminal.

## Prerequisites

You will need a StatusCake API token. You can view and manage your API tokens
from the [StatusCake account panel](https://app.statuscake.com/User.php).

## Installation

Grab the binary for your system from the [releases
page](https://github.com/StatusCakeDev/cakectl/releases).

Alternatively, install with Go:

```bash
go install github.com/StatusCakeDev/cakectl@latest
```

> [!NOTE]
> Installing from source using `go install` will build from the master branch
> and won't report version information via `cakectl --version`.

### Documentation

Full documentation is available on the [developer portal][support].

## License

This project is licensed under the [Mozilla Public License](LICENSE).
