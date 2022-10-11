package capabilities

const MYSQL = 1
/* Found instead of affected rows */
const FOUND_ROWS = 2
/* get all column flags */
const LONG_FLAG = 4
/* one can specify db on connect */
const CONNECT_WITH_DB = 8
/* don't allow database.table.column */
const NO_SCHEMA = 1 << 4;
/* can use compression protocol */
const COMPRESS = 1 << 5;
/* odbc client */
const ODBC = 1 << 6;
/* can use LOAD DATA LOCAL */
const LOCAL_FILES = 1 << 7;
/* ignore spaces before '' */
const IGNORE_SPACE = 1 << 8;
/* new 4.1 protocol */
const PROTOCOL_41 = 1 << 9;
/* this is an interactive client */
const INTERACTIVE = 1 << 10;
/* switch to ssl after handshake */
const SSL = 1 << 11;
/* IGNORE sigpipes */
const IGNORE_SIGPIPE = 1 << 12;
/* client knows about transactions */
const TRANSACTIONS = 1 << 13;
/* old flag for 4.1 protocol  */
const RESERVED = 1 << 14;
/* new 4.1 authentication */
const SECURE_CONNECTION = 1 << 15;
/* enable/disable multi-stmt support */
const MULTI_STATEMENTS = 1 << 16;
/* enable/disable multi-results */
const MULTI_RESULTS = 1 << 17;
/* multi-results in ps-protocol */
const PS_MULTI_RESULTS = 1 << 18;
/* client supports plugin authentication */
const PLUGIN_AUTH = 1 << 19;
/* permits connection attributes */
const CONNECT_ATTRS = 1 << 20;
/* Enable authentication response packet to be larger than 255 bytes. */
const PLUGIN_AUTH_LENENC_CLIENT_DATA = 1 << 21;
/* Don't close the connection for a connection with expired password. */
const CAN_HANDLE_EXPIRED_PASSWORDS = 1 << 22;
/* Capable of handling server state change information. Its a hint to the
  server to include the state change information in Ok packet. */
const SESSION_TRACK = 1 << 23;
/* Client no longer needs EOF packet */
const DEPRECATE_EOF = 1 << 24;
const SSL_VERIFY_SERVER_CERT = 1 << 30;

/* MariaDB extended capabilities */

/* Permit bulk insert*/
const MARIADB_CLIENT_STMT_BULK_OPERATIONS = 1 << 34;

/* Clients supporting extended metadata */
const MARIADB_CLIENT_EXTENDED_TYPE_INFO = 1 << 35;
const MARIADB_CLIENT_CACHE_METADATA = 1 << 36;

var DEFAULT uint64 =
    FOUND_ROWS |
    IGNORE_SPACE |
    PROTOCOL_41 |
    TRANSACTIONS |
    SECURE_CONNECTION |
    MULTI_RESULTS |
    PS_MULTI_RESULTS |
    PLUGIN_AUTH_LENENC_CLIENT_DATA |
    SESSION_TRACK |
    DEPRECATE_EOF



