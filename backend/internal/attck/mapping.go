package attck

import (
	"regexp"
	"strings"
)

type Technique struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Tactic string `json:"tactic"`
	URL    string `json:"url"`
}

var eventIDToTechniques = map[string][]Technique{
	"4625": {{ID: "T1110", Name: "Brute Force", Tactic: "credential-access", URL: "https://attack.mitre.org/techniques/T1110/"}},
	"4634": {{ID: "T1110", Name: "Brute Force", Tactic: "credential-access", URL: "https://attack.mitre.org/techniques/T1110/"}},
	"4648": {{ID: "T1550", Name: "Use Alternate Authentication Material", Tactic: "credential-access", URL: "https://attack.mitre.org/techniques/T1550/"}},
	"4672": {{ID: "T1548", Name: "Abuse Elevation Control Mechanism", Tactic: "privilege-escalation", URL: "https://attack.mitre.org/techniques/T1548/"}},
	"4688": {{ID: "T1059", Name: "Command and Scripting Interpreter", Tactic: "execution", URL: "https://attack.mitre.org/techniques/T1059/"}},
	"4698": {{ID: "T1053", Name: "Scheduled Task/Job", Tactic: "persistence", URL: "https://attack.mitre.org/techniques/T1053/"}},
	"4663": {{ID: "T1070", Name: "Indicator Removal on Host", Tactic: "defense-evasion", URL: "https://attack.mitre.org/techniques/T1070/"}},
	"4660": {{ID: "T1070.004", Name: "File Deletion", Tactic: "defense-evasion", URL: "https://attack.mitre.org/techniques/T1070/004/"}},
	"4657": {{ID: "T1112", Name: "Modify Registry", Tactic: "defense-evasion", URL: "https://attack.mitre.org/techniques/T1112/"}},
	"4768": {{ID: "T1558", Name: "Steal or Forge Kerberos Tickets", Tactic: "credential-access", URL: "https://attack.mitre.org/techniques/T1558/"}},
	"4769": {{ID: "T1558", Name: "Steal or Forge Kerberos Tickets", Tactic: "credential-access", URL: "https://attack.mitre.org/techniques/T1558/"}},
	"5156": {{ID: "T1021", Name: "Remote Services", Tactic: "lateral-movement", URL: "https://attack.mitre.org/techniques/T1021/"}},
	"1074": {{ID: "T1562", Name: "Impair Defenses", Tactic: "defense-evasion", URL: "https://attack.mitre.org/techniques/T1562/"}},
	"7045": {{ID: "T1053", Name: "Scheduled Task/Job", Tactic: "persistence", URL: "https://attack.mitre.org/techniques/T1053/"}},
	"7036": {{ID: "T1562", Name: "Impair Defenses", Tactic: "defense-evasion", URL: "https://attack.mitre.org/techniques/T1562/"}},
	"4626": {{ID: "T1087", Name: "Account Discovery", Tactic: "discovery", URL: "https://attack.mitre.org/techniques/T1087/"}},
}

type patternMatch struct {
	pattern    *regexp.Regexp
	techniques []Technique
}

var patternMatchers = []patternMatch{
	{
		pattern: regexp.MustCompile(`(?i)mimikatz|kerberos.*ticket|hashdump|lsass`),
		techniques: []Technique{{ID: "T1003", Name: "OS Credential Dumping", Tactic: "credential-access", URL: "https://attack.mitre.org/techniques/T1003/"}},
	},
	{
		pattern: regexp.MustCompile(`(?i)net\s+user|net\s+group|nltest|dsquery`),
		techniques: []Technique{{ID: "T1087", Name: "Account Discovery", Tactic: "discovery", URL: "https://attack.mitre.org/techniques/T1087/"}},
	},
	{
		pattern: regexp.MustCompile(`(?i)copy.*\\\\|xcopy|robocopy.*\/mir`),
		techniques: []Technique{{ID: "T1005", Name: "Data from Local System", Tactic: "collection", URL: "https://attack.mitre.org/techniques/T1005/"}},
	},
	{
		pattern: regexp.MustCompile(`(?i)curl.*upload|wget.*post|tftp.*put`),
		techniques: []Technique{{ID: "T1048", Name: "Exfiltration Over Alternative Protocol", Tactic: "exfiltration", URL: "https://attack.mitre.org/techniques/T1048/"}},
	},
}

func MapEventByEventID(eventID string) []Technique {
	return eventIDToTechniques[eventID]
}

func MapEventByPatterns(commandLine, processName string) []Technique {
	combined := commandLine + " " + processName
	var results []Technique
	for _, pm := range patternMatchers {
		if pm.pattern.MatchString(combined) {
			results = append(results, pm.techniques...)
		}
	}
	return results
}

func MapEventToTechniques(eventID, eventType, commandLine, processName string) []Technique {
	seen := make(map[string]bool)
	var results []Technique

	for _, t := range MapEventByEventID(eventID) {
		if !seen[t.ID] {
			seen[t.ID] = true
			results = append(results, t)
		}
	}

	for _, t := range MapEventByPatterns(commandLine, processName) {
		if !seen[t.ID] {
			seen[t.ID] = true
			results = append(results, t)
		}
	}

	return results
}

func AllTechniques() []Technique {
	seen := make(map[string]bool)
	var results []Technique

	for _, techniques := range eventIDToTechniques {
		for _, t := range techniques {
			if !seen[t.ID] {
				seen[t.ID] = true
				results = append(results, t)
			}
		}
	}

	for _, pm := range patternMatchers {
		for _, t := range pm.techniques {
			if !seen[t.ID] {
				seen[t.ID] = true
				results = append(results, t)
			}
		}
	}

	return results
}

func TechniqueIDs(techniques []Technique) string {
	var ids []string
	for _, t := range techniques {
		ids = append(ids, t.ID)
	}
	return strings.Join(ids, ",")
}

func TechniqueNames(techniques []Technique) string {
	var names []string
	for _, t := range techniques {
		names = append(names, t.Name)
	}
	return strings.Join(names, ",")
}

func TacticNames(techniques []Technique) string {
	seen := make(map[string]bool)
	var tactics []string
	for _, t := range techniques {
		if !seen[t.Tactic] {
			seen[t.Tactic] = true
			tactics = append(tactics, t.Tactic)
		}
	}
	return strings.Join(tactics, ",")
}
