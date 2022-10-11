package main

import "github.com/vasflam/lab-mysql-connector/mysql"

func main() {
    conf := mysql.Config{
        Uri: "localhost:3306",
        Username: "root",
        Password: "123123",
        Database: "test",
    }
    _, err := mysql.Connect(conf)
    if err != nil {
        panic(err)
    }
}
