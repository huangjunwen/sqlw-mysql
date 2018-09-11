SHELL:=/bin/bash

# Generate wrapper code using default template.
gen: mysql_server_started
	@source env.sh && sqlw-mysql \
		-dsn "root:$$MYSQL_ROOT_PASSWORD@tcp(localhost:$$MYSQL_PORT)/$$MYSQL_DATABASE?parseTime=true" \
		-stmt stmts \
		-out models

# Generate png.
gen_png: gen_dot
	@dot -Tpng models/db.dot > models/db.png

# Generate dot file using graphviz template.
gen_dot: mysql_server_started
	@source env.sh && sqlw-mysql \
		-dsn "root:$$MYSQL_ROOT_PASSWORD@tcp(localhost:$$MYSQL_PORT)/$$MYSQL_DATABASE?parseTime=true" \
		-tmpl @graphviz \
		-out models

# Start a client.
mysql_client: mysql_server_started
	@source env.sh && docker exec -it "$$MYSQL_NAME" mysql -uroot -p$$MYSQL_ROOT_PASSWORD $$MYSQL_DATABASE

# Check whether the MySQL docker container is started.
mysql_server_started:
	@source env.sh && docker container inspect "$$MYSQL_NAME" > /dev/null 2>&1 || \
		{ echo >&2 "MySQL container \"$$MYSQL_NAME\" not started. Run 'make mysql_server' first"; exit 1; }

# Start a MySQL docker container and initialize it with schemas under initdb directory.
# NOTE: Use "ctrl+\" to quit.
mysql_server:
	@source env.sh && docker container inspect "$$MYSQL_NAME" > /dev/null 2>&1 || \
		docker run --rm --name "$$MYSQL_NAME" \
		-p 127.0.0.1:$$MYSQL_PORT:3306 \
		-v $(shell pwd)/initdb:/docker-entrypoint-initdb.d \
		-e MYSQL_ROOT_PASSWORD=$$MYSQL_ROOT_PASSWORD \
		-e MYSQL_DATABASE=$$MYSQL_DATABASE \
		mysql:$$MYSQL_VER
