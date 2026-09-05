package sharebackup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// SchemaVersion paket backup share.
const PackageSchemaVersion = "1.0"

// SectionKey kunci seksi yang boleh dipilih penerima saat restore.
type SectionKey string

const (
	// SectionCategories seksi Kategori (induk master pekerjaan).
	SectionCategories SectionKey = "categories"
	// SectionWorkItems seksi Master Pekerjaan (pekerjaan + sub-pekerjaan).
	SectionWorkItems SectionKey = "workItems"
	// SectionTemplates seksi Template.
	SectionTemplates SectionKey = "templates"
	// SectionCustomers seksi Customer (+ kapal).
	SectionCustomers SectionKey = "customers"
	// SectionMaterials seksi Material.
	SectionMaterials SectionKey = "materials"
	// SectionSph seksi Semua Data SPH.
	SectionSph SectionKey = "sph"
)

// AllSectionKeys daftar seksi dalam urutan tampilan.
func AllSectionKeys() []SectionKey {
	return []SectionKey{
		SectionSph,
		SectionWorkItems,
		SectionCategories,
		SectionTemplates,
		SectionCustomers,
		SectionMaterials,
	}
}

// ShareBackupPackage adalah isi file backup yang dapat dibagikan (".sphbak").
// Seluruh relasi memakai natural key berbasis NAMA (case-insensitive) karena
// kode dihasilkan per mesin dan bisa berbeda antara pengirim & penerima.
type ShareBackupPackage struct {
	SchemaVersion string                  `json:"schemaVersion"`
	Metadata      ShareBackupMetadata     `json:"metadata"`
	Categories    []PackageCategory       `json:"categories,omitempty"`
	WorkItems     []PackageWorkItem       `json:"workItems,omitempty"`
	WorkSubItems  []PackageWorkSubItem    `json:"workSubItems,omitempty"`
	Templates     []PackageTemplate       `json:"templates,omitempty"`
	Customers     []PackageCustomer       `json:"customers,omitempty"`
	Materials     []PackageMaterial       `json:"materials,omitempty"`
	SphDocuments  []PackageSphDocument    `json:"sphDocuments,omitempty"`
	Checksum      string                  `json:"checksum"`
}

// ShareBackupMetadata identitas sumber backup.
type ShareBackupMetadata struct {
	PackageID  string    `json:"packageId"`
	DeviceName string    `json:"deviceName"`
	CreatedAt  time.Time `json:"createdAt"`
}

// PackageCategory — Kategori; natural key dedup = Name (ci).
type PackageCategory struct {
	Code        string `json:"code,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Sequence    int    `json:"sequence"`
	IsActive    bool   `json:"isActive"`
}

// PackageWorkItem — WorkItem; natural key dedup = (Nama Kategori, Nama) (ci).
type PackageWorkItem struct {
	Code                 string `json:"code,omitempty"`
	Name                 string `json:"name"`
	Description          string `json:"description,omitempty"`
	DefaultUnit          string `json:"defaultUnit,omitempty"`
	DefaultQuantity      float64 `json:"defaultQuantity"`
	DefaultServicePrice  int64   `json:"defaultServicePrice"`
	DefaultMaterialPrice int64   `json:"defaultMaterialPrice"`
	Notes                string `json:"notes,omitempty"`
	Sequence             int     `json:"sequence"`
	IsActive             bool    `json:"isActive"`
	CategoryName         string  `json:"categoryName,omitempty"`
}

// PackageWorkSubItem — WorkSubItem; natural key dedup = (Pekerjaan induk, Nama) (ci).
type PackageWorkSubItem struct {
	Code                 string  `json:"code,omitempty"`
	Sequence             int     `json:"sequence"`
	Name                 string  `json:"name"`
	Description          string  `json:"description,omitempty"`
	DifficultyWeight     int     `json:"difficultyWeight"`
	DefaultUnit          string  `json:"defaultUnit,omitempty"`
	DefaultQuantity      float64 `json:"defaultQuantity"`
	DefaultServicePrice  int64   `json:"defaultServicePrice"`
	DefaultMaterialPrice int64   `json:"defaultMaterialPrice"`
	Notes                string  `json:"notes,omitempty"`
	IsActive             bool    `json:"isActive"`
	CategoryName         string  `json:"categoryName,omitempty"`
	WorkItemName         string  `json:"workItemName,omitempty"`
}

// PackageTemplateItem — item template; induk dirujuk via (kategori, pekerjaan).
type PackageTemplateItem struct {
	CategoryName string `json:"categoryName,omitempty"`
	WorkItemName string `json:"workItemName,omitempty"`
	Notes        string `json:"notes,omitempty"`
}

// PackageTemplate — Template; natural key dedup = Name (ci).
type PackageTemplate struct {
	Code        string                `json:"code,omitempty"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	Notes       string                `json:"notes,omitempty"`
	Sequence    int                   `json:"sequence"`
	IsActive    bool                  `json:"isActive"`
	Items       []PackageTemplateItem `json:"items,omitempty"`
}

