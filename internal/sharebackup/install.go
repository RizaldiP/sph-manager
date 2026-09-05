package sharebackup

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/RizaldiP/sph-manager/internal/models"
)

// InstallOptions menentukan seksi mana yang diimpor.
type InstallOptions struct {
	Sections map[SectionKey]bool
}

// InstallSelectedSections membuat InstallOptions dari daftar kunci seksi.
func InstallSelectedSections(keys []string) InstallOptions {
	sel := map[SectionKey]bool{}
	for _, k := range keys {
		sel[SectionKey(k)] = true
	}
	return InstallOptions{Sections: sel}
}

// SectionInstallResult ringkasan hasil satu seksi.
type SectionInstallResult struct {
	Added         int `json:"added"`
	Skipped       int `json:"skipped"`
	CodeGenerated int `json:"codeGenerated"`
}

// InstallSummary ringkasan seluruh proses restore.
type InstallSummary struct {
	Categories SectionInstallResult `json:"categories"`
	WorkItems  SectionInstallResult `json:"workItems"`
	SubItems   SectionInstallResult `json:"subItems"`
	Templates  SectionInstallResult `json:"templates"`
	Customers  SectionInstallResult `json:"customers"`
	Vessels    SectionInstallResult `json:"vessels"`
	Materials  SectionInstallResult `json:"materials"`
	Sph        SectionInstallResult `json:"sph"`

	TemplateItemsAdded  int `json:"templateItemsAdded"`
	TemplateItemsMissed int `json:"templateItemsMissed"`
	SphItemsUnlinked    int `json:"sphItemsUnlinked"`
}

func allSelected() map[SectionKey]bool {
	sel := map[SectionKey]bool{}
	for _, k := range AllSectionKeys() {
		sel[k] = true
	}
	return sel
}

// normalizeSelection mengembalikan peta seksi terpilih; bila opts kosong,
// semua seksi dianggap dipilih.
func normalizeSelection(opts InstallOptions) map[SectionKey]bool {
	sel := allSelected()
	if opts.Sections != nil {
		return opts.Sections
	}
	return sel
}

