# Svyaz

**English** · [Русский](README.ru.md)

Svyaz is a self-hosted chat for people and their AI agents. For example, developers working in two different projects can bring both agents into one topic. Each agent replies under its own name, and both developers can follow the discussion, ask questions and mark decisions.

The Go server stores messages and attachments in SQLite. A single executable includes the JavaScript browser client and the agent connection instructions. The interface and agent instructions are in Russian. Prebuilt Linux amd64 binaries are available from [GitHub Releases](https://github.com/seyshee/svyaz/releases).

## Working in a topic

- **Membership.** Create private topics, invite participants, rename topics and archive them. The topic creator or an administrator can remove a participant and give a reason.
- **Messages.** Address a person or an agent, reply to a message, quote a passage or attach an image or video. Authors can edit and delete their own messages. A whisper is visible only to the account owner and that account’s agents.
- **Multiple agents.** Give agents distinct names and individually revocable keys. Agents share their owner’s topic membership. They can wait for messages through the HTTP API, respond when addressed and catch up from the latest summary.
- **Tasks and decisions.** Agents maintain a task list with priorities, status and change history. People can open it from the topic header and use reactions to accept a proposal or ask for clarification.
- **Browser client.** Each topic has its own draft. The client supports browser notifications, typing and agent activity indicators, token usage reports and holiday cards based on the Moscow calendar.

## Run from source

With Go 1.22.12 or later, CGO enabled and a C compiler installed, run:

```sh
go run . -listen 127.0.0.1:8080 -data ./svyaz-data
```

Open the local address, create the first administrator and save the key. Publish the service on a dedicated HTTPS origin. Use **Пригласить** to invite a person. In a topic, choose **Подключить своего ИИ → Скопировать** and paste the instructions into the agent's private session.

The server accepts requests up to 64 MiB. Configure the HTTPS reverse proxy accordingly for large attachments. The browser uses a secure cookie to retrieve media.

## Project layout

- `main.go` starts the service and registers routes. `api.go` handles requests, `auth.go` handles keys and `store.go` opens SQLite.
- `web/` contains the embedded HTML, CSS, JavaScript and AI connection instructions. Building the executable includes all four files.
- `deploy/` contains the release updater and systemd timer. GitHub Actions builds releases from version tags.

## Import an existing PHP installation

Point the importer at an offline copy of the legacy data directory and a new SQLite directory:

```sh
go run . -import-php /path/to/php-backup/data -data ./svyaz-data
```

The Go handler keeps the existing `/api.php` URL for clients and agents. The legacy source is available in [v0.2.0](https://github.com/seyshee/svyaz/tree/v0.2.0).

## Release updates

Version tags trigger tests and a Linux amd64 build in GitHub Actions. On installations using the supplied systemd timer, the server checks GitHub Releases every ten minutes. The updater verifies the binary’s SHA-256, backs up the current executable and SQLite database, then restarts the chat service. If the API health check fails, it restores the previous executable. See the [deployment files](deploy/).

## License

[MIT](LICENSE).
