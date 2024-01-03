# Tailscale Healthcheck

This is a script that performs a healthcheck on an HTTP service running in a [Tailscale](https://tailscale.com) network, without requiring a Tailscale daemon. This makes it useful in a cloud or serverless environment, where there is not a persistent machine to be registered into the Tailscale network.

## Usage

The most straightforward way to use and deploy the script is via Docker:

```sh
docker run --rm \
-e TAILSCALE_TAILNET=<your admin email address> \
-e TAILSCALE_API_KEY=<an api key from https://login.tailscale.com/admin/settings/keys> \
-e HEALTHCHECK_MACHINE_NAME=<the hostname of the machine to check> \
-e HEALTHCHECK_PORT=<the port running the HTTP service> \
jasongwartz/tailscale-healthcheck
```

Or with a .env file:

```sh
docker run --rm --env-file .env jasongwartz/tailscale-healthcheck
```

You can also build the binary locally, or run with `go run` (see below).

## Development

You must provide the same environment variables as listed above. You can add them at the command line, to the shell environment, or in a `.env` file in the repo root.

Then run:

```bash
go run main.go
```

To include log output from Tailscale's `tsnet` library (may be noisy), add an environment variable `DEBUG` set to any non-empty value.

## Possible Future Features

- Support for HTTPS healthchecks, possibly including certificate validation
- Support for healthchecking non-HTTP services
