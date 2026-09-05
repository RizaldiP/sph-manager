package sharebackup

import (
	"os"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// Build mengambil seluruh data yang dapat dibagikan dari database dan
// mengemasnya dalam ShareBackupPackage (natural key = nama).
func (s *Service) Build() (*ShareBackupPackage, error) {
	device, _ := os.Hostname()
	pkg := &ShareBackupPackage{
		SchemaVersion: PackageSchemaVersion,
		Metadata: ShareBackupMetadata{
			PackageID:  newID(),
			DeviceName: device,
			CreatedAt:  now(),
		},
	}

	if err := s.buildCategories(pkg); err != nil {
		return nil, err
	}
	if err := s.buildWorkItems(pkg); err != nil {
		return nil, err
	}
	if err := s.buildTemplates(pkg); err != nil {
		return nil, err
	}
	if err := s.buildCustomers(pkg); err != nil {
		return nil, err
	}
	if err := s.buildMaterials(pkg); err != nil {
		return nil, err
	}
	if err := s.buildSphDocuments(pkg); err != nil {
		return nil, err
	}

	checksum, err := pkg.ComputeChecksum()
	if err != nil {
		return nil, err
	}
	pkg.Checksum = checksum
	return pkg, nil
}

func (s *Service) buildCategories(pkg *ShareBackupPackage) error {
	var rows []models.Category
	if err := s.db.Where("deleted_at IS NULL").Order("sequence asc, id asc").Find(&rows).Error; err != nil {
		return err
	}
	for _, c := range rows {
		pkg.Categories = append(pkg.Categories, PackageCategory{
			Code:        c.Code,
			Name:        c.Name,
			Description: c.Description,
			Sequence:    c.Sequence,
			IsActive:    c.IsActive,
		})
	}
	return nil
}

func (s *Service) buildWorkItems(pkg *ShareBackupPackage) error {
	var rows []models.WorkItem
	q := s.db.Where("deleted_at IS NULL").Order("sequence asc, id asc").Find(&rows)
	if q.Error != nil {
		return q.Error
	}
	catIDs := make(map[uint]*models.Category)
	for i := range rows {
		w := rows[i]
		cat, ok := catIDs[w.CategoryID]
		if !ok {
			cid, err := s.categoryByID(w.CategoryID)
			if err != nil {
				return err
			}
			cat = cid
			catIDs[w.CategoryID] = cid
		}
		var subs []models.WorkSubItem
		if q := s.db.Where("work_item_id = ? AND deleted_at IS NULL", w.ID).Order("sequence asc, id asc").Find(&subs); q.Error != nil {
			return q.Error
		}
		for _, sub := range subs {
			pkg.WorkSubItems = append(pkg.WorkSubItems, PackageWorkSubItem{
				Code:                 sub.Code,
				Sequence:             sub.Sequence,
				Name:                 sub.Name,
				Description:          sub.Description,
				DifficultyWeight:     sub.DifficultyWeight,
				DefaultUnit:          sub.DefaultUnit,
				DefaultQuantity:      sub.DefaultQuantity,
				DefaultServicePrice:  sub.DefaultServicePrice,
				DefaultMaterialPrice: sub.DefaultMaterialPrice,
				Notes:                sub.Notes,
				IsActive:             sub.IsActive,
				CategoryName:         cat.Name,
				WorkItemName:         w.Name,
			})
		}
		pkg.WorkItems = append(pkg.WorkItems, PackageWorkItem{
			Code:                 w.Code,
			Name:                 w.Name,
			Description:          w.Description,
			DefaultUnit:          w.DefaultUnit,
			DefaultQuantity:      w.DefaultQuantity,
			DefaultServicePrice:  w.DefaultServicePrice,
			DefaultMaterialPrice: w.DefaultMaterialPrice,
			Notes:                w.Notes,
			Sequence:             w.Sequence,
			IsActive:             w.IsActive,
			CategoryName:         cat.Name,
		})
	}
	return nil
}

func (s *Service) categoryByID(id uint) (*models.Category, error) {
	var c models.Category
	err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) buildTemplates(pkg *ShareBackupPackage) error {
	var rows []models.Template
	q := s.db.Where("deleted_at IS NULL").Order("sequence asc, id asc").Find(&rows)
	if q.Error != nil {
		return q.Error
	}
	for _, t := range rows {
		p := PackageTemplate{
			Code:        t.Code,
			Name:        t.Name,
			Description: t.Description,
			Notes:       t.Notes,
			Sequence:    t.Sequence,
			IsActive:    t.IsActive,
		}
		var items []models.TemplateItem
		if q := s.db.Where("template_id = ?", t.ID).Order("sequence asc, id asc").Find(&items); q.Error != nil {
			return q.Error
		}
		for _, it := range items {
			item := PackageTemplateItem{Notes: it.Notes}
			var wi models.WorkItem
			if err := s.db.Where("id = ? AND deleted_at IS NULL", it.WorkItemID).First(&wi).Error; err == nil {
				item.CategoryName = ""
				item.WorkItemName = wi.Name
				if cat, err := s.categoryByID(wi.CategoryID); err == nil {
					item.CategoryName = cat.Name
				}
			}
			p.Items = append(p.Items, item)
		}
		pkg.Templates = append(pkg.Templates, p)
	}
	return nil
}

func (s *Service) buildCustomers(pkg *ShareBackupPackage) error {
	var rows []models.Customer
	q := s.db.Where("deleted_at IS NULL").Order("id asc").Find(&rows)
	if q.Error != nil {
		return q.Error
	}
	for _, c := range rows {
		p := PackageCustomer{
			Code:        c.Code,
			Name:        c.Name,
			Address:     c.Address,
			Phone:       c.Phone,
			Email:       c.Email,
			PicName:     c.PicName,
			PicPosition: c.PicPosition,
			Notes:       c.Notes,
			IsActive:    c.IsActive,
		}
		var vessels []models.Vessel
		if q := s.db.Where("customer_id = ? AND deleted_at IS NULL", c.ID).Order("id asc").Find(&vessels); q.Error != nil {
			return q.Error
		}
		for _, v := range vessels {
			p.Vessels = append(p.Vessels, PackageVessel{
				Code:         v.Code,
				Name:         v.Name,
				VesselNumber: v.VesselNumber,
				VesselType:   v.VesselType,
				Notes:        v.Notes,
				IsActive:     v.IsActive,
			})
		}
		pkg.Customers = append(pkg.Customers, p)
	}
	return nil
}

func (s *Service) buildMaterials(pkg *ShareBackupPackage) error {
	var rows []models.Material
	q := s.db.Where("deleted_at IS NULL").Order("id asc").Find(&rows)
	if q.Error != nil {
		return q.Error
	}
	for _, m := range rows {
		pkg.Materials = append(pkg.Materials, PackageMaterial{
			Code:         m.Code,
			Name:         m.Name,
			Description:  m.Description,
			Unit:         m.Unit,
			DefaultPrice: m.DefaultPrice,
			Supplier:     m.Supplier,
			Notes:        m.Notes,
			IsActive:     m.IsActive,
		})
	}
	return nil
}

func (s *Service) buildSphDocuments(pkg *ShareBackupPackage) error {
	var docs []models.SphDocument
	q := s.db.Where("deleted_at IS NULL").Order("date desc, id desc").Find(&docs)
	if q.Error != nil {
		return q.Error
	}
	for _, d := range docs {
		p := PackageSphDocument{
			DocumentNumber:   d.DocumentNumber,
			Revision:         d.Revision,
			Date:             d.Date,
			ProjectName:      d.ProjectName,
			Subject:          d.Subject,
			Reference:        d.Reference,
			Location:         d.Location,
			ValidUntil:       d.ValidUntil,
			PicName:          d.PicName,
			Status:           d.Status,
			SubtotalService:  d.SubtotalService,
			SubtotalMaterial: d.SubtotalMaterial,
			GrandTotal:       d.GrandTotal,
			Terbilang:        d.Terbilang,
			Notes:            d.Notes,
			FinalizedAt:      d.FinalizedAt,
		}
		var c models.Customer
		if err := s.db.Where("id = ? AND deleted_at IS NULL", d.CustomerID).First(&c).Error; err == nil {
			p.CustomerName = c.Name
		}
		if d.VesselID != nil {
			var v models.Vessel
			if err := s.db.Where("id = ? AND deleted_at IS NULL", *d.VesselID).First(&v).Error; err == nil {
				p.VesselName = v.Name
			}
		}
		var items []models.SphItem
		if q := s.db.Where("sph_document_id = ?", d.ID).Order("sequence asc, id asc").Find(&items); q.Error != nil {
			return q.Error
		}
		for _, it := range items {
			pi := PackageSphItem{
				Sequence:            it.Sequence,
				NameSnapshot:        it.NameSnapshot,
				DescriptionSnapshot: it.DescriptionSnapshot,
				Quantity:            it.Quantity,
				Unit:                it.Unit,
				ServiceUnitPrice:    it.ServiceUnitPrice,
				MaterialUnitPrice:   it.MaterialUnitPrice,
				ServiceTotal:        it.ServiceTotal,
				MaterialTotal:       it.MaterialTotal,
				Total:               it.Total,
				PricingMode:         it.PricingMode,
				Notes:               it.Notes,
			}
			if it.WorkItemID != nil {
				var wi models.WorkItem
				if err := s.db.Where("id = ? AND deleted_at IS NULL", *it.WorkItemID).First(&wi).Error; err == nil {
					pi.WorkItemName = wi.Name
				}
			}
			var subs []models.SphSubItem
			if q := s.db.Where("sph_item_id = ?", it.ID).Order("sequence asc, id asc").Find(&subs); q.Error != nil {
				return q.Error
			}
			for _, sub := range subs {
				pi.SubItems = append(pi.SubItems, PackageSphSubItem{
					Sequence:            sub.Sequence,
					NameSnapshot:        sub.NameSnapshot,
					DescriptionSnapshot: sub.DescriptionSnapshot,
					Quantity:            sub.Quantity,
					Unit:                sub.Unit,
					Weight:              sub.Weight,
					AllocatedValue:      sub.AllocatedValue,
					ServiceUnitPrice:    sub.ServiceUnitPrice,
					MaterialUnitPrice:   sub.MaterialUnitPrice,
					ServiceTotal:        sub.ServiceTotal,
					MaterialTotal:       sub.MaterialTotal,
					Total:               sub.Total,
					Notes:               sub.Notes,
				})
			}
			p.Items = append(p.Items, pi)
		}
		var revs []models.SphRevision
		if q := s.db.Where("sph_document_id = ?", d.ID).Order("revision_number asc, id asc").Find(&revs); q.Error != nil {
			return q.Error
		}
		for _, r := range revs {
			p.Revisions = append(p.Revisions, PackageSphRevision{
				RevisionNumber: r.RevisionNumber,
				Note:           r.Note,
				CreatedAt:      r.CreatedAt,
			})
		}
		pkg.SphDocuments = append(pkg.SphDocuments, p)
	}
	return nil
}