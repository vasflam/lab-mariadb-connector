// Labaratory work - database conector
//
// Simple maraidb connector. For education purposes only.
//
// Features
//   - Goroutine safe (threading safe) - queries are served from channel.
//   - Supports only integer data type in column definition
//
// Information about mysql protocol
//   - https://dev.mysql.com/doc/internals/en/client-server-protocol.html
//   - https://mariadb.com/kb/en/clientserver-protocol/
package mariadb
