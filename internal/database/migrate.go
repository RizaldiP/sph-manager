package database

import (
	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Category{},
		&models.WorkItem{},
		&models.WorkSubItem{},
		&models.Template{},
		&models.TemplateItem{},
		&models.Customer{},
		&models.Vessel{},
		&models.Material{},
		&models.SphDocument{},
		&models.SphItem{},
		&models.SphSubItem{},
		&models.SphRevision{},
		&models.AuditLog{},
		&models.Setting{},
	)
	if err != nil {
		return err
	}
	return createPartialUniqueIndexes(db)
}

func createPartialUniqueIndexes(db *gorm.DB) error {
	stmts := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_categories_code_live ON categories(code) WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_work_items_code_live ON work_items(code) WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_customers_code_live ON customers(code) WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_vessels_customer_code_live ON vessels(customer_id, code) WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_materials_code_live ON materials(code) WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_templates_code_live ON templates(code) WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_sph_documents_number_rev_live ON sph_documents(document_number, revision) WHERE deleted_at IS NULL`,
		"CREATE UNIQUE INDEX IF NOT EXISTS uq_settings_key ON settings(\"key\")",
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}

func ExistingTables(db *gorm.DB) ([]string, error) {
	var tables []string
	err := db.Raw(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`).Scan(&tables).Error
	return tables, err
}

func ForeignKeyViolations(db *gorm.DB) ([]map[string]interface{}, error) {
	var rows []map[string]interface{}
	err := db.Raw(`PRAGMA foreign_key_check`).Find(&rows).Error
	return rows, err
}
