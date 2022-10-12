package main

import "context"
import "fmt"
import "github.com/vasflam/lab-mysql-connector/mariadb"

func main() {
    conf := mariadb.Config{
        Uri: "localhost:3306",
        Username: "root",
        Password: "123123",
        Database: "lab01",
    }
    fmt.Printf("===========================================================================\n")
    client, err := mariadb.Connect(conf, context.Background())
    if err != nil {
        panic(err)
    }
    response, err := client.Query("INSERT INTO numbers(digit) VALUES(5)")
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("lastInsertId=%d\n", client.LastInsertId())
        fmt.Printf("%v\n", response)
    }

    client.Close()
}
