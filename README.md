# cakectl

A command line tool for [StatusCake](https://www.statuscake.com). Create,
inspect and manage monitoring from the terminal, in the way `gcloud` manages
Google Cloud.

## Authentication

Create an API token in the StatusCake dashboard, then save it:

```console
$ cakectl auth login --api-token "$TOKEN"
```

The token is stored in `~/.config/statuscake/cakectl.yaml`, readable only by
you. Instead of saving it you can set `STATUSCAKE_API_TOKEN`, or pass
`--api-token` to any command. A flag wins over the environment, and the
environment wins over the file.

## Commands

| Group      | Command              | Manages                                    |
| ---------- | -------------------- | ------------------------------------------ |
| Monitoring | `uptime`             | Uptime checks, with alerts, history, periods |
| Monitoring | `heartbeat`          | Heartbeat checks                           |
| Monitoring | `pagespeed`          | Pagespeed checks, with history             |
| Monitoring | `ssl`                | SSL certificate checks                     |
| Alerting   | `contact-group`      | Contact groups                             |
| Alerting   | `maintenance-window` | Maintenance windows                        |
| Reference  | `locations`          | Where uptime and pagespeed checks run from |

Each resource has `create`, `update ID`, `delete ID`, `describe ID` and `list`.
Run `cakectl <command> --help` for the flags and examples.

## Describing a resource

A resource is described with flags, with a YAML or JSON file, or both. Flag
names are the field names with dashes, so the field `website_url` is
`--website-url`.

```console
# From flags
$ cakectl uptime create --name "Home page" --test-type HTTP \
    --website-url https://example.com --check-rate 300

# From a file
$ cakectl uptime create --filename uptime.yaml

# From standard input
$ cat uptime.yaml | cakectl uptime create --filename -

# From a file, overriding one field
$ cakectl uptime create --filename uptime.yaml --name "Staging"
```

```yaml
# uptime.yaml
name: Home page
test_type: HTTP
website_url: https://example.com
check_rate: 300
contact_groups: ["1234"]
tags: [prod, web]
```

Files are read strictly: a field that does not exist is an error, so a typo
cannot silently leave a check configured differently from what you meant.

### Updating

An update changes only the fields you supply. Everything else is left alone.

```console
$ cakectl uptime update 123 --paused
$ cakectl uptime update 123 --contact-groups ""   # empty list: clears the field
```

`describe` prints a resource in the shape `-f` reads, so you can edit a
resource as a file:

```console
$ cakectl uptime describe 123 --format yaml > uptime.yaml
$ $EDITOR uptime.yaml
$ cakectl uptime update 123 --filename uptime.yaml
```

Basic authentication credentials and monitoring regions can be set but are
never returned by the API, so they do not appear in a description.

## Output

Output is plain text by default. Use `-f`/`--format` for `json`, `prettyjson` or `yaml`:

```console
$ cakectl uptime list --status down --format json | jq '.[].name'
```

Listings leave out secrets and personal data: heartbeat ping URLs, and contact
group email addresses, phone numbers and ping URLs. `describe` shows them.

## Deleting

`delete` asks for confirmation. When input is not a terminal, as in a script,
it refuses unless you pass `--yes`.

```console
$ cakectl uptime delete 123
Delete uptime check 123? [y/N]: y
uptime 123 deleted

$ cakectl uptime delete 123 --yes
```

## Configuration

| Setting           | Flag                | Environment                  | Default |
| ----------------- | ------------------- | ---------------------------- | ------- |
| `api_token`       | `--api-token`       | `STATUSCAKE_API_TOKEN`       |         |
| `rps`             | `--rps`             | `STATUSCAKE_RPS`             | `4`     |
| `retries`         | `--retries`         | `STATUSCAKE_RETRIES`         | `3`     |
| `min_backoff`     | `--min-backoff`     | `STATUSCAKE_MIN_BACKOFF`     | `1s`    |
| `max_backoff`     | `--max-backoff`     | `STATUSCAKE_MAX_BACKOFF`     | `30s`   |
| `custom_endpoint` | `--custom-endpoint` | `STATUSCAKE_CUSTOM_ENDPOINT` |         |

`cakectl config get|set|unset|view` edits the file. `--config` names a
different file.

## Development

```console
$ go test -race ./...
```

Commands are tested against a fake StatusCake API using the real SDK client.
Releases inject the version with `-X .../internal/cmd.version` (see
`goreleaser.yaml`).
