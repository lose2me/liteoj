package db

import (
	"log"

	"gorm.io/gorm"

	"github.com/liteoj/liteoj/backend/internal/models"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.User{},
		&models.TagGroup{},
		&models.Tag{},
		&models.Problem{},
		&models.Testcase{},
		&models.ProblemSet{},
		&models.ProblemSetItem{},
		&models.ProblemSetMember{},
		&models.ProblemSetBan{},
		&models.Submission{},
		&models.AITask{},
		&models.HomePage{},
	); err != nil {
		return err
	}
	if err := mergeLegacyIOFormatColumns(db); err != nil {
		return err
	}
	return dropLegacyTestcaseColumns(db)
}

// dropLegacyTestcaseColumns removes the now-unused testcases.{score,hidden}
// columns. Runs at most once per deployment: once the columns are gone,
// HasColumn returns false and the function no-ops.
func dropLegacyTestcaseColumns(db *gorm.DB) error {
	m := db.Migrator()
	for _, col := range []string{"score", "hidden"} {
		if m.HasColumn(&models.Testcase{}, col) {
			if err := m.DropColumn(&models.Testcase{}, col); err != nil {
				return err
			}
			log.Printf("migrate: dropped legacy testcases.%s column", col)
		}
	}
	return nil
}

// mergeLegacyIOFormatColumns is a one-shot migration that folds the now-removed
// problems.{input_format,output_format,sample_input,sample_output} columns
// into the description, then drops them. Runs at most once per deployment:
// once the columns are gone HasColumn returns false and the function no-ops.
func mergeLegacyIOFormatColumns(db *gorm.DB) error {
	m := db.Migrator()
	hasIn := m.HasColumn(&models.Problem{}, "input_format")
	hasOut := m.HasColumn(&models.Problem{}, "output_format")
	hasSampleIn := m.HasColumn(&models.Problem{}, "sample_input")
	hasSampleOut := m.HasColumn(&models.Problem{}, "sample_output")
	if !hasIn && !hasOut && !hasSampleIn && !hasSampleOut {
		return nil
	}
	log.Printf("migrate: folding legacy input/output/sample columns into description")
	// Build the concatenation in raw SQL — portable across SQLite and Postgres
	// (both use || for string concat and CHAR(10) for LF).
	section := func(col, header string) string {
		return `CASE WHEN COALESCE(` + col + `, '') != ''` +
			` THEN CHAR(10)||CHAR(10)||'## ` + header + `'||CHAR(10)||CHAR(10)||` + col +
			` ELSE '' END`
	}
	fenced := func(col, header string) string {
		// wrap the column value in a Markdown code fence so multi-line samples
		// render verbatim. triple backticks expressed as three CHAR(96) calls.
		fence := "CHAR(96)||CHAR(96)||CHAR(96)"
		return `CASE WHEN COALESCE(` + col + `, '') != ''` +
			` THEN CHAR(10)||CHAR(10)||'### ` + header + `'||CHAR(10)||CHAR(10)||` +
			fence + `||CHAR(10)||` + col + `||CHAR(10)||` + fence +
			` ELSE '' END`
	}
	var parts []string
	if hasIn {
		parts = append(parts, section("input_format", "输入格式"))
	}
	if hasOut {
		parts = append(parts, section("output_format", "输出格式"))
	}
	if hasSampleIn || hasSampleOut {
		// Only add the 样例 header if at least one sample exists.
		any := "COALESCE(sample_input, '') != '' OR COALESCE(sample_output, '') != ''"
		if !hasSampleIn {
			any = "COALESCE(sample_output, '') != ''"
		} else if !hasSampleOut {
			any = "COALESCE(sample_input, '') != ''"
		}
		var inner []string
		if hasSampleIn {
			inner = append(inner, fenced("sample_input", "样例输入"))
		}
		if hasSampleOut {
			inner = append(inner, fenced("sample_output", "样例输出"))
		}
		sample := `CASE WHEN ` + any +
			` THEN CHAR(10)||CHAR(10)||'## 样例'||` + joinExpr(inner, "||") +
			` ELSE '' END`
		parts = append(parts, sample)
	}
	expr := "description"
	for _, p := range parts {
		expr = expr + " || " + p
	}
	if err := db.Exec(`UPDATE problems SET description = ` + expr).Error; err != nil {
		return err
	}
	for _, col := range []string{"input_format", "output_format", "sample_input", "sample_output"} {
		if m.HasColumn(&models.Problem{}, col) {
			if err := m.DropColumn(&models.Problem{}, col); err != nil {
				return err
			}
		}
	}
	log.Printf("migrate: legacy IO/sample columns merged and dropped")
	return nil
}

// joinExpr is a small helper to stitch SQL expression fragments with a
// separator — avoids importing strings just for this.
func joinExpr(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += sep + p
	}
	return out
}
