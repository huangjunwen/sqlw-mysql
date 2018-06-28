
gen:
	sqlw-mysql -dsn "root:123456@tcp(localhost:16033)/dev?parseTime=true" -stmt ./stmts

# use Ctrl+\ to stop it.
mysqld:
	[ "$(shell docker ps -a --format '{{.Names}}' | grep sqlw-mysql-quickstart)" ] || \
		docker run --rm --name sqlw-mysql-quickstart -p 127.0.0.1:16033:3306 \
		-v $(shell pwd)/sqls:/docker-entrypoint-initdb.d \
		-e MYSQL_ROOT_PASSWORD=123456 \
		-e MYSQL_DATABASE=dev \
		mysql:5.7.21

clean:
	rm models/*
