package config

import "testing"

func TestJavaFenceNacosDataIDsPreferYml(t *testing.T) {
	if JavaFenceNacosDataIDs[0] != "ebike-fence.yml" {
		t.Fatalf("expected ebike-fence.yml first, got %v", JavaFenceNacosDataIDs)
	}
	if JavaFenceNacosDataID != "ebike-fence.yml" {
		t.Fatalf("JavaFenceNacosDataID = %q", JavaFenceNacosDataID)
	}
}

func TestApplyJavaFenceYAMLCancelAuthTenantIds(t *testing.T) {
	GlobalConfig.Xyy.CancelAuthTenantIds = nil
	yaml := `
xyy:
  remoteLockDistance: 100
  cancelAuthTenantIds:
    - "1007"
`
	if err := ApplyJavaFenceYAML(yaml); err != nil {
		t.Fatal(err)
	}
	if len(GlobalConfig.Xyy.CancelAuthTenantIds) != 1 || GlobalConfig.Xyy.CancelAuthTenantIds[0] != "1007" {
		t.Fatalf("cancelAuthTenantIds = %v", GlobalConfig.Xyy.CancelAuthTenantIds)
	}
	if GlobalConfig.Xyy.RemoteLockDistance == nil || *GlobalConfig.Xyy.RemoteLockDistance != 100 {
		t.Fatalf("remoteLockDistance = %v", GlobalConfig.Xyy.RemoteLockDistance)
	}
}

func TestApplyJavaFenceYAMLDoesNotOverwriteExisting(t *testing.T) {
	GlobalConfig.Xyy.CancelAuthTenantIds = []string{"1008"}
	yaml := `
xyy:
  cancelAuthTenantIds:
    - "1007"
`
	if err := ApplyJavaFenceYAML(yaml); err != nil {
		t.Fatal(err)
	}
	if len(GlobalConfig.Xyy.CancelAuthTenantIds) != 1 || GlobalConfig.Xyy.CancelAuthTenantIds[0] != "1008" {
		t.Fatalf("expected keep 1008, got %v", GlobalConfig.Xyy.CancelAuthTenantIds)
	}
}

func TestMergeXyyIncomingPreservesCancelAuthWhenGoYamlOmits(t *testing.T) {
	GlobalConfig.Xyy.CancelAuthTenantIds = []string{"1007"}
	GlobalConfig.Xyy.RemoteLockDistance = intPtr(100)
	incoming := XyyConfig{RemoteLockDistance: intPtr(200)}
	mergeXyyIncoming(&GlobalConfig.Xyy, incoming)
	if len(GlobalConfig.Xyy.CancelAuthTenantIds) != 1 || GlobalConfig.Xyy.CancelAuthTenantIds[0] != "1007" {
		t.Fatalf("cancelAuthTenantIds cleared: %v", GlobalConfig.Xyy.CancelAuthTenantIds)
	}
	if GlobalConfig.Xyy.RemoteLockDistance == nil || *GlobalConfig.Xyy.RemoteLockDistance != 200 {
		t.Fatalf("remoteLockDistance = %v", GlobalConfig.Xyy.RemoteLockDistance)
	}
}

func intPtr(v int) *int { return &v }