// Restore mengimpor paket ke database secara transaksional, hanya menambah
// data baru (data yang sudah ada dilewati, tidak pernah ditimpa/dihapus).
func (s *Service) Restore(pkg *ShareBackupPackage, opts InstallOptions) (*InstallSummary, error) {
	in := newInstaller(s.db, pkg, normalizeSelection(opts))
	if err := in.load(); err != nil {
		return nil, err
	}
	err := in.db.Transaction(func(tx *gorm.DB) error {
		in.tx = tx
		if in.sel[SectionCategories] {
			if err := in.importCategories(); err != nil {
				return err
			}
		}
		if in.sel[SectionWorkItems] {
			if err := in.importWorkItems(); err != nil {
				return err
			}
			if err := in.importSubItems(); err != nil {
				return err
			}
		}
		if in.sel[SectionTemplates] {
			if err := in.importTemplates(); err != nil {
				return err
			}
		}
		if in.sel[SectionCustomers] {
			if err := in.importCustomers(); err != nil {
				return err
			}
		}
		if in.sel[SectionMaterials] {
			if err := in.importMaterials(); err != nil {
				return err
			}
		}
		if in.sel[SectionSph] {
			if err := in.importSph(); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return in.sum, nil
}

// installer menyimpan state bersama selama proses restore dalam satu transaksi.
type installer struct {
	db  *gorm.DB
	tx  *gorm.DB
	pkg *ShareBackupPackage
	sel map[SectionKey]bool
	sum *InstallSummary

	// categories
	w  map[string]uint // by name
	c  map[string]uint // by code
	caS int            // max sequence
	// work items
	wi  map[string]uint // (catName, wiName)
	wiN map[string]uint // wiName (untuk tautan opsional)
	wio map[string]uint // by code
	wiS map[string]int  // per-cat max sequence (key: low(catName))
	// sub items
	sub  map[string]uint // (wiName, subName)
	subo map[string]uint // by code
	subS map[string]int  // per-wi max sequence (key: low(wiName))
	// templates
	tp  map[string]uint // by name
	tpl map[string]uint // by code
	tpS int             // max sequence
	// customers & vessels
	cu  map[string]uint // by name
	cuo map[string]uint // by code
	ve  map[string]uint // (cusID, name)
	vec map[string]uint // by code
	// materials
	ma  map[string]uint // by name
	mao map[string]uint // by code
	// sph
	sph map[string]uint // by document number
}

func newInstaller(db *gorm.DB, pkg *ShareBackupPackage, sel map[SectionKey]bool) *installer {
	return &installer{
		db:   db,
		pkg:  pkg,
		sel:  sel,
		sum:  &InstallSummary{},
		w:    make(map[string]uint),
		c:    make(map[string]uint),
		wi:   make(map[string]uint),
		wiN:  make(map[string]uint),
		wio:  make(map[string]uint),
		wiS:  make(map[string]int),
		sub:  make(map[string]uint),
		subo: make(map[string]uint),
		subS: make(map[string]int),
		tp:   make(map[string]uint),
		tpl:  make(map[string]uint),
		cu:   make(map[string]uint),
		cuo:  make(map[string]uint),
		ve:   make(map[string]uint),
		vec:  make(map[string]uint),
		ma:   make(map[string]uint),
		mao:  make(map[string]uint),
		sph:  make(map[string]uint),
	}
}

func low(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func wiKey(catName, wiName string) string { return low(catName) + "\x00" + low(wiName) }
func subKey(wiName, subName string) string { return low(wiName) + "\x00" + low(subName) }
func vesKey(cusID uint, name string) string {
	return fmt.Sprintf("%d\x00%s", cusID, low(name))
}

func (in *installer) load() error {
	var cats []models.Category
	if err := in.db.Where("deleted_at IS NULL").Find(&cats).Error; err != nil {
		return err
	}
	for _, cat := range cats {
		in.w[low(cat.Name)] = cat.ID
		in.c[low(cat.Code)] = cat.ID
		if cat.Sequence > in.caS {
			in.caS = cat.Sequence
		}
	}

	type wiRow struct {
		ID           uint
		Code         string
		Name         string
		Sequence     int
		CategoryID   uint
		CategoryName string
	}
	var wis []wiRow
	if err := in.db.Table("work_items AS w").
		Select("w.id, w.code, w.name, w.sequence, w.category_id, c.name AS category_name").
		Joins("JOIN categories c ON c.id = w.category_id").
		Where("w.deleted_at IS NULL AND c.deleted_at IS NULL").
		Scan(&wis).Error; err != nil {
		return err
	}
	for _, r := range wis {
		in.wi[wiKey(r.CategoryName, r.Name)] = r.ID
		in.wiN[low(r.Name)] = r.ID
		in.wio[low(r.Code)] = r.ID
		k := low(r.CategoryName)
		if r.Sequence > in.wiS[k] {
			in.wiS[k] = r.Sequence
		}
	}

	type subRow struct {
		ID           uint
		Code         string
		Name         string
		Sequence     int
		WorkItemName string
	}
	var subs []subRow
	if err := in.db.Table("work_sub_items AS s").
		Select("s.id, s.code, s.name, s.sequence, w.name AS work_item_name").
		Joins("JOIN work_items w ON w.id = s.work_item_id").
		Where("s.deleted_at IS NULL AND w.deleted_at IS NULL").
		Scan(&subs).Error; err != nil {
		return err
	}
	for _, r := range subs {
		in.sub[subKey(r.WorkItemName, r.Name)] = r.ID
		in.subo[low(r.Code)] = r.ID
		k := low(r.WorkItemName)
		if r.Sequence > in.subS[k] {
			in.subS[k] = r.Sequence
		}
	}

	var tpls []models.Template
	if err := in.db.Where("deleted_at IS NULL").Find(&tpls).Error; err != nil {
		return err
	}
	for _, t := range tpls {
		in.tp[low(t.Name)] = t.ID
		in.tpl[low(t.Code)] = t.ID
		if t.Sequence > in.tpS {
			in.tpS = t.Sequence
		}
	}

	var cus []models.Customer
	if err := in.db.Where("deleted_at IS NULL").Find(&cus).Error; err != nil {
		return err
	}
	for _, cu := range cus {
		in.cu[low(cu.Name)] = cu.ID
		in.cuo[low(cu.Code)] = cu.ID
	}

	var ves []models.Vessel
	if err := in.db.Where("deleted_at IS NULL").Find(&ves).Error; err != nil {
		return err
	}
	for _, v := range ves {
		in.ve[vesKey(v.CustomerID, v.Name)] = v.ID
		in.vec[low(v.Code)] = v.ID
	}

	var mats []models.Material
	if err := in.db.Where("deleted_at IS NULL").Find(&mats).Error; err != nil {
		return err
	}
	for _, m := range mats {
		in.ma[low(m.Name)] = m.ID
		in.mao[low(m.Code)] = m.ID
	}

	var nums []string
	if err := in.db.Table("sph_documents").Where("deleted_at IS NULL").Pluck("document_number", &nums).Error; err != nil {
		return err
	}
	for _, n := range nums {
		in.sph[low(n)] = 1
	}
	return nil
}

// resolveCode mengembalikan kode yang dipakai: kode pengirim bila masih bebas
// di perangkat lokal; bila kosong/sudah dipakai entitas lain dihasilkan kode
// lokal baru.
func (in *installer) resolveCode(incoming string, byCode map[string]uint, table, prefix string, out *SectionInstallResult) (string, error) {
	code := strings.TrimSpace(incoming)
	if code != "" {
		if _, taken := byCode[low(code)]; !taken {
			byCode[low(code)] = 0
			return code, nil
		}
	}
	nc, err := nextCode(in.tx, table, prefix)
	if err != nil {
		return "", err
	}
	byCode[low(nc)] = 0
	out.CodeGenerated++
	return nc, nil
}

// ensureCategory membuat kategori lokal bila belum ada. Bila seksi kategori
// tidak dipilih, data tetap diambil dari paket agar induk rujukan valid.
func (in *installer) ensureCategory(name string) (uint, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, nil
	}
	if id, ok := in.w[low(name)]; ok {
		return id, nil
	}
	p := PackageCategory{Name: name, Sequence: in.caS + 1, IsActive: true}
	for _, c := range in.pkg.Categories {
		if low(c.Name) == low(name) {
			p = c
			break
		}
	}
	if p.Sequence <= in.caS {
		p.Sequence = in.caS + 1
	}
	code, err := in.resolveCode(p.Code, in.c, "categories", "KAT-", &in.sum.Categories)
	if err != nil {
		return 0, err
	}
	cat := models.Category{
		Code:        code,
		Name:        p.Name,
		Description: p.Description,
		Sequence:    p.Sequence,
		IsActive:    p.IsActive,
	}
	if err := in.tx.Create(&cat).Error; err != nil {
		return 0, err
	}
	in.w[low(cat.Name)] = cat.ID
	if cat.Sequence > in.caS {
		in.caS = cat.Sequence
	}
	in.sum.Categories.Added++
	return cat.ID, nil
}

// ensureWorkItem membuat pekerjaan lokal bila belum ada, mengambil data dari
// paket bila seksi master pekerjaan tidak dipilih.
func (in *installer) ensureWorkItem(catName, wiName string) (uint, error) {
	catName = strings.TrimSpace(catName)
	wiName = strings.TrimSpace(wiName)
	if wiName == "" {
		return 0, nil
	}
	if id, ok := in.wi[wiKey(catName, wiName)]; ok {
		return id, nil
	}
	catID, err := in.ensureCategory(catName)
	if err != nil {
		return 0, err
	}
	if catID == 0 {
		return 0, nil
	}
	var p *PackageWorkItem
	for i := range in.pkg.WorkItems {
		if wiKey(in.pkg.WorkItems[i].CategoryName, in.pkg.WorkItems[i].Name) == wiKey(catName, wiName) {
			p = &in.pkg.WorkItems[i]
			break
		}
	}
	if p == nil {
		return 0, nil
	}
	k := low(catName)
	seq := in.wiS[k] + 1
	code, err := in.resolveCode(p.Code, in.wio, "work_items", "PEK-", &in.sum.WorkItems)
	if err != nil {
		return 0, err
	}
	wi := models.WorkItem{
		CategoryID:           catID,
		Code:                 code,
		Name:                 p.Name,
		Description:          p.Description,
		DefaultUnit:          p.DefaultUnit,
		DefaultQuantity:      p.DefaultQuantity,
		DefaultServicePrice:  p.DefaultServicePrice,
		DefaultMaterialPrice: p.DefaultMaterialPrice,
		Notes:                p.Notes,
		Sequence:             seq,
		IsActive:             p.IsActive,
	}
	if err := in.tx.Create(&wi).Error; err != nil {
		return 0, err
	}
	in.wi[wiKey(catName, wiName)] = wi.ID
	in.wiN[low(p.Name)] = wi.ID
	in.wiS[k] = seq
	in.sum.WorkItems.Added++
	return wi.ID, nil
}

func (in *installer) importCategories() error {
	for _, c := range in.pkg.Categories {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			in.sum.Categories.Skipped++
			continue
		}
		if _, exists := in.w[low(name)]; exists {
			in.sum.Categories.Skipped++
			continue
		}
		if _, err := in.ensureCategory(name); err != nil {
			return err
		}
	}
	return nil
}

func (in *installer) importWorkItems() error {
	for _, p := range in.pkg.WorkItems {
		wiName := strings.TrimSpace(p.Name)
		if wiName == "" {
			in.sum.WorkItems.Skipped++
			continue
		}
		if _, exists := in.wi[wiKey(p.CategoryName, p.Name)]; exists {
			in.sum.WorkItems.Skipped++
			continue
		}
		if _, err := in.ensureWorkItem(p.CategoryName, p.Name); err != nil {
			return err
		}
	}
	return nil
}

func (in *installer) importSubItems() error {
	for _, p := range in.pkg.WorkSubItems {
		subName := strings.TrimSpace(p.Name)
		if subName == "" {
			in.sum.SubItems.Skipped++
			continue
		}
		wiID, err := in.ensureWorkItem(p.CategoryName, p.WorkItemName)
		if err != nil {
			return err
		}
		if wiID == 0 {
			in.sum.SubItems.Skipped++
			continue
		}
		if _, exists := in.sub[subKey(p.WorkItemName, subName)]; exists {
			in.sum.SubItems.Skipped++
			continue
		}
		seq := in.subS[low(p.WorkItemName)] + 1
		code, err := in.resolveCode(p.Code, in.subo, "work_sub_items", "SUB-", &in.sum.SubItems)
		if err != nil {
			return err
		}
		sub := models.WorkSubItem{
			WorkItemID:           wiID,
			Code:                 code,
			Sequence:             seq,
			Name:                 p.Name,
			Description:          p.Description,
			DifficultyWeight:     p.DifficultyWeight,
			DefaultUnit:          p.DefaultUnit,
			DefaultQuantity:      p.DefaultQuantity,
			DefaultServicePrice:  p.DefaultServicePrice,
			DefaultMaterialPrice: p.DefaultMaterialPrice,
			Notes:                p.Notes,
			IsActive:             p.IsActive,
		}
		if err := in.tx.Create(&sub).Error; err != nil {
			return err
		}
		in.sub[subKey(p.WorkItemName, p.Name)] = sub.ID
		in.subS[low(p.WorkItemName)] = seq
		in.sum.SubItems.Added++
	}
	return nil
}

func (in *installer) importTemplates() error {
	for _, p := range in.pkg.Templates {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			in.sum.Templates.Skipped++
			continue
		}
		if _, exists := in.tp[low(name)]; exists {
			in.sum.Templates.Skipped++
			continue
		}
		seq := in.tpS + 1
		code, err := in.resolveCode(p.Code, in.tpl, "templates", "TPL-", &in.sum.Templates)
		if err != nil {
			return err
		}
		tpl := models.Template{
			Code:        code,
			Name:        p.Name,
			Description: p.Description,
			Notes:       p.Notes,
			Sequence:    seq,
			IsActive:    p.IsActive,
		}
		if err := in.tx.Create(&tpl).Error; err != nil {
			return err
		}
		in.tp[low(tpl.Name)] = tpl.ID
		in.tpS = seq
		itemSeq := 1
		for _, it := range p.Items {
			wiID, err := in.ensureWorkItem(it.CategoryName, it.WorkItemName)
			if err != nil {
				return err
			}
			if wiID == 0 {
				in.sum.TemplateItemsMissed++
				continue
			}
			ti := models.TemplateItem{
				TemplateID: tpl.ID,
				Sequence:   itemSeq,
				WorkItemID: wiID,
				Notes:      it.Notes,
			}
			if err := in.tx.Create(&ti).Error; err != nil {
				return err
			}
			itemSeq++
			in.sum.TemplateItemsAdded++
		}
		in.sum.Templates.Added++
	}
	return nil
}

func (in *installer) ensureVessels(cusID uint, p PackageCustomer) error {
	for _, v := range p.Vessels {
		if strings.TrimSpace(v.Name) == "" {
			in.sum.Vessels.Skipped++
			continue
		}
		if _, exists := in.ve[vesKey(cusID, v.Name)]; exists {
			in.sum.Vessels.Skipped++
			continue
		}
		code, err := in.resolveCode(v.Code, in.vec, "vessels", "KAP-", &in.sum.Vessels)
		if err != nil {
			return err
		}
		ves := models.Vessel{
			CustomerID:   cusID,
			Code:         code,
			Name:         v.Name,
			VesselNumber: v.VesselNumber,
			VesselType:   v.VesselType,
			Notes:        v.Notes,
			IsActive:     v.IsActive,
		}
		if err := in.tx.Create(&ves).Error; err != nil {
			return err
		}
		in.ve[vesKey(cusID, v.Name)] = ves.ID
		in.sum.Vessels.Added++
	}
	return nil
}

func (in *installer) importCustomers() error {
	for _, p := range in.pkg.Customers {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			in.sum.Customers.Skipped++
			continue
		}
		cusID, exists := in.cu[low(name)]
		if !exists {
			code, err := in.resolveCode(p.Code, in.cuo, "customers", "CUS-", &in.sum.Customers)
			if err != nil {
				return err
			}
			cus := models.Customer{
				Code:        code,
				Name:        p.Name,
				Address:     p.Address,
				Phone:       p.Phone,
				Email:       p.Email,
				PicName:     p.PicName,
				PicPosition: p.PicPosition,
				Notes:       p.Notes,
				IsActive:    p.IsActive,
			}
			if err := in.tx.Create(&cus).Error; err != nil {
				return err
			}
			in.cu[low(cus.Name)] = cus.ID
			in.sum.Customers.Added++
			cusID = cus.ID
		}
		if err := in.ensureVessels(cusID, p); err != nil {
			return err
		}
	}
	return nil
}

