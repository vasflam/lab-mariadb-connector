package mariadb

// See https://mariadb.com/kb/en/result-set-packets/#field-types
const MYSQL_TYPE_DECIMAL = 0
const MYSQL_TYPE_TINY = 1
const MYSQL_TYPE_SHORT = 2
const MYSQL_TYPE_LONG = 3
const MYSQL_TYPE_FLOAT = 4
const MYSQL_TYPE_DOUBLE = 5
const MYSQL_TYPE_NULL = 6
const MYSQL_TYPE_TIMESTAMP = 7
const MYSQL_TYPE_LONGLONG = 8
const MYSQL_TYPE_INT24 = 9
const MYSQL_TYPE_DATE = 10
const MYSQL_TYPE_TIME = 11
const MYSQL_TYPE_DATETIME = 12
const MYSQL_TYPE_YEAR = 13
const MYSQL_TYPE_NEWDATE = 14
const MYSQL_TYPE_VARCHAR = 15
const MYSQL_TYPE_BIT = 16
const MYSQL_TYPE_TIMESTAMP2 = 17
const MYSQL_TYPE_DATETIME2 = 18
const MYSQL_TYPE_TIME2 = 19
const MYSQL_TYPE_JSON = 245
const MYSQL_TYPE_NEWDECIMAL = 246
const MYSQL_TYPE_ENUM = 247
const MYSQL_TYPE_SET = 248
const MYSQL_TYPE_TINY_BLOB = 249
const MYSQL_TYPE_MEDIUM_BLOB = 250
const MYSQL_TYPE_LONG_BLOB = 251
const MYSQL_TYPE_BLOB = 252
const MYSQL_TYPE_VAR_STRING = 253
const MYSQL_TYPE_STRING = 254
const MYSQL_TYPE_GEOMETRY = 255

const FIELD_FLAG_NOT_NULL = 1
const FIELD_FLAG_PRIMARY_KEY = 2
const FIELD_FLAG_UNIQUE_KEY = 4
const FIELD_FLAG_MULTIPLE_KEY = 8
const FIELD_FLAG_BLOB = 16
const FIELD_FLAG_UNSIGNED = 32
const FIELD_FLAG_ZEROFILL_FLAG = 64
const FIELD_FLAG_BINARY_COLLATION = 128
const FIELD_FLAG_ENUM = 256
const FIELD_FLAG_AUTO_INCREMENT = 512
const FIELD_FLAG_TIMESTAMP = 1024
const FIELD_FLAG_SET = 2048
const FIELD_FLAG_NO_DEFAULT_VALUE_FLAG = 4096
const FIELD_FLAG_ON_UPDATE_NOW_FLAG = 8192
const FIELD_FLAG_NUM_FLAG = 32768


type tableColumn struct {
    name string
    fixedFields int
    charset uint16
    maxSize uint32
    kind uint8
    flag uint16
    decimals uint8
    unused uint16
}

type QueryResultRow map[string]interface{}
type QueryResultRows []QueryResultRow
