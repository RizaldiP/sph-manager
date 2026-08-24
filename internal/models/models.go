package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusDraft     = "DRAFT"
	StatusReview    = "REVIEW"
	StatusFinal     = "FINAL"
	StatusSent      = "SENT"
	StatusAccepted  = "ACCEPTED"
	StatusRejected  = "REJECTED"
	StatusCancelled = "CANCELLED"
)

var AllStatuses = []string{
	StatusDraft,
	StatusReview,
	StatusFinal,
	StatusSent,
	StatusAccepted,
	StatusRejected,
	StatusCancelled,
}

const (
	PricingModeDirect = "HARGA_LANGSUNG"
	PricingModeWeight = "PEMBOBOTAN"
)

type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Code        string         `gorm:"size:50;notNull;index" json:"code"`
	Name        string         `gorm:"size:150;notNull" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	Sequence    int            `gorm:"notNull;default:0;index" json:"sequence"`
	IsActive    bool           `gorm:"notNull;default:true;index" json:"isActive"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	WorkItems []WorkItem `gorm:"foreignKey:CategoryID;constraint:OnDelete:RESTRICT" json:"workItems,omitempty"`
}

func (Category) TableName() string { return "categories" }

type WorkItem struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	CategoryID           uint           `gorm:"notNull;index" json:"categoryId"`
	Code                 string         `gorm:"size:50;notNull;index" json:"code"`
	Name                 string         `gorm:"size:300;notNull" json:"name"`
	Description          string         `gorm:"size:1000" json:"description"`
	DefaultUnit          string         `gorm:"size:30" json:"defaultUnit"`
	DefaultQuantity      float64        `gorm:"notNull;default:1" json:"defaultQuantity"`
	DefaultServicePrice  int64          `gorm:"notNull;default:0" json:"defaultServicePrice"`
	DefaultMaterialPrice int64          `gorm:"notNull;default:0" json:"defaultMaterialPrice"`
	Notes                string         `gorm:"size:1000" json:"notes"`
	Sequence             int            `gorm:"notNull;default:0;index" json:"sequence"`
	IsActive             bool           `gorm:"notNull;default:true;index" json:"isActive"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	Category Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	SubItems []WorkSubItem `gorm:"foreignKey:WorkItemID;constraint:OnDelete:RESTRICT" json:"subItems,omitempty"`
}

func (WorkItem) TableName() string { return "work_items" }

type WorkSubItem struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	WorkItemID           uint           `gorm:"notNull;index" json:"workItemId"`
	Code                 string         `gorm:"size:50;index" json:"code"`
	Sequence             int            `gorm:"notNull;default:0;index" json:"sequence"`
	Name                 string         `gorm:"size:300;notNull" json:"name"`
	Description          string         `gorm:"size:1000" json:"description"`
	DifficultyWeight     int            `gorm:"notNull;default:0" json:"difficultyWeight"`
	DefaultUnit          string         `gorm:"size:30" json:"defaultUnit"`
	DefaultQuantity      float64        `gorm:"notNull;default:1" json:"defaultQuantity"`
	DefaultServicePrice  int64          `gorm:"notNull;default:0" json:"defaultServicePrice"`
	DefaultMaterialPrice int64          `gorm:"notNull;default:0" json:"defaultMaterialPrice"`
	Notes                string         `gorm:"size:1000" json:"notes"`
	IsActive             bool           `gorm:"notNull;default:true;index" json:"isActive"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	WorkItem WorkItem `gorm:"foreignKey:WorkItemID" json:"workItem,omitempty"`
}

func (WorkSubItem) TableName() string { return "work_sub_items" }

type Template struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Code        string         `gorm:"size:50;index" json:"code"`
	Name        string         `gorm:"size:200;notNull" json:"name"`
	Description string         `gorm:"size:1000" json:"description"`
	Notes       string         `gorm:"size:1000" json:"notes"`
	Sequence    int            `gorm:"notNull;default:0;index" json:"sequence"`
	IsActive    bool           `gorm:"notNull;default:true;index" json:"isActive"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	Items []TemplateItem `gorm:"foreignKey:TemplateID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
}

func (Template) TableName() string { return "templates" }

