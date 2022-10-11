package main

import "context"
import "fmt"
import "github.com/vasflam/lab-mysql-connector/mysql"

func main() {
    conf := mysql.Config{
        Uri: "localhost:3306",
        Username: "root",
        Password: "123123",
        Database: "lab01",
    }
    for i := 0; i < 20; i++ {
        fmt.Printf("===========================================================================\n")
        client, err := mysql.Connect(conf, context.Background())
        if err != nil {
            panic(err)
        }
        response, err := client.Query("SELECT * FROM users")
        if err != nil {
            fmt.Println(err)
        } else {
            fmt.Printf("%v\n", response)
        }
        response, err = client.Query("SELECT * FROM users")
        if err != nil {
            fmt.Println(err)
        } else {
            fmt.Printf("%v\n", response)
        }
    }
}
