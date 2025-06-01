# PTT (Penetration Testing Toolkit)

PTT is a collaborative and easy-to-use penetration testing toolkit featuring
plugin-based modules. It provides a web interface for users to interact with
pentest tools.

## Status: In Development

## Features

- Plugin system.
  - Dynamic plugin loading.
  - HTTP request proxying.
  - SSE support.
  - Example plugin.
- Web interface.
  - No need to memorize CLI pentest tools!
- Self-hostable binary.
  - Only a single binary file needed to get started!

## TODO

- Project contexts.

## How PTT Works

PTT runs as a single binary file and loads plugins from any `./plugins`
directory in the current environment (if this directory does not exist, PTT will
create one). Plugin executables must be placed into this directory and contain
`.plugin` in its filename (ex. `example_plugin.plugin`). PTT will then try to
dynamically launch and register the plugins it finds. Note that **the user does
not launch plugins directly**, they will be launched and managed by PTT.

## Plugin System

PTT uses [Hashicorp's go-plugin](https://github.com/hashicorp/go-plugin) library
to interact with plugins over a local gRPC connection. When a request is made to
the plugin's URL, PTT will proxy this request (over gRPC) to the associated
plugin. The plugin must then respond (back to PTT over gRPC) with the HTML that
will be rendered and sent to the frontend HTML body.

## Plugin Development

- Go is recommended, but any language that supports gRPC should work (not yet
  tested).
- When using Go, import `"github.com/Penetration-Testing-Toolkit/ptt/shared"`
- See the `example_plugin` directory for an example of a module plugin.
  - Note: the only internal package used in the example is the `templates`
    package. This is to simplify PTT's build process. When creating your own
    plugin, you must create your own templating and build process.
- See the `shared` package for the `Module` interface that a module plugin must
  implement, and other plugin documentation.
- See [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin) for more
  information on the plugin system.
- The frontend comes with [Alpine.js](https://alpinejs.dev/) and
  [Alpine AJAX](https://alpine-ajax.js.org/) preloaded.

## Running PTT

To run PTT, build (`make build`) then run the executable in a terminal
(`./build/ptt`) and access the web interface hosted on port 8080
(http://localhost:8080). Alternatively, run `make dev` for a live reloading
environment. See [Options](#options) for information on launch options. Make
sure plugins are placed in the `./plugins` directory and PTT automatically
discover and attempt to run them.

## Default User Account

- Username: `root`
- Password: `changeme!!`

## Options

- Use Environment Variables before running PTT to use these options.
- `JSON=FALSE`
  - Disables logging as JSON.
- `LOG_LEVEL=2`
  - Sets logging level (1-5).
  - See: `hclog.Level`.
- `ENV=DEV` (EXPERIMENTAL)
  - Switches to dev environment.
    - Hosts static assets from the file system instead of the binary's embedded
      file system.

## Development Requirements

- [Go](https://go.dev/) >= 1.24
- [Node.JS](https://nodejs.org/en) (LTS)
  - JS runtime for pnpm & tailwindcss.
- [pnpm](https://pnpm.io/)
  - JS package manager.
- [tailwindcss](https://tailwindcss.com/) CLI >= 4.0
  - CSS framework.
  - Requires Node.JS.
  - Install Tailwind CLI via pnpm.
- [templ](https://github.com/a-h/templ)
  - Template engine for Go.
- [air](https://github.com/air-verse/air)
  - Code live reloading.
- make
  - Build automation.
- [gRPC](https://grpc.io/) &
  [protoc-compiler](https://protobuf.dev/installation/)
  - rpc communication handler.
  - Used for communication between PTT & plugins.

> **Note**  
> The packages above can be installed using `install-deps.sh`.

**Important:** After running `install-deps.sh`, package paths won't be updated in the current terminal session.  
To apply changes, either:

- Start a **new terminal window**, or
- Run the following command manually:

```bash
source ~/.bashrc
```


## Development Commands

- Build project:

```bash
make build
```

- Run development (code live reloading) environment:

```bash
make dev
```

- Delete temporary directories:

```bash
make clean
```

## Docker
- Build and run Docker image
```bash
sudo docker-compose up --build
```
- Subsequent launches can be run with:
```bash
sudo docker-compose up
```

For persistent database data, create `db.sqlite` before launching.
Copy desired plugins to `plugins` directory before launching.
