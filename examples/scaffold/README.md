## How to use this scaffold

0. Need docker installed.
1. Copy this directory.
2. Modify `env.sh`.
3. 
  - Put database initialize sql files into `initdb` directory.
  - Optionally put statement xml files into `stmts` directory.
4. `$ make mysql_server` to start a mysql container.
5. 
  - `$ make` to generate warpper code into `models` directory.
  - `$ make gen_png` to generate database diagram (.png) into `models` directory.
  - `$ make mysql_client` to start a mysql client.
