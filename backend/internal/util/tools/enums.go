package tools

import (
	"fmt"
	"strconv"
)

type PostgresqlExtension string

const (
	// needed for system monitoring (CPU, RAM)
	PostgresqlExtensionPgProctab PostgresqlExtension = "pg_proctab"
)

type PostgresqlVersion string

const (
	PostgresqlVersion13 PostgresqlVersion = "13"
	PostgresqlVersion14 PostgresqlVersion = "14"
	PostgresqlVersion15 PostgresqlVersion = "15"
	PostgresqlVersion16 PostgresqlVersion = "16"
	PostgresqlVersion17 PostgresqlVersion = "17"
)

type PostgresqlExecutable string

const (
	PostgresqlExecutablePgDump PostgresqlExecutable = "pg_dump"
	PostgresqlExecutablePsql   PostgresqlExecutable = "psql"
)

func GetPostgresqlVersionEnum(version string) PostgresqlVersion {
	switch version {
	case "13":
		return PostgresqlVersion13
	case "14":
		return PostgresqlVersion14
	case "15":
		return PostgresqlVersion15
	case "16":
		return PostgresqlVersion16
	case "17":
		return PostgresqlVersion17
	default:
		panic(fmt.Sprintf("invalid postgresql version: %s", version))
	}
}

func IsBackupDbVersionHigherThanRestoreDbVersion(
	backupDbVersion, restoreDbVersion PostgresqlVersion,
) bool {
	backupDbVersionInt, err := strconv.Atoi(string(backupDbVersion))
	if err != nil {
		return false
	}

	restoreDbVersionInt, err := strconv.Atoi(string(restoreDbVersion))
	if err != nil {
		return false
	}

	return backupDbVersionInt > restoreDbVersionInt
}