type TemplateItem struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TemplateID uint      `gorm:"notNull;index" json:"templateId"`
	Sequence   int       `gorm:"notNull;default:0;index" json:"sequence"`
	WorkItemID uint      `gorm:"notNull;index" json:"workItemId"`
	Notes      string    `gorm:"size:500" json:"notes"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`

	WorkItem WorkItem `gorm:"foreignKey:WorkItemID;constraint:OnDelete:RESTRICT" json:"workItem,omitempty"`
}

func (TemplateItem) TableName() string { return "template_items" }

type Customer struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Code        string         `gorm:"size:50;index" json:"code"`
	Name        string         `gorm:"size:200;notNull" json:"name"`
	Address     string         `gorm:"size:500" json:"address"`
	Phone       string         `gorm:"size:50" json:"phone"`
	Email       string         `gorm:"size:150" json:"email"`
	PicName     string         `gorm:"size:150" json:"picName"`
	PicPosition string         `gorm:"size:150" json:"picPosition"`
	Notes       string         `gorm:"size:1000" json:"notes"`
	IsActive    bool           `gorm:"notNull;default:true;index" json:"isActive"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	Vessels []Vessel `gorm:"foreignKey:CustomerID;constraint:OnDelete:RESTRICT" json:"vessels,omitempty"`
}

func (Customer) TableName() string { return "customers" }

type Vessel struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CustomerID   uint           `gorm:"notNull;index" json:"customerId"`
	Code         string         `gorm:"size:50;index" json:"code"`
	Name         string         `gorm:"size:200;notNull" json:"name"`
	VesselNumber string         `gorm:"size:100" json:"vesselNumber"`
	VesselType   string         `gorm:"size:150" json:"vesselType"`
	Notes        string         `gorm:"size:1000" json:"notes"`
	IsActive     bool           `gorm:"notNull;default:true;index" json:"isActive"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	Customer Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}

func (Vessel) TableName() string { return "vessels" }