func (in *installer) importMaterials() error {
	for _, p := range in.pkg.Materials {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			in.sum.Materials.Skipped++
			continue
		}
		if _, exists := in.ma[low(name)]; exists {
			in.sum.Materials.Skipped++
			continue
		}
		code, err := in.resolveCode(p.Code, in.mao, "materials", "MAT-", &in.sum.Materials)
		if err != nil {
			return err
		}
		mat := models.Material{
			Code:         code,
			Name:         p.Name,
			Description:  p.Description,
			Unit:         p.Unit,
			DefaultPrice: p.DefaultPrice,
			Supplier:     p.Supplier,
			Notes:        p.Notes,
			IsActive:     p.IsActive,
		}
		if err := in.tx.Create(&mat).Error; err != nil {
			return err
		}
		in.ma[low(mat.Name)] = mat.ID
		in.sum.Materials.Added++
	}
	return nil
}

// ensureCustomerFromPkg membuat customer lokal dari data paket bila belum ada
// (dipakai ketika seksi sph diimpor tanpa seksi customer).
func (in *installer) ensureCustomerFromPkg(name string) (uint, error) {
	if id, ok := in.cu[low(name)]; ok {
		return id, nil
	}
	for _, c := range in.pkg.Customers {
		if low(c.Name) != low(name) {
			continue
		}
		code, err := in.resolveCode(c.Code, in.cuo, "customers", "CUS-", &in.sum.Customers)
		if err != nil {
			return 0, err
		}
		cus := models.Customer{
			Code:        code,
			Name:        c.Name,
			Address:     c.Address,
			Phone:       c.Phone,
			Email:       c.Email,
			PicName:     c.PicName,
			PicPosition: c.PicPosition,
			Notes:       c.Notes,
			IsActive:    c.IsActive,
		}
		if err := in.tx.Create(&cus).Error; err != nil {
			return 0, err
		}
		in.cu[low(cus.Name)] = cus.ID
		in.sum.Customers.Added++
		return cus.ID, nil
	}
	return 0, nil
}

