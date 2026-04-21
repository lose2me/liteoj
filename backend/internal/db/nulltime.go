package db

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// NullTime is a nullable time value that implements sql.Scanner to tolerate
// the zoo of representations a SQLite driver can hand back.
//
// Why this type exists:
//   - Pure-Go drivers (modernc.org/sqlite via glebarez/sqlite) return DATETIME
//     columns as Go strings (ISO-8601 text) at the driver layer.
//   - GORM normally bridges string → time.Time by consulting the target model's
//     schema. But GORM only knows the target type when the column maps to a
//     declared struct field of a known model; aggregate aliases like
//     `MIN(CASE WHEN ... THEN created_at END) AS first_ac` lose that type hint
//     and fall back to database/sql's convertAssign — which does NOT know how
//     to stuff a string into a *time.Time and errors out with:
//         "unsupported Scan, storing driver.Value type string into type *time.Time"
//   - stdlib's sql.NullTime only accepts time.Time / nil and has the same hole.
//
// Any handler that runs Raw(...).Scan(&rows) where rows has a datetime-ish
// aggregate column should use db.NullTime for those fields.
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Layouts we'll try, ordered by how often we expect to see them. The first few
// cover glebarez/sqlite's default serialization of time.Time inserted by GORM.
var nullTimeLayouts = []string{
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02 15:04:05-07:00",
	"2006-01-02 15:04:05Z07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
	time.RFC3339Nano,
	time.RFC3339,
}

// Scan implements sql.Scanner.
func (n *NullTime) Scan(value any) error {
	if value == nil {
		n.Time = time.Time{}
		n.Valid = false
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		n.Time = v
		n.Valid = !v.IsZero()
		return nil
	case string:
		if v == "" {
			n.Time = time.Time{}
			n.Valid = false
			return nil
		}
		t, err := parseNullTime(v)
		if err != nil {
			return err
		}
		n.Time = t
		n.Valid = true
		return nil
	case []byte:
		if len(v) == 0 {
			n.Time = time.Time{}
			n.Valid = false
			return nil
		}
		t, err := parseNullTime(string(v))
		if err != nil {
			return err
		}
		n.Time = t
		n.Valid = true
		return nil
	}
	return fmt.Errorf("db.NullTime: unsupported scan type %T", value)
}

// Value implements driver.Valuer so NullTime can also be used in WHERE / binds.
func (n NullTime) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time, nil
}

func parseNullTime(s string) (time.Time, error) {
	for _, layout := range nullTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("db.NullTime: cannot parse %q", s)
}
