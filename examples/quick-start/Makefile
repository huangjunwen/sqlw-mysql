SHELL:=/bin/bash

# Starts a mysql client.
mysql_client: mysql_server_started
	@source env.sh && docker exec -it "$$MYSQL_NAME" mysql -uroot -p$$MYSQL_ROOT_PASSWORD $$MYSQL_DATABASE
	
# Generates wrapper code using default template.
gen: mysql_server_started
	@source env.sh && sqlw-mysql \
		-dsn "root:$$MYSQL_ROOT_PASSWORD@tcp(localhost:$$MYSQL_PORT)/$$MYSQL_DATABASE?parseTime=true" \
		-stmt stmts \
		-out models && \
		gofmt -w models/*.go

# Generates png.
gen_png: gen_dot
	@dot -Tpng models/db.dot > models/db.png

# Generates dot file using graphviz template.
gen_dot: mysql_server_started
	@source env.sh && sqlw-mysql \
		-dsn "root:$$MYSQL_ROOT_PASSWORD@tcp(localhost:$$MYSQL_PORT)/$$MYSQL_DATABASE?parseTime=true" \
		-tmpl @graphviz \
		-out models

# Resets the database: drop all tables and re init them.
reset_db: mysql_server_started
	@source env.sh && \
		docker exec "$$MYSQL_NAME" mysql -uroot -p$$MYSQL_ROOT_PASSWORD $$MYSQL_DATABASE -e "DROP DATABASE $$MYSQL_DATABASE; CREATE DATABASE $$MYSQL_DATABASE" && \
		cat $(shell pwd)/initdb/*.sql | docker exec -i "$$MYSQL_NAME" mysql -uroot -p$$MYSQL_ROOT_PASSWORD $$MYSQL_DATABASE

# Checks mysql container started and healthy.
mysql_server_started:
	@source env.sh && \
		[[ $$(docker container inspect --format "{{json .State.Health.Status }}" $$MYSQL_NAME 2>/dev/null) == *healthy* ]] || \
		{ echo >&2 "MySQL container \"$$MYSQL_NAME\" not started or starting."; exit 1; }

# Starts mysql container in background.
# Also see https://stackoverflow.com/a/46625302/157235
start_mysql_server:
	@source env.sh && docker container inspect "$$MYSQL_NAME" > /dev/null 2>&1 || \
		docker run -d --name "$$MYSQL_NAME" \
		--health-cmd='mysqladmin ping --silent' \
		-v $(shell pwd)/initdb:/docker-entrypoint-initdb.d \
		-p 127.0.0.1:$$MYSQL_PORT:3306 \
		-e MYSQL_ROOT_PASSWORD=$$MYSQL_ROOT_PASSWORD \
		-e MYSQL_DATABASE=$$MYSQL_DATABASE \
		mysql:$$MYSQL_VER

# Stops mysql container.
stop_mysql_server:
	@source env.sh && docker rm -f "$$MYSQL_NAME" 2>/dev/null

# Cleans generated files.
clean:
	@rm models/*
