# How To Contribute

Contributions are welcome as we strive to make this application as useful as
possible for everyone. However time is not always on our side, and changes may
not be reviewed or merged in a timely manner.

If this application is found to be missing in functionality, please open an
issue describing the proposed change - discussing changes ahead of time reduces
friction within pull requests.

## Prerequisites

You will need the following things properly installed on your computer:

- [Git](https://git-scm.com/)
- [Go](https://go.dev/) (1.27+)
- [Make](https://www.gnu.org/software/make/) (4.0+)

## Installation

If you wish to work on `cakectl`, you'll first need [Go](https://go.dev/)
installed on your machine (the version named in `go.mod`).

To compile the application, run `make build`. This will build the `cakectl`
binary, which can then be moved to the `$GOPATH/bin` directory to use it from
anywhere.

```sh
git clone <repository-url>` this repository
cd cakectl
make build
mv cakectl $GOPATH/bin/cakectl
```

You will need a StatusCake API token to try the binary against the real API.
Save it with `cakectl auth login`, or set `STATUSCAKE_API_TOKEN`.

### Pointing cakectl at another API

`--custom-endpoint` (or `STATUSCAKE_CUSTOM_ENDPOINT`, or `custom_endpoint` in
the config file) exists only in development builds, made with `make build
TAGS=dev`. A release build has no such setting and ignores it if set. Most tests
do not need the tag: they hand `adapter.Config` an `Endpoint`. Those that drive
the command through its config file are built with `-tags dev`.

## Running Tests

- `make test` vets and tests both the release and the development build.

The tests talk to a local fake of the API, so they need no API token and make no
real requests. Run `make test TESTARGS="-run TestName"` to run a single test.

## SDK Boundary

Exactly one package, `internal/adapter`, imports the StatusCake SDK. Everything
else works with the models in `internal/domain`, and the adapter reports a
missing resource as `domain.ErrNotFound` and a rejected token as
`domain.ErrUnauthorized`. Please keep SDK types out of the other packages.

## Making Changes

For additional contributing guidelines visit
[devhandbook.org](https://devhandbook.org/contributing)
