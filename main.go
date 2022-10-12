package main

import (
    "context"
    "log"
    "fmt"
    "os"
    "github.com/joho/godotenv"
    "github.com/vasflam/lab-mysql-connector/mariadb"
)

var _ = fmt.Errorf("")

func printRows(rows mariadb.QueryResultRows) {
    log.Println("selected rows")
    for _, row := range rows {
        log.Printf("id=%d, number=%d\n", row["id"], row["number"])
    }
}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal(err)
    }

    conf := mariadb.Config{
        Uri: os.Getenv("DB_HOST"),
        Username: os.Getenv("DB_USERNAME"),
        Password: os.Getenv("DB_PASSWORD"),
        Database: os.Getenv("DB_NAME"),
    }
    client, err := mariadb.Connect(conf, context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // CREATE 10 RECORDS
    log.Println("insert rows")
    for i := 0; i < 11; i++ {
        query := fmt.Sprintf("INSERT INTO numbers(number) VALUES(%d)", i)
        _, err = client.Query(query)
        if err != nil {
            log.Fatal(err)
        }
        lastInsertId := client.LastInsertId()
        affectedRows := client.AffectedRows()
        log.Printf("row %d was inserted. id=%d, affectedRows=%d\n", i, lastInsertId, affectedRows)
    }

    // READ RECORD
    rows, err := client.Query("SELECT * FROM numbers")
    if err != nil {
        log.Fatal(err)
    }
    printRows(rows)

    // UPDATE ROW
    _, err = client.Query("UPDATE numbers SET number=100 WHERE id=1")
    if err != nil {
        log.Fatal(err)
    }

    // DELETE ROWS
    _, err = client.Query("DELETE FROM numbers WHERE id > 1")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("%d row(s) were deleted\n", client.AffectedRows())

    rows, err = client.Query("SELECT * FROM numbers")
    if err != nil {
        log.Fatal(err)
    }
    printRows(rows)

    client.Close()
}
