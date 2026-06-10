package incident

import (
	"fmt"
	"sort"
	"time"

	"github.com/superduperpiyuxh/narrator-ai/backend/internal/attck"
	"github.com/superduperpiyuxh/narrator-ai/backend/internal/database"
)

func GroupEventsIntoIncidents(events []database.Event, windowMinutes int) [][]database.Event {
	if len(events) == 0 {
		return nil
	}

	var groups [][]database.Event
	currentGroup := []database.Event{events[0]}

	for i := 1; i < len(events); i++ {
		prev := events[i-1]
		curr := events[i]

		if curr.SourceIP != prev.SourceIP {
			groups = append(groups, currentGroup)
			currentGroup = []database.Event{curr}
			continue
		}

		prevTime, err1 := time.Parse(time.RFC3339, prev.Timestamp)
		currTime, err2 := time.Parse(time.RFC3339, curr.Timestamp)
		if err1 != nil || err2 != nil {
			currentGroup = append(currentGroup, curr)
			continue
		}

		gap := currTime.Sub(prevTime)
		if gap > time.Duration(windowMinutes)*time.Minute {
			groups = append(groups, currentGroup)
			currentGroup = []database.Event{curr}
		} else {
			currentGroup = append(currentGroup, curr)
		}
	}

	groups = append(groups, currentGroup)
	return groups
}

func BuildIncidentFromGroup(events []database.Event) database.Incident {
	if len(events) == 0 {
		return database.Incident{}
	}

	inc := database.Incident{
		SourceIP:   events[0].SourceIP,
		StartTime:  events[0].Timestamp,
		EndTime:    events[len(events)-1].Timestamp,
		EventCount: len(events),
		Status:     "new",
	}

	users := make(map[string]bool)
	ips := make(map[string]bool)
	hostnames := make(map[string]bool)
	techniqueCounts := make(map[string]int)
	techniqueInfo := make(map[string]attck.Technique)

	for _, e := range events {
		if e.UserName != "" {
			users[e.UserName] = true
		}
		if e.SourceIP != "" {
			ips[e.SourceIP] = true
		}
		if e.DestIP != "" {
			ips[e.DestIP] = true
		}
		if e.Hostname != "" {
			hostnames[e.Hostname] = true
		}

		techniques := attck.MapEventToTechniques(e.EventID, e.EventType, e.CommandLine, e.ProcessName)
		for _, t := range techniques {
			techniqueCounts[t.ID]++
			techniqueInfo[t.ID] = t
		}
	}

	inc.UniqueUsers = mapToSortedSlice(users)
	inc.UniqueIPs = mapToSortedSlice(ips)
	inc.UniqueHostnames = mapToSortedSlice(hostnames)

	var techniques []database.TechniqueRef
	var mitreIDs []string
	var allTactics []string
	tacticSeen := make(map[string]bool)

	for id, count := range techniqueCounts {
		info := techniqueInfo[id]
		techniques = append(techniques, database.TechniqueRef{
			TechniqueID: id,
			Name:        info.Name,
			Tactic:      info.Tactic,
			EventCount:  count,
		})
		mitreIDs = append(mitreIDs, id)
		if !tacticSeen[info.Tactic] {
			tacticSeen[info.Tactic] = true
			allTactics = append(allTactics, info.Tactic)
		}
	}

	inc.Techniques = techniques
	inc.MitreAttackIDs = mitreIDs
	inc.Tactics = allTactics

	eventTypes := make(map[string]bool)
	for _, e := range events {
		eventTypes[e.EventType] = true
	}
	var etSlice []string
	for et := range eventTypes {
		etSlice = append(etSlice, et)
	}
	inc.Severity = CalculateSeverity(len(events), etSlice, mitreIDs)

	if len(techniques) > 0 {
		sort.Slice(techniques, func(i, j int) bool {
			return techniques[i].EventCount > techniques[j].EventCount
		})
		inc.Title = fmt.Sprintf("%s from %s", techniques[0].Name, inc.SourceIP)
	} else {
		inc.Title = fmt.Sprintf("Security Incident from %s", inc.SourceIP)
	}

	inc.Confidence = float64(len(events)) / 50.0
	if inc.Confidence > 1.0 {
		inc.Confidence = 1.0
	}

	return inc
}

func CalculateSeverity(eventCount int, eventTypes, techniques []string) string {
	eventTypeWeights := map[string]int{
		"privilege_escalation": 10,
		"authentication":       3,
		"network_activity":     5,
		"file_activity":        4,
		"file_create":          4,
		"file_delete":          6,
		"process_activity":     5,
		"registry_access":      7,
		"database_query":       4,
		"system":               1,
		"ntlm_auth_success":    4,
	}

	techniqueWeights := map[string]int{
		"T1110":     8,
		"T1021":     9,
		"T1548":     9,
		"T1059":     6,
		"T1053":     7,
		"T1550":     7,
		"T1558":     8,
		"T1070":     5,
		"T1112":     4,
		"T1562":     7,
		"T1003":     9,
		"T1087":     3,
		"T1005":     4,
		"T1048":     8,
		"T1070.004": 5,
	}

	score := eventCount
	for _, et := range eventTypes {
		score += eventTypeWeights[et]
	}
	for _, tid := range techniques {
		score += techniqueWeights[tid]
	}

	switch {
	case score > 50:
		return "critical"
	case score > 30:
		return "high"
	case score > 15:
		return "medium"
	default:
		return "low"
	}
}

func mapToSortedSlice(m map[string]bool) []string {
	var result []string
	for k := range m {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
