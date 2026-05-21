<h1 align="center" style="border-bottom: none">
  <div>
    WA Scheduler
  </div>
  WhatsApp Message Scheduler<br>
</h1>

<p align="center">
A simple message scheduling tool for WhatsApp private chats or groups. Built to make sure your messages are seen at the right time.
</p>

<p align="center">
  <em>This repository is also meant as a welcoming <strong>open source contribution exercise</strong>: a small, real codebase where you can practice filing issues, opening pull requests, and code review. See <a href="./CONTRIBUTING.md">CONTRIBUTING.md</a> for how we work together.</em>
</p>

## Why We Built This

In our group, important messages often vanished into the noise — sent at odd hours, buried under a flood of chats, and seen too late (or not at all). We built this tool to change that. With manual scheduling, you control exactly when your message hits.

## Architecture

![High Level Architecture](./docs/architecture.drawio.svg)

Available services:

- `Server and Dashboard Service` => Handling dashboard and API requests from clients. For API details see [REST API](./docs/rest_api.md).
- `WhatsApp Publisher` => External WhatsApp publisher service. This service sends messages to WhatsApp. The built-in adapter targets [go-whatsapp-web-multidevice](https://github.com/aldinokemal/go-whatsapp-web-multidevice). WA Scheduler does not manage publisher login, session, or deployment—you run that separately.
- `Storage` => Stores message state using SQLite in a database file. Schema reference for contributors: [docs/db/schema.sql](./docs/db/schema.sql).

## Features

- Schedule messages for private chats or groups
- Set exact send times
- Retry send

## Getting Started

### Prerequisites

- **Docker** and **Make**
- A **WhatsApp publisher** running outside this repo (logged in and reachable from Docker). See [go-whatsapp-web-multidevice](https://github.com/aldinokemal/go-whatsapp-web-multidevice) if you need one.

### Locally (Docker)

The compose file maps **`host.docker.internal`** to your host (via `extra_hosts` / `host-gateway`), which works on Docker Desktop for macOS and Windows and on modern Docker Engine on Linux.

1. Run and log in to your WhatsApp publisher on the host.
2. Clone and start WA Scheduler (adjust `WA_PUBLISHER_API_BASE_URL` to match your publisher’s host port—the local compose default assumes **`8474`**):

    ```bash
    git clone https://github.com/ghazlabs/wa-scheduler.git
    cd wa-scheduler

    WA_PUBLISHER_API_BASE_URL=http://host.docker.internal:8474 \
    WA_PUBLISHER_USERNAME=admin \
    WA_PUBLISHER_PASSWORD=admin \
      make run
    ```

3. Open <http://localhost:9866> for the dashboard.
4. Sign in with the dashboard credentials (defaults in local compose: username **`admin`**, password **`admin`** — override with `DASHBOARD_CLIENT_USERNAME` / `DASHBOARD_CLIENT_PASSWORD` if you changed them).
5. Schedule messages from the dashboard.
6. To resolve group recipient IDs, use your WhatsApp publisher’s tooling or API.

### Production

WA Scheduler is **self-hosted**; there is no hosted SaaS. Deploy by building [`build/package/Dockerfile`](./build/package/Dockerfile) or by adapting [`deploy/local/run/docker-compose.yml`](./deploy/local/run/docker-compose.yml): set the [environment variables](#environment-variables) (`DASHBOARD_CLIENT_*`, `WA_PUBLISHER_*`, `DB_PATH`, persistence volume, etc.), wire your publisher URL, and expose the listen port (default `9866`).

For broader deployment notes or questions, check [existing issues](https://github.com/ghazlabs/wa-scheduler/issues) or open a new one.

## Environment Variables

| Variable Name               | Required | Default | Description                                                                                                                                      |
| --------------------------- | -------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `LISTEN_PORT`               | Yes      | `9866`  | Port number the server listens on                                                                                                                |
| `DB_PATH`                   | Yes      | `/data/wa-scheduler.db` | SQLite database file path                                                                                                          |
| `DASHBOARD_CLIENT_USERNAME` | Yes      | –       | Username for dashboard authentication                                                                                                            |
| `DASHBOARD_CLIENT_PASSWORD` | Yes      | –       | Password for dashboard authentication                                                                                                            |
| `WA_DEFAULT_NUMBERS`        | No       | –       | Comma-separated list of default numbers could be private numbers or group id WhatsApp. E.g. `6287822334455@s.whatsapp.net,120363020892687898@g.us` |
| `WA_PUBLISHER_API_BASE_URL` | Yes      | –       | Base URL for WA Publisher API                                                                                                                    |
| `WA_PUBLISHER_USERNAME`     | Yes      | –       | Username for WA Publisher API                                                                                                                    |
| `WA_PUBLISHER_PASSWORD`     | Yes      | –       | Password for WA Publisher API                                                                                                                    |
| `WEB_CLIENT_PUBLIC_DIR`     | Yes      | `web`   | Directory for serving the web client                                                                                                             |

## Contributing

Thank you for your interest in contributing to WA Scheduler.

**Contributor guide:** **[CONTRIBUTING.md](./CONTRIBUTING.md)** (`@CONTRIBUTING.md`) — claiming issues, feature approvals vs bugs, `make test`, and Docker smoke checks.

There are many ways to contribute, and most of them don’t require writing code.

- [Spread the word](#spread-the-word)
- [Engage with the community](#engage-with-the-community)
- [Contribute code](#contribute-code)

### Spread the word

This might be the biggest help of all. Share WA Scheduler with your network or anyone who needs a simple way to schedule WhatsApp messages.

### Engage with the community

Every message, reaction, or bit of feedback counts. It keeps us motivated and reminds us that real people find this project useful.

### Contribute code

Code is just one piece of the puzzle—and contributing doesn’t always mean writing code. But if you do want to dive in, start small! Fix typos, report or squash bugs from the [issues page](https://github.com/ghazlabs/wa-scheduler/issues), polish up the docs, or add helpful features.

> [!TIP]
>
> Code matters, but it’s just one part of what makes a great product. Sometimes the easiest code fix isn’t the best choice overall. Don’t forget—there are plenty of other ways to contribute too!

#### Quick steps to contribute

1. Read **[CONTRIBUTING.md](./CONTRIBUTING.md)** (`@CONTRIBUTING.md`) for bug vs feature rules.
2. Fork the repo via the ["Fork"](https://github.com/ghazlabs/wa-scheduler/fork) button.
3. Clone your fork locally.
4. Create a branch:

    ```bash
    git checkout -b your-feature-name
    ```

5. Run **`make test`** before pushing.
6. Open a pull request.

## License

This project is licensed under the MIT License — see [LICENSE](./LICENSE).
