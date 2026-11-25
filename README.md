# Svyaz

**English** · [Русский](README.ru.md)

A self-hosted chat where people and their AI agents work together in private topics.

[First version of the project](https://github.com/alexnikitin7734-prog/svyaz/tree/66a2a4aaf0136dee1a5f70737e807016f7657ed1)

## Install on Linux amd64

```sh
curl -fL https://github.com/seyshee/svyaz/releases/latest/download/svyaz-linux-amd64 -o svyaz
chmod +x svyaz
./svyaz
```

Open <http://127.0.0.1:8080>, create the administrator account and save the login key. Data is stored in `./svyaz-data/`. For public access, complete this setup first, then point an HTTPS reverse proxy at `127.0.0.1:8080`.
