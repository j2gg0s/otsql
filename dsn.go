package otsql

import "strings"

func parseDSN(dsn string) (string, string) {
	addr, database := "localhost", ""

	if i := strings.Index(dsn, "://"); i > -1 {
		dsn = dsn[i+3:]
	}

	if i := strings.Index(dsn, "/"); i > -1 {
		if i > 0 {
			if j := strings.Index(dsn[:i], "@"); j > 0 && j+1 < i {
				addr = dsn[j+1 : i]
			}
		}
		if i < len(dsn) {
			j := strings.Index(dsn[i+1:], "?")
			if j > -1 {
				j = i + 1 + j
			} else {
				j = len(dsn)
			}
			database = dsn[i+1 : j]
		}
		return addr, database
	}

	host, port := "", ""
	for _, s := range strings.Split(dsn, " ") {
		if strings.HasPrefix(s, "dbname=") {
			database = s[len("dbname="):]
		} else if strings.HasPrefix(s, "host=") {
			host = s[len("host="):]
		} else if strings.HasPrefix(s, "port=") {
			port = s[len("port="):]
		}
	}

	addr = host
	if port != "" {
		addr += ":" + port
	}
	return addr, database
}
