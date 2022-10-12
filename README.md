# Labaratory work - database conector

Simple maraidb connector. For education purposes only.

Features:
* Goroutine safe (threading safe) - queries are seved from channel.
* Supports only integer data type in column definition

## Installation

```
git clone https://github.com/vasflam/lab-mariadb-connector.git
cd lab-mariadb-connector
go mod tidy
cp .env.dist .env
# Update .env file according your database settings
```

Create database:
```
CREATE TABLE `numbers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `number` int(11) NOT NULL,
  PRIMARY KEY (`id`)
)
```

Run program:
```
go run main.go
```
