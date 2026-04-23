package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// ProvenanceExporter projects BrainSentry's audit log + decisions into the
// W3C PROV-O vocabulary so compliance teams can ingest provenance alongside
// other semantic-web tooling. Turtle and JSON-LD serialisations share the same
// graph model; the exporter emits either on demand.
type ProvenanceExporter struct {
	auditRepo    *postgres.AuditRepository
	decisionRepo *postgres.DecisionRepository
	baseURI      string
}

// NewProvenanceExporter wires the exporter.
func NewProvenanceExporter(audit *postgres.AuditRepository, decisions *postgres.DecisionRepository, baseURI string) *ProvenanceExporter {
	if baseURI == "" {
		baseURI = "https://brainsentry.local/prov"
	}
	return &ProvenanceExporter{auditRepo: audit, decisionRepo: decisions, baseURI: strings.TrimRight(baseURI, "/")}
}

// ExportOptions controls the exported window and format.
type ExportOptions struct {
	Since  *time.Time
	Until  *time.Time
	Limit  int
	Format string // "turtle" or "jsonld"; defaults to turtle
}

// ExportPROV writes a PROV-O document for the tenant into w.
func (e *ProvenanceExporter) ExportPROV(ctx context.Context, w io.Writer, opts ExportOptions) error {
	if opts.Limit <= 0 {
		opts.Limit = 500
	}
	format := strings.ToLower(opts.Format)
	if format == "" {
		format = "turtle"
	}

	logs, err := e.auditRepo.ListByTenant(ctx, opts.Limit)
	if err != nil {
		return fmt.Errorf("fetching audit logs: %w", err)
	}

	decisions, err := e.decisionRepo.List(ctx, postgres.DecisionFilter{Limit: opts.Limit})
	if err != nil {
		return fmt.Errorf("fetching decisions: %w", err)
	}

	switch format {
	case "jsonld":
		return e.writeJSONLD(w, logs, decisions, tenant.FromContext(ctx))
	case "turtle", "ttl":
		return e.writeTurtle(w, logs, decisions, tenant.FromContext(ctx))
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func (e *ProvenanceExporter) writeTurtle(w io.Writer, logs []domain.AuditLog, decisions []*domain.Decision, tenantID string) error {
	var b strings.Builder
	b.WriteString("@prefix prov: <http://www.w3.org/ns/prov#> .\n")
	b.WriteString("@prefix xsd:  <http://www.w3.org/2001/XMLSchema#> .\n")
	b.WriteString("@prefix bs:   <" + e.baseURI + "#> .\n")
	b.WriteString("@prefix tenant: <" + e.baseURI + "/tenant/> .\n\n")

	// Tenant as prov:Organization
	fmt.Fprintf(&b, "tenant:%s a prov:Agent, prov:Organization ;\n    prov:label %q .\n\n",
		safeID(tenantID), "tenant:"+tenantID)

	for _, log := range logs {
		actID := "bs:activity-" + safeID(log.ID)
		fmt.Fprintf(&b, "%s a prov:Activity ;\n", actID)
		fmt.Fprintf(&b, "    prov:startedAtTime %q^^xsd:dateTime ;\n", log.Timestamp.UTC().Format(time.RFC3339))
		fmt.Fprintf(&b, "    prov:endedAtTime %q^^xsd:dateTime ;\n", log.Timestamp.UTC().Format(time.RFC3339))
		fmt.Fprintf(&b, "    prov:wasAssociatedWith tenant:%s ;\n", safeID(tenantID))
		fmt.Fprintf(&b, "    prov:label %q ", log.EventType)

		if log.UserID != "" {
			fmt.Fprintf(&b, ";\n    prov:wasStartedBy bs:user-%s ", safeID(log.UserID))
		}
		if log.Outcome != "" {
			fmt.Fprintf(&b, ";\n    bs:outcome %q ", log.Outcome)
		}
		b.WriteString(".\n")

		for _, mid := range log.MemoriesCreated {
			fmt.Fprintf(&b, "bs:memory-%s a prov:Entity ;\n    prov:wasGeneratedBy %s .\n",
				safeID(mid), actID)
		}
		for _, mid := range log.MemoriesModified {
			fmt.Fprintf(&b, "bs:memory-%s a prov:Entity ;\n    prov:wasRevisionOf %s .\n",
				safeID(mid), actID)
		}
		for _, mid := range log.MemoriesAccessed {
			fmt.Fprintf(&b, "%s prov:used bs:memory-%s .\n", actID, safeID(mid))
		}
		b.WriteString("\n")
	}

	for _, d := range decisions {
		decID := "bs:decision-" + safeID(d.ID)
		fmt.Fprintf(&b, "%s a prov:Entity, bs:Decision ;\n", decID)
		fmt.Fprintf(&b, "    prov:generatedAtTime %q^^xsd:dateTime ;\n", d.CreatedAt.UTC().Format(time.RFC3339))
		fmt.Fprintf(&b, "    bs:category %q ;\n", d.Category)
		fmt.Fprintf(&b, "    bs:outcome %q ;\n", d.Outcome)
		fmt.Fprintf(&b, "    bs:confidence %q^^xsd:decimal ", fmt.Sprintf("%.4f", d.Confidence))

		if d.AgentID != "" {
			fmt.Fprintf(&b, ";\n    prov:wasAttributedTo bs:agent-%s ", safeID(d.AgentID))
		}
		if d.ParentDecisionID != "" {
			fmt.Fprintf(&b, ";\n    prov:wasDerivedFrom bs:decision-%s ", safeID(d.ParentDecisionID))
		}
		for _, mid := range d.MemoryIDs {
			fmt.Fprintf(&b, ";\n    bs:citedMemory bs:memory-%s ", safeID(mid))
		}
		for _, eid := range d.EntityIDs {
			fmt.Fprintf(&b, ";\n    bs:relatedEntity bs:entity-%s ", safeID(eid))
		}
		b.WriteString(".\n\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

type jsonldNode struct {
	ID      string   `json:"@id"`
	Type    []string `json:"@type"`
	Label   string   `json:"prov:label,omitempty"`
	Started string   `json:"prov:startedAtTime,omitempty"`
	Ended   string   `json:"prov:endedAtTime,omitempty"`
	Used    []string `json:"prov:used,omitempty"`
	Gen     []string `json:"prov:wasGeneratedBy,omitempty"`
	Derived string   `json:"prov:wasDerivedFrom,omitempty"`
	AssocWith string `json:"prov:wasAssociatedWith,omitempty"`
	Outcome string   `json:"bs:outcome,omitempty"`
	Category string  `json:"bs:category,omitempty"`
	Confidence *float64 `json:"bs:confidence,omitempty"`
	Memories []string `json:"bs:citedMemory,omitempty"`
	Entities []string `json:"bs:relatedEntity,omitempty"`
}

func (e *ProvenanceExporter) writeJSONLD(w io.Writer, logs []domain.AuditLog, decisions []*domain.Decision, tenantID string) error {
	doc := map[string]any{
		"@context": map[string]any{
			"prov":   "http://www.w3.org/ns/prov#",
			"xsd":    "http://www.w3.org/2001/XMLSchema#",
			"bs":     e.baseURI + "#",
			"tenant": e.baseURI + "/tenant/",
		},
	}

	graph := []jsonldNode{
		{
			ID:    "tenant:" + safeID(tenantID),
			Type:  []string{"prov:Agent", "prov:Organization"},
			Label: "tenant:" + tenantID,
		},
	}

	for _, log := range logs {
		used := make([]string, 0, len(log.MemoriesAccessed))
		for _, mid := range log.MemoriesAccessed {
			used = append(used, "bs:memory-"+safeID(mid))
		}
		graph = append(graph, jsonldNode{
			ID:        "bs:activity-" + safeID(log.ID),
			Type:      []string{"prov:Activity"},
			Label:     log.EventType,
			Started:   log.Timestamp.UTC().Format(time.RFC3339),
			Ended:     log.Timestamp.UTC().Format(time.RFC3339),
			AssocWith: "tenant:" + safeID(tenantID),
			Used:      used,
			Outcome:   log.Outcome,
		})
		for _, mid := range log.MemoriesCreated {
			graph = append(graph, jsonldNode{
				ID:   "bs:memory-" + safeID(mid),
				Type: []string{"prov:Entity"},
				Gen:  []string{"bs:activity-" + safeID(log.ID)},
			})
		}
	}

	for _, d := range decisions {
		memIDs := make([]string, 0, len(d.MemoryIDs))
		for _, mid := range d.MemoryIDs {
			memIDs = append(memIDs, "bs:memory-"+safeID(mid))
		}
		entIDs := make([]string, 0, len(d.EntityIDs))
		for _, eid := range d.EntityIDs {
			entIDs = append(entIDs, "bs:entity-"+safeID(eid))
		}
		conf := d.Confidence
		node := jsonldNode{
			ID:         "bs:decision-" + safeID(d.ID),
			Type:       []string{"prov:Entity", "bs:Decision"},
			Started:    d.CreatedAt.UTC().Format(time.RFC3339),
			Category:   d.Category,
			Outcome:    string(d.Outcome),
			Confidence: &conf,
			Memories:   memIDs,
			Entities:   entIDs,
		}
		if d.ParentDecisionID != "" {
			node.Derived = "bs:decision-" + safeID(d.ParentDecisionID)
		}
		graph = append(graph, node)
	}

	doc["@graph"] = graph

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

// safeID escapes characters that break IRI syntax.
func safeID(s string) string {
	r := strings.NewReplacer(" ", "_", "\"", "", "<", "", ">", "", "#", "")
	return r.Replace(s)
}