type Material struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Code         string         `gorm:"size:50;index" json:"code"`
	Name         string         `gorm:"size:300;notNull" json:"name"`
	Description  string         `gorm:"size:1000" json:"description"`
	Unit         string         `gorm:"size:30" json:"unit"`
	DefaultPrice int64          `gorm:"notNull;default:0" json:"defaultPrice"`
	Supplier     string         `gorm:"size:200" json:"supplier"`
	Notes        string         `gorm:"size:1000" json:"notes"`
	IsActive     bool           `gorm:"notNull;default:true;index" json:"isActive"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

func (Material) TableName() string { return "materials" }

type SphDocument struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	DocumentNumber   string         `gorm:"size:100;notNull;index" json:"documentNumber"`
	Revision         int            `gorm:"notNull;default:0;index" json:"revision"`
	Date             time.Time      `gorm:"notNull" json:"date"`
	CustomerID       uint           `gorm:"notNull;index" json:"customerId"`
	VesselID         *uint          `gorm:"index" json:"vesselId,omitempty"`
	ProjectName      string         `gorm:"size:300" json:"projectName"`
	Subject          string         `gorm:"size:500" json:"subject"`
	Reference        string         `gorm:"size:300" json:"reference"`
	Location         string         `gorm:"size:200" json:"location"`
	ValidUntil       *time.Time     `json:"validUntil,omitempty"`
	PicName          string         `gorm:"size:150" json:"picName"`
	Status           string         `gorm:"size:20;notNull;default:DRAFT;index" json:"status"`
	SubtotalService  int64          `gorm:"notNull;default:0" json:"subtotalService"`
	SubtotalMaterial int64          `gorm:"notNull;default:0" json:"subtotalMaterial"`
	GrandTotal       int64          `gorm:"notNull;default:0" json:"grandTotal"`
	Terbilang        string         `gorm:"size:500" json:"terbilang"`
	Notes            string         `gorm:"size:2000" json:"notes"`
	FinalizedAt      *time.Time     `json:"finalizedAt,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	Customer  Customer      `gorm:"foreignKey:CustomerID;constraint:OnDelete:RESTRICT" json:"customer,omitempty"`
	Vessel    *Vessel       `gorm:"foreignKey:VesselID;constraint:OnDelete:RESTRICT" json:"vessel,omitempty"`
	Items     []SphItem     `gorm:"foreignKey:SphDocumentID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	Revisions []SphRevision `gorm:"foreignKey:SphDocumentID;constraint:OnDelete:CASCADE" json:"revisions,omitempty"`
}

func (SphDocument) TableName() string { return "sph_documents" }

type SphItem struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	SphDocumentID       uint      `gorm:"notNull;index" json:"sphDocumentId"`
	Sequence            int       `gorm:"notNull;default:0;index" json:"sequence"`
	WorkItemID          *uint     `gorm:"index" json:"workItemId,omitempty"`
	NameSnapshot        string    `gorm:"size:300;notNull" json:"nameSnapshot"`
	DescriptionSnapshot string    `gorm:"size:1000" json:"descriptionSnapshot"`
	Quantity            float64   `gorm:"notNull;default:1" json:"quantity"`
	Unit                string    `gorm:"size:30" json:"unit"`
	ServiceUnitPrice    int64     `gorm:"notNull;default:0" json:"serviceUnitPrice"`
	MaterialUnitPrice   int64     `gorm:"notNull;default:0" json:"materialUnitPrice"`
	ServiceTotal        int64     `gorm:"notNull;default:0" json:"serviceTotal"`
	MaterialTotal       int64     `gorm:"notNull;default:0" json:"materialTotal"`
	Total               int64     `gorm:"notNull;default:0" json:"total"`
	PricingMode         string    `gorm:"size:20;notNull;default:HARGA_LANGSUNG" json:"pricingMode"`
	Notes               string    `gorm:"size:1000" json:"notes"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`

	WorkItem *WorkItem    `gorm:"foreignKey:WorkItemID;constraint:OnDelete:RESTRICT" json:"workItem,omitempty"`
	SubItems []SphSubItem `gorm:"foreignKey:SphItemID;constraint:OnDelete:CASCADE" json:"subItems,omitempty"`
}

func (SphItem) TableName() string { return "sph_items" }

type SphSubItem struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	SphItemID           uint      `gorm:"notNull;index" json:"sphItemId"`
	Sequence            int       `gorm:"notNull;default:0;index" json:"sequence"`
	NameSnapshot        string    `gorm:"size:300;notNull" json:"nameSnapshot"`
	DescriptionSnapshot string    `gorm:"size:1000" json:"descriptionSnapshot"`
	Quantity            float64   `gorm:"notNull;default:1" json:"quantity"`
	Unit                string    `gorm:"size:30" json:"unit"`
	Weight              int       `gorm:"notNull;default:0" json:"weight"`
	AllocatedValue      int64     `gorm:"notNull;default:0" json:"allocatedValue"`
	ServiceUnitPrice    int64     `gorm:"notNull;default:0" json:"serviceUnitPrice"`
	MaterialUnitPrice   int64     `gorm:"notNull;default:0" json:"materialUnitPrice"`
	ServiceTotal        int64     `gorm:"notNull;default:0" json:"serviceTotal"`
	MaterialTotal       int64     `gorm:"notNull;default:0" json:"materialTotal"`
	Total               int64     `gorm:"notNull;default:0" json:"total"`
	Notes               string    `gorm:"size:1000" json:"notes"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

func (SphSubItem) TableName() string { return "sph_sub_items" }

type SphRevision struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SphDocumentID  uint      `gorm:"notNull;index" json:"sphDocumentId"`
	FromDocumentID *uint     `gorm:"index" json:"fromDocumentId,omitempty"`
	RevisionNumber int       `gorm:"notNull" json:"revisionNumber"`
	Note           string    `gorm:"size:1000" json:"note"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (SphRevision) TableName() string { return "sph_revisions" }

type AuditLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Action      string    `gorm:"size:20;notNull;index" json:"action"`
	Entity      string    `gorm:"size:60;notNull;index" json:"entity"`
	EntityID    uint      `gorm:"notNull;index" json:"entityId"`
	Description string    `gorm:"size:1000" json:"description"`
	CreatedAt   time.Time `gorm:"index" json:"createdAt"`
}

func (AuditLog) TableName() string { return "audit_logs" }

type Setting struct {
	Key       string    `gorm:"column:key;size:100;primaryKey" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Setting) TableName() string { return "settings" }