// ensureVesselFromPkg membuat kapal lokal dari data paket bila belum ada.
func (in *installer) ensureVesselFromPkg(cusID uint, cusName, vesselName string) (uint, error) {
	if id, ok := in.ve[vesKey(cusID, vesselName)]; ok {
		return id, nil
	}
	for _, c := range in.pkg.Customers {
		if low(c.Name) != low(cusName) {
			continue
		}
		for _, v := range c.Vessels {
			if low(v.Name) != low(vesselName) {
				continue
			}
			code, err := in.resolveCode(v.Code, in.vec, "vessels", "KAP-", &in.sum.Vessels)
			if err != nil {
				return 0, err
			}
			ves := models.Vessel{
				CustomerID:   cusID,
				Code:         code,
				Name:         v.Name,
				VesselNumber: v.VesselNumber,
				VesselType:   v.VesselType,
				Notes:        v.Notes,
				IsActive:     v.IsActive,
			}
			if err := in.tx.Create(&ves).Error; err != nil {
				return 0, err
			}
			in.ve[vesKey(cusID, v.Name)] = ves.ID
			in.sum.Vessels.Added++
			return ves.ID, nil
		}
	}
	return 0, nil
}

func (in *installer) importSph() error {
	for _, d := range in.pkg.SphDocuments {
		num := strings.TrimSpace(d.DocumentNumber)
		if num == "" {
			in.sum.Sph.Skipped++
			continue
		}
		if _, exists := in.sph[low(num)]; exists {
			in.sum.Sph.Skipped++
			continue
		}
		cusName := strings.TrimSpace(d.CustomerName)
		if cusName == "" {
			in.sum.Sph.Skipped++
			continue
		}
		cusID, err := in.ensureCustomerFromPkg(cusName)
		if err != nil {
			return err
		}
		if cusID == 0 {
			in.sum.Sph.Skipped++
			continue
		}
		var vesselID *uint
		vesName := strings.TrimSpace(d.VesselName)
		if vesName != "" {
			vid, verr := in.ensureVesselFromPkg(cusID, cusName, vesName)
			if verr != nil {
				return verr
			}
			if vid != 0 {
				vesselID = &vid
			}
		}
		doc := models.SphDocument{
			DocumentNumber:   num,
			Revision:         d.Revision,
			Date:             d.Date,
			CustomerID:       cusID,
			VesselID:         vesselID,
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
		if err := in.tx.Create(&doc).Error; err != nil {
			return err
		}
		in.sph[low(num)] = doc.ID
		seq := 1
		for _, it := range d.Items {
			item := models.SphItem{
				SphDocumentID:       doc.ID,
				Sequence:            seq,
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
			link := strings.TrimSpace(it.WorkItemName)
			if link != "" {
				if wid, ok := in.wiN[low(link)]; ok {
					item.WorkItemID = &wid
				} else {
					in.sum.SphItemsUnlinked++
				}
			}
			if err := in.tx.Create(&item).Error; err != nil {
				return err
			}
			subSeq := 1
			for _, sub := range it.SubItems {
				s := models.SphSubItem{
					SphItemID:           item.ID,
					Sequence:            subSeq,
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
				}
				if err := in.tx.Create(&s).Error; err != nil {
					return err
				}
				subSeq++
			}
			seq++
		}
		for _, r := range d.Revisions {
			rev := models.SphRevision{
				SphDocumentID:  doc.ID,
				FromDocumentID: nil,
				RevisionNumber: r.RevisionNumber,
				Note:           r.Note,
				CreatedAt:      r.CreatedAt,
			}
			if err := in.tx.Create(&rev).Error; err != nil {
				return err
			}
		}
		in.sum.Sph.Added++
	}
	return nil
}