# Default vars.
DDL_DIR:=ddl
DML_DIR:=dml
OUTPUT_DIR=models
MYSQL_NAME:=scaffold
MYSQL_PORT:=13306
MYSQL_ROOT_PASSWORD:=654321
MYSQL_DATABASE:=scaffold
MYSQL_VER:=5.7.21

# You can override them in vars.mk.
-include vars.mk

# Converts to absolute paths.
DDL_DIR:=$(abspath $(DDL_DIR))
DML_DIR:=$(abspath $(DML_DIR))
OUTPUT_DIR:=$(abspath $(OUTPUT_DIR))

# Starts a mysql client.
mysql_client: mysql_server_started
	@docker exec -it $(MYSQL_NAME) mysql -uroot -p$(MYSQL_ROOT_PASSWORD) $(MYSQL_DATABASE)
	
# Generates wrapper code using default template.
gen: mysql_server_started
	@sqlw-mysql \
		-dsn "root:$(MYSQL_ROOT_PASSWORD)@tcp(localhost:$(MYSQL_PORT))/$(MYSQL_DATABASE)?parseTime=true" \
		-stmt $(DML_DIR) \
		-out $(OUTPUT_DIR) && \
		gofmt -w $(OUTPUT_DIR)/*.go

# Generates png.
gen_png: gen_dot
	@dot -Tpng $(OUTPUT_DIR)/db.dot > $(OUTPUT_DIR)/db.png

# Generates dot file using graphviz template.
gen_dot: mysql_server_started
	@sqlw-mysql \
		-dsn "root:$(MYSQL_ROOT_PASSWORD)@tcp(localhost:$(MYSQL_PORT))/$(MYSQL_DATABASE)?parseTime=true" \
		-tmpl @graphviz \
		-out $(OUTPUT_DIR)

# Resets the database: drop all tables and re init them.
# NOTE: files in $(DDL_DIR) are cat in alphanumeric order
# See
# 	https://superuser.com/q/192280
reset_db: mysql_server_started
	@docker exec $(MYSQL_NAME) mysql -uroot -p$(MYSQL_ROOT_PASSWORD) $(MYSQL_DATABASE) \
		-e "DROP DATABASE $(MYSQL_DATABASE); CREATE DATABASE $(MYSQL_DATABASE)" && \
		cat $(DDL_DIR)/* | docker exec -i $(MYSQL_NAME) mysql -uroot -p$(MYSQL_ROOT_PASSWORD) $(MYSQL_DATABASE)

# Checks mysql container started and healthy.
mysql_server_started:
	@[[ $$(docker container inspect --format "{{json .State.Health.Status }}" $(MYSQL_NAME) 2>/dev/null) == *healthy* ]] || \
		{ echo >&2 "MySQL container \"$(MYSQL_NAME)\" not started or starting."; exit 1; }

# Starts mysql container in background.
# Also see 
# 	https://stackoverflow.com/a/46625302
# 	https://stackoverflow.com/a/50529359
start_mysql_server:
	@docker container inspect $(MYSQL_NAME) > /dev/null 2>&1 || \
		docker run -d --name $(MYSQL_NAME) \
		--health-cmd='mysqladmin ping --silent' \
		-v $(DDL_DIR):/docker-entrypoint-initdb.d \
		-p 127.0.0.1:$(MYSQL_PORT):3306 \
		-e MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) \
		-e MYSQL_DATABASE=$(MYSQL_DATABASE) \
		mysql:$(MYSQL_VER) --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci

# Stops mysql container.
stop_mysql_server:
	@docker rm -f $(MYSQL_NAME) 2>/dev/null

# Cleans generated files.
clean:
	@rm $(OUTPUT_DIR)/*
