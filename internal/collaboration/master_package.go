package collaboration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// SchemaVersion package Master Data.
const PackageSchemaVersion = "1"

// MasterDataPackage adalah representasi Master Data yang dikirim antar PC melalui
// LAN sebagai object JSON terstruktur (bukan salinan file DB mentah).
// Metadata + daftar entity; relasi dipetakan lewat natural key pada tiap entity
// (mis. WorkItem.CategoryCode, WorkSubItem.WorkItemCode).
type MasterDataPackage struct {
	Metadata MasterPackageMetadata `json:"metadata"`
	Data     MasterPackageData     `json:"data"`
	Checksum string                `json:"checksum"`
}

// MasterPackageMetadata menyimpan identitas & sumber package.
type MasterPackageMetadata struct {
	PackageID      string    `json:"packageId"`
	SenderID       string    `json:"senderId"`
	SenderName     string    `json:"senderName"`
	RoomID         string    `json:"roomId"`
	CreatedAt      time.Time `json:"createdAt"`
	SchemaVersion  string    `json:"schemaVersion"`
	PackageVersion string    `json:"packageVersion"`
	SourceVersion  int       `json:"sourceVersion"`
}

// MasterPackageData adalah payload entity Master Data (tanpa kolom audit/ID lokal).
type MasterPackageData struct {
	Categories   []PackageCategory    `json:"categories,omitempty"`
	WorkItems    []PackageWorkItem    `json:"workItems,omitempty"`
	WorkSubItems []PackageWorkSubItem `json:"workSubItems,omitempty"`
	Materials    []PackageMaterial    `json:"materials,omitempty"`
}

// PackageCategory — Category tanpa kolom audit/relasi, memakai natural key.
type PackageCategory struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Sequence    int    `json:"sequence"`
	IsActive    bool   `json:"isActive"`
}

// PackageWorkItem — WorkItem; CategoryCode ialah natural key induknya.
type PackageWorkItem struct {
	Code                 string  `json:"code"`
	Name                 string  `json:"name"`
	Description          string  `json:"description,omitempty"`
	DefaultUnit          string  `json:"defaultUnit,omitempty"`
	DefaultQuantity      float64 `json:"defaultQuantity"`
	DefaultServicePrice  int64   `json:"defaultServicePrice"`
	DefaultMaterialPrice int64   `json:"defaultMaterialPrice"`
	Notes                string  `json:"notes,omitempty"`
	Sequence             int     `json:"sequence"`
	IsActive             bool    `json:"isActive"`
	CategoryCode         string  `json:"categoryCode,omitempty"`
}

// PackageWorkSubItem — WorkSubItem; WorkItemCode ialah natural key induknya.
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
	WorkItemCode         string  `json:"workItemCode,omitempty"`
}

// PackageMaterial — Material mandiri.
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

// Serialize marshals package ke JSON.
func (p *MasterDataPackage) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

// ComputeChecksum menghitung SHA-256 dari bagian data+metadata (tanpa field checksum).
func (p *MasterDataPackage) ComputeChecksum() (string, error) {
	frame, err := json.Marshal(struct {
		Metadata MasterPackageMetadata `json:"metadata"`
		Data     MasterPackageData     `json:"data"`
	}{p.Metadata, p.Data})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(frame)
	return hex.EncodeToString(sum[:]), nil
}

// VerifyChecksum memverifikasi checksum yang tersimpan.
func (p *MasterDataPackage) VerifyChecksum() (bool, error) {
	calc, err := p.ComputeChecksum()
	if err != nil {
		return false, err
	}
	return calc == p.Checksum, nil
}

// DeserializeMasterDataPackage mem-parse JSON package dari wire/DB.
func DeserializeMasterDataPackage(raw []byte) (*MasterDataPackage, error) {
	var p MasterDataPackage
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
