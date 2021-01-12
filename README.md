# Minecraft Sidecart

![CI](https://github.com/Coderlane/minecraft-sidecart/workflows/CI/badge.svg) [![codecov](https://codecov.io/gh/Coderlane/minecraft-sidecart/branch/master/graph/badge.svg?token=8G3GBG1CAY)](https://codecov.io/gh/Coderlane/minecraft-sidecart)

`minecraft-sidecart` handles uploading Minecraft server metadata and eventually responding to remote requests.

## Usage

Using minecraft-sidecart is simple for now, simply point it at the root of your Minecraft server.

```
Usage of ./minecraft-sidecart:
  -server string
    	Path to the root of the server. (default "./")
```

For example: `./minecraft-sidecart --server /opt/minecraft/server` 

The first run will require authenticating. You'll need to open the URL that `minecraft-sidecart` outputs in your browsers, authenticate, and copy the code back in to `minecraft-sidecart`. After that, `minecraft-sidecart` will automatically detect changes and upload them as necessary.
