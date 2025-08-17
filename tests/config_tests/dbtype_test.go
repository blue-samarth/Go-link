package tests

import (
    "math/rand"
    "testing"
    "time"

    "github.com/blue-samarth/go-link/config"
)

// Ideally this should come from config.AllDbTypes to avoid drift.
var allDbTypes = []config.DbType{
    config.MongoDb,
    config.Postgres,
    config.MySQL,
    config.MsSQL,
    config.SQLite,
}

func TestDbType_IsValid(t *testing.T) {
    cases := []struct {
        name   string
        dbType config.DbType
        valid  bool
    }{
        {"MongoDB", config.MongoDb, true},
        {"Postgres", config.Postgres, true},
        {"MySQL", config.MySQL, true},
        {"MsSQL", config.MsSQL, true},
        {"SQLite", config.SQLite, true},

        {"empty string", config.DbType(""), false},
        {"unknown db", config.DbType("cassandra"), false},
        {"case sensitive", config.DbType("MONGODB"), false},
        {"with spaces", config.DbType(" mongodb "), false},
        {"partial match", config.DbType("mongo"), false},
        {"with unicode", config.DbType("🔥db"), false},
    }

    for _, tt := range cases {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.dbType.IsValid(); got != tt.valid {
                t.Errorf("IsValid(%q) = %v, want %v", tt.dbType, got, tt.valid)
            }
        })
    }
}

func TestDbType_AllConstantsCovered(t *testing.T) {
    for _, dbType := range allDbTypes {
        if !dbType.IsValid() {
            t.Errorf("defined constant %q should be valid", dbType)
        }
    }
}

func TestDbType_StringValues(t *testing.T) {
    expected := map[config.DbType]string{
        config.MongoDb:  "mongodb",
        config.Postgres: "postgres",
        config.MySQL:    "mysql",
        config.MsSQL:    "mssql",
        config.SQLite:   "sqlite",
    }

    for dbType, want := range expected {
        if got := string(dbType); got != want {
            t.Errorf("string(%q) = %q, want %q", dbType, got, want)
        }
    }
}

func TestDbType_ExhaustiveSwitch(t *testing.T) {
    for _, dbType := range allDbTypes {
        if !isKnownType(dbType) {
            t.Errorf("switch missing case for %q", dbType)
        }
    }
}

func isKnownType(d config.DbType) bool {
    switch d {
    case config.MongoDb, config.Postgres, config.MySQL, config.MsSQL, config.SQLite:
        return true
    default:
        return false
    }
}

// Property-style fuzz test: ensure random garbage never validates
func TestDbType_FuzzInvalid(t *testing.T) {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    for i := 0; i < 100; i++ {
        s := make([]rune, 8)
        for j := range s {
            s[j] = rune(r.Intn(0x80)) // random ASCII
        }
        db := config.DbType(string(s))
        if db.IsValid() {
            t.Errorf("random garbage %q should not be valid", db)
        }
    }
}

func BenchmarkDbType_IsValid_Valid(b *testing.B) {
    for i := 0; i < b.N; i++ {
        for _, dbType := range allDbTypes {
            _ = dbType.IsValid()
        }
    }
}

func BenchmarkDbType_IsValid_Invalid(b *testing.B) {
    invalids := []config.DbType{"foo", "bar", "123", "🔥db"}
    for i := 0; i < b.N; i++ {
        for _, dbType := range invalids {
            _ = dbType.IsValid()
        }
    }
}
