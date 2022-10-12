# Labaratory work - database conector

Simple maraidb connector. For education purposes only.

### Features
* Goroutine safe (threading safe) - queries are served from channel.
* Supports only integer data type in column definition

### Information about mysql protocol
* https://dev.mysql.com/doc/internals/en/client-server-protocol.html
* https://mariadb.com/kb/en/clientserver-protocol/

# How to install

```
git clone https://github.com/vasflam/lab-mariadb-connector.git
cd lab-mariadb-connector
go mod tidy
cp .env.dist .env
# Update .env file according your database settings
```

Create table:
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

# Documentation
```
go install golang.org/x/tools/cmd/godoc@latest
~/go/bin/godoc -http=127.0.0.1:6060
```
Open [http://127.0.0.1:6060/pkg/github.com/vasflam/lab-mysql-connector/mariadb/](http://127.0.0.1:6060/pkg/github.com/vasflam/lab-mysql-connector/mariadb/)

### Example
```
package main

import (
  "log"
  "github.com/vasflam/lab-mysql-connector/mariadb"
)

func main() {
  config := mariadb.Config{
    Host: "127.0.0.1:3306",
    Username: "user",
    Password: "pass",
    Database: "test",
  }
  client, err := mariadb.Connection(config)
  if err != nil {
    log.Fatal(err)
  }
  
  rows, err := client.Query("SELECT 1")
  if err != nil {
    log.Fatal(err)
  }
  log.Printf("rows=%+v\n", rows)
}
```

