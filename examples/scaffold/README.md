## How to use this scaffold

- Need docker installed.
- Copy this directory.
- Modify `env.sh`.
- Add files:
  - Add database initialize sql files into `initdb` directory.
  - Optionally add statement xml files into `stmts` directory.
- `$ make start_mysql_server` and `$ make stop_mysql_server` to start/stop a mysql container.
- Daily usage:
  - `$ make mysql_client` or just `$ make` to start a mysql client.
  - `$ make reset_db` to drop the database and reload everything in `initdb`. Useful after modify sqls.
  - `$ make gen` to generate warpper code into `models` directory.
  - `$ make gen_png` to generate database diagram (.png) into `models` directory. Needs `dot` installed.
  - `$ make clean` remove generated files.
