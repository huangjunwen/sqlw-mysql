## Quick start

- Need docker installed.
- `$ make mysql_server` to start a mysql container.
- Makes:
  - `$ make` to generate warpper code into `models` directory.
  - `$ make gen_png` to generate database diagram (.png) into `models` directory.
  - `$ make mysql_client` to start a mysql client.
- `go run ./main.go` to run some test code.
