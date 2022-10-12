package main

import (
    "testing"
    "github.com/vasflam/lab-mysql-connector/mariadb"
    "github.com/vasflam/lab-mysql-connector/mariadb/capabilities"
    "fmt"
)

var a = fmt.Errorf("")
var b mariadb.Connection
var c = capabilities.DEFAULT

func TestCapabilities(t *testing.T) {
    capsInit := uint64(120287306506)
    capsInit = uint64(51567829770)
    //caps := 28
    //caps := 28222218
    //caps := capsInit >> 20
    caps := capsInit 
    fmt.Printf("caps: [%d] %035b\n", caps, caps)
    if caps & capabilities.PLUGIN_AUTH != 0 {
        fmt.Println("PLUGIN_AUTH")
    }
    if caps & capabilities.IGNORE_SPACE != 0 {
        fmt.Println("IGNORE_SPACE")
    }
    if caps & capabilities.PROTOCOL_41 != 0 {
        fmt.Println("PROTOCOL_41")
    }
    if caps & capabilities.SECURE_CONNECTION != 0 {
        fmt.Println("SECURE_CONNECTION")
    }
    if caps & capabilities.PLUGIN_AUTH_LENENC_CLIENT_DATA != 0 {
        fmt.Println("PLUGIN_AUTH_LENENC_CLIENT_DATA")
    }
    if caps & capabilities.SESSION_TRACK != 0 {
        fmt.Println("SESSION_TRACK")
    }
    if caps & capabilities.TRANSACTIONS != 0 {
        fmt.Println("TRANSACTIONS")
    }
    if caps & capabilities.MULTI_RESULTS != 0 {
        fmt.Println("MULTI_RESULTS")
    }
    if caps & capabilities.PS_MULTI_RESULTS != 0 {
        fmt.Println("PS_MULTI_RESULTS")
    }
    if caps & capabilities.MARIADB_CLIENT_EXTENDED_TYPE_INFO != 0 {
        fmt.Println("MARIADB_CLIENT_EXTENDED_TYPE_INFO")
    }
    if caps & capabilities.CONNECT_WITH_DB != 0 {
        fmt.Println("CONNECT_WITH_DB")
    }
    if caps & capabilities.CONNECT_ATTRS != 0 {
        fmt.Println("CONNECT_ATTRS")
    }
    if caps & capabilities.MULTI_STATEMENTS != 0 {
        fmt.Println("CONNECT_ATTRS")
    }
    if caps & capabilities.DEPRECATE_EOF != 0 {
        fmt.Println("DEPRECATE_EOF")
    }
    if caps & capabilities.MYSQL != 0 {
        fmt.Println("MYSQL")
    }
    if caps & capabilities.FOUND_ROWS != 0 {
        fmt.Println("FOUND_ROWS")
    }
    if caps & capabilities.COMPRESS != 0 {
        fmt.Println("COMPRESS")
    }
    if caps & capabilities.LOCAL_FILES != 0 {
        fmt.Println("LOCAL_FILES")
    }
    if caps & capabilities.MARIADB_CLIENT_STMT_BULK_OPERATIONS != 0 {
        fmt.Println("MARIADB_CLIENT_STMT_BULK_OPERATIONS")
    }
    if caps & capabilities.RESERVED != 0 {
        fmt.Println("RESERVED")
    }
}

func TestChan(t *testing.T) {
    c1 := make(chan int)
    go func() {
        c1 <- 1
        close(c1)
        fmt.Printf("closed\n")
    }()

    for a := range c1 {
        fmt.Printf("a=%d\n", a)
    }
}