// PackageCustomer — Customer; natural key dedup = Name (ci).
type PackageCustomer struct {
	Code        string         `json:"code,omitempty"`
	Name        string         `json:"name"`
	Address     string         `json:"address,omitempty"`
	Phone       string         `json:"phone,omitempty"`
	Email       string         `json:"email,omitempty"`
	PicName     string         `json:"picName,omitempty"`
	PicPosition string         `json:"picPosition,omitempty"`
	Notes       string         `json:"notes,omitempty"`
	IsActive    bool           `json:"isActive"`
	Vessels     []PackageVessel `json:"vessels,omitempty"`
}

// PackageVessel — Vessel; natural key dedup = (customer, Name) (ci).
type PackageVessel struct {
	Code         string `json:"code,omitempty"`
	Name         string `json:"name"`
	VesselNumber string `json:"vesselNumber,omitempty"`
	VesselType   string `json:"vesselType,omitempty"`
	Notes        string `json:"notes,omitempty"`
	IsActive     bool   `json:"isActive"`
}

// PackageMaterial — Material; natural key dedup = Name (ci).
type PackageMaterial struct {
	Code         string `json:"code,omitempty"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Unit         string `json:"unit,omitempty"`
	DefaultPrice int64  `json:"defaultPrice"`
	Supplier     string `json:"supplier,omitempty"`
	Notes        string `json:"notes,omitempty"`
	IsActive     bool   `json:"isActive"`
}

// PackageSphDocument — dokumen SPH lengkap beserta snapshot item & revisi.
// Natural key dedup = DocumentNumber (ci).
type PackageSphDocument struct {
	DocumentNumber   string              `json:"documentNumber"`
	Revision         int                 `json:"revision"`
	Date             time.Time           `json:"date"`
	CustomerName     string              `json:"customerName,omitempty"`
	VesselName       string              `json:"vesselName,omitempty"`
	ProjectName      string              `json:"projectName,omitempty"`
	Subject          string              `json:"subject,omitempty"`
	Reference        string              `json:"reference,omitempty"`
	Location         string              `json:"location,omitempty"`
	ValidUntil       *time.Time          `json:"validUntil,omitempty"`
	PicName          string              `json:"picName,omitempty"`
	Status           string              `json:"status"`
	SubtotalService  int64               `json:"subtotalService"`
	SubtotalMaterial int64               `json:"subtotalMaterial"`
	GrandTotal       int64               `json:"grandTotal"`
	Terbilang        string              `json:"terbilang,omitempty"`
	Notes            string              `json:"notes,omitempty"`
	FinalizedAt      *time.Time          `json:"finalizedAt,omitempty"`
	Items            []PackageSphItem    `json:"items"`
	Revisions        []PackageSphRevision `json:"revisions,omitempty"`
}

// PackageSphItem — snapshot main point SPH. WorkItemName opsional untuk tautan.
type PackageSphItem struct {
	Sequence            int               `json:"sequence"`
	WorkItemName        string            `json:"workItemName,omitempty"`
	NameSnapshot        string            `json:"nameSnapshot"`
	DescriptionSnapshot string            `json:"descriptionSnapshot,omitempty"`
	Quantity            float64           `json:"quantity"`
	Unit                string            `json:"unit,omitempty"`
	ServiceUnitPrice    int64             `json:"serviceUnitPrice"`
	MaterialUnitPrice   int64             `json:"materialUnitPrice"`
	ServiceTotal        int64             `json:"serviceTotal"`
	MaterialTotal       int64             `json:"materialTotal"`
	Total               int64             `json:"total"`
	PricingMode         string            `json:"pricingMode"`
	Notes               string            `json:"notes,omitempty"`
	SubItems            []PackageSphSubItem `json:"subItems,omitempty"`
}

// PackageSphSubItem — snapshot sub point SPH.
type PackageSphSubItem struct {
	Sequence            int     `json:"sequence"`
	NameSnapshot        string  `json:"nameSnapshot"`
	DescriptionSnapshot string  `json:"descriptionSnapshot,omitempty"`
	Quantity            float64 `json:"quantity"`
	Unit                string  `json:"unit,omitempty"`
	Weight              int     `json:"weight"`
	AllocatedValue      int64   `json:"allocatedValue"`
	ServiceUnitPrice    int64   `json:"serviceUnitPrice"`
	MaterialUnitPrice   int64   `json:"materialUnitPrice"`
	ServiceTotal        int64   `json:"serviceTotal"`
	MaterialTotal       int64   `json:"materialTotal"`
	Total               int64   `json:"total"`
	Notes               string  `json:"notes,omitempty"`
}

// PackageSphRevision — riwayat revisi dokumen SPH.
type PackageSphRevision struct {
	RevisionNumber int       `json:"revisionNumber"`
	Note           string    `json:"note,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

// Serialize mengubah paket ke JSON.
func (p *ShareBackupPackage) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

// ComputeChecksum menghitung SHA-256 dari metadata + seluruh seksi.
func (p *ShareBackupPackage) ComputeChecksum() (string, error) {
	frame, err := json.Marshal(struct {
		SchemaVersion string               `json:"schemaVersion"`
		Metadata      ShareBackupMetadata  `json:"metadata"`
		Categories    []PackageCategory    `json:"categories"`
		WorkItems     []PackageWorkItem    `json:"workItems"`
		WorkSubItems  []PackageWorkSubItem `json:"workSubItems"`
		Templates     []PackageTemplate    `json:"templates"`
		Customers     []PackageCustomer    `json:"customers"`
		Materials     []PackageMaterial    `json:"materials"`
		SphDocuments  []PackageSphDocument `json:"sphDocuments"`
	}{
		p.SchemaVersion, p.Metadata,
		p.Categories, p.WorkItems, p.WorkSubItems,
		p.Templates, p.Customers, p.Materials, p.SphDocuments,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(frame)
	return hex.EncodeToString(sum[:]), nil
}

// VerifyChecksum memverifikasi checksum yang tersimpan.
func (p *ShareBackupPackage) VerifyChecksum() (bool, error) {
	calc, err := p.ComputeChecksum()
	if err != nil {
		return false, err
	}
	return calc == p.Checksum, nil
}

// Deserialize mem-parse JSON paket.
func Deserialize(raw []byte) (*ShareBackupPackage, error) {
	var p ShareBackupPackage
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// SectionCounts jumlah item per seksi di dalam paket.
type SectionCounts struct {
	Sph        int `json:"sph"`
	WorkItems  int `json:"workItems"`
	Categories int `json:"categories"`
	Templates  int `json:"templates"`
	Customers  int `json:"customers"`
	Materials  int `json:"materials"`
}

// Counts menghitung isi tiap seksi.
func (p *ShareBackupPackage) Counts() SectionCounts {
	return SectionCounts{
		Sph:        len(p.SphDocuments),
		WorkItems:  len(p.WorkItems),
		Categories: len(p.Categories),
		Templates:  len(p.Templates),
		Customers:  len(p.Customers),
		Materials:  len(p.Materials),
	}
}