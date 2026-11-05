package catalog

import "testing"

func TestBuildResourceTreeUnifiesVMAndPayableAddons(t *testing.T) {
	vmID := int64(101)
	flat := []PayableResourceNode{
		{ResourceID: vmID, Kind: KindVM, Name: "web-01"},
		{ResourceID: 102, ParentResourceID: &vmID, Kind: KindIPv4, Name: "ipv4-198.51.100.10"},
		{ResourceID: 103, ParentResourceID: &vmID, Kind: KindBlockSSD, Name: "vol-data"},
	}
	roots := BuildResourceTree(flat)
	if len(roots) != 1 {
		t.Fatalf("expected 1 root VM resource, got %d", len(roots))
	}
	if len(roots[0].Children) != 2 {
		t.Fatalf("expected 2 attached child payable items (IPv4 + SSD), got %d", len(roots[0].Children))
	}
}
