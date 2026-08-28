package collaboration

import (
	"testing"
)

func TestMasterDataPackageSerializeDeserialize(t *testing.T) {
	pkg := &MasterDataPackage{
		Metadata: MasterPackageMetadata{
			PackageID:      "pkg-1",
			SenderID:       "host",
			SenderName:     "Admin",
			RoomID:         "room-1",
			SchemaVersion:  PackageSchemaVersion,
			PackageVersion: "1",
			SourceVersion:  2,
		},
		Data: MasterPackageData{
			Categories: []PackageCategory{{Code: "EL", Name: "Electrical"}},
			WorkItems:  []PackageWorkItem{{Code: "EL-001", Name: "Repair PLC", CategoryCode: "EL"}},
			WorkSubItems: []PackageWorkSubItem{{
				Code: "sub-1", WorkItemCode: "EL-001", Name: "Inspection", Sequence: 1,
			}},
			Materials: []PackageMaterial{{Code: "M-1", Name: "Cable", DefaultPrice: 50000}},
		},
	}

	raw, err := pkg.Serialize()
	if err != nil {
		t.Fatalf("Serialize error: %v", err)
	}
	back, err := DeserializeMasterDataPackage(raw)
	if err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if back.Metadata.PackageID != "pkg-1" {
		t.Errorf("PackageID = %q, mau pkg-1", back.Metadata.PackageID)
	}
	if len(back.Data.WorkItems) != 1 || back.Data.WorkItems[0].Code != "EL-001" {
		t.Errorf("WorkItems salah: %+v", back.Data.WorkItems)
	}
	if len(back.Data.Categories) != 1 || back.Data.Categories[0].Code != "EL" {
		t.Errorf("Categories salah: %+v", back.Data.Categories)
	}
}

func TestMasterDataPackageChecksum(t *testing.T) {
	pkg := &MasterDataPackage{
		Metadata: MasterPackageMetadata{PackageID: "pkg-2", SenderID: "s", RoomID: "r"},
		Data: MasterPackageData{
			Categories: []PackageCategory{{Code: "EL", Name: "Electrical"}},
		},
	}
	sum, err := pkg.ComputeChecksum()
	if err != nil {
		t.Fatalf("ComputeChecksum error: %v", err)
	}
	if len(sum) != 64 { // sha256 hex
		t.Fatalf("checksum panjang = %d, mau 64", len(sum))
	}

	// Setelah checksum disimpan, Verify harus true.
	pkg.Checksum = sum
	ok, err := pkg.VerifyChecksum()
	if err != nil {
		t.Fatalf("VerifyChecksum error: %v", err)
	}
	if !ok {
		t.Error("VerifyChecksum harus true untuk checksum yang benar")
	}

	// Data rusak → checksum tidak cocok.
	pkg.Data.Materials = []PackageMaterial{{Code: "X", Name: "Unsafe"}}
	tampered, _ := pkg.ComputeChecksum()
	if tampered == sum {
		t.Fatal("checksum seharusnya berubah ketika data berubah")
	}
	ok, err = pkg.VerifyChecksum()
	if err != nil {
		t.Fatalf("VerifyChecksum error setelah tamper: %v", err)
	}
	if ok {
		t.Error("VerifyChecksum harus false untuk checksum yang tidak cocok")
	}
}

func TestMasterPackageChecksumStableAcrossReencode(t *testing.T) {
	pkg := &MasterDataPackage{
		Metadata: MasterPackageMetadata{PackageID: "pkg-3", SenderID: "s", RoomID: "r"},
		Data:     MasterPackageData{WorkItems: []PackageWorkItem{{Code: "A", Name: "B"}}},
	}
	sum1, _ := pkg.ComputeChecksum()
	sum2, _ := pkg.ComputeChecksum()
	if sum1 != sum2 {
		t.Errorf("checksum tidak stabil: %q != %q", sum1, sum2)
	}
}
