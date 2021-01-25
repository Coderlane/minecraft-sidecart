# Minecraft Sidecart

![CI](https://github.com/Coderlane/minecraft-sidecart/workflows/CI/badge.svg) [![codecov](https://codecov.io/gh/Coderlane/minecraft-sidecart/branch/master/graph/badge.svg?token=8G3GBG1CAY)](https://codecov.io/gh/Coderlane/minecraft-sidecart)

`minecraft-sidecart` handles uploading Minecraft server metadata and eventually responding to remote requests.

## Usage

### Authenticate

Use `minecraft-sidecart auth signin` to authenticate. It will output a URL for
you to copy and paste into your browser. Once authenticated, copy the code from
your browser back in to `minecraft-sidecart`.

### Daemon

Launch the daemon with `minecraft-sidecart daemon`. The daemon will detect
changes on the Minecraft server and upload them as they occur.

### Server

Use `minecraft-sidecart server add` to add a server for the daemon to watch.
Simply provide a `name` and a `path` to the root of the server directory. For
example:

```
./minecraft-sidecart server add \
  --name "Main Server" --path /opt/minecraft/server
```
