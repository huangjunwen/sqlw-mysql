## How to use this scaffold

- Need docker installed.
- Copy this directory.
- Modify `env.sh`.
- Add files:
  - Add database initialize sql files into `initdb` directory.
  - Optionally add statement xml files into `stmts` directory.
- `$ make mysql_server` to start a mysql container.
- Makes:
  - `$ make` to generate warpper code into `models` directory.
  - `$ make gen_png` to generate database diagram (.png) into `models` directory.
  - `$ make mysql_client` to start a mysql client.
