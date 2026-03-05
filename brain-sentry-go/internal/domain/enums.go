package domain

// ImportanceLevel indicates how strongly memories should be followed.
type ImportanceLevel string

const (
	ImportanceCritical  ImportanceLevel = "CRITICAL"
	ImportanceImportant ImportanceLevel = "IMPORTANT"
	ImportanceMinor     ImportanceLevel = "MINOR"
)

// MemoryCategory represents universal memory categories.
type MemoryCategory string

const (
	CategoryInsight      MemoryCategory = "INSIGHT"
	CategoryDecision     MemoryCategory = "DECISION"
	CategoryWarning      MemoryCategory = "WARNING"
	CategoryKnowledge    MemoryCategory = "KNOWLEDGE"
	CategoryAction       MemoryCategory = "ACTION"
	CategoryContext      MemoryCategory = "CONTEXT"
	CategoryReference    MemoryCategory = "REFERENCE"
	CategoryPattern      MemoryCategory = "PATTERN"
	CategoryAntipattern  MemoryCategory = "ANTIPATTERN"
	CategoryDomain       MemoryCategory = "DOMAIN"
	CategoryBug          MemoryCategory = "BUG"
	CategoryOptimization MemoryCategory = "OPTIMIZATION"
	CategoryIntegration  MemoryCategory = "INTEGRATION"
)

// ValidationStatus represents validation states for memories.
type ValidationStatus string

const (
	ValidationApproved ValidationStatus = "APPROVED"
	ValidationPending  ValidationStatus = "PENDING"
	ValidationFlagged  ValidationStatus = "FLAGGED"
	ValidationRejected ValidationStatus = "REJECTED"
)

// RelationshipType represents memory relationship types.
type RelationshipType string

const (
	RelUsedWith      RelationshipType = "USED_WITH"
	RelConflictsWith RelationshipType = "CONFLICTS_WITH"
	RelSupersedes    RelationshipType = "SUPERSEDES"
	RelRelatedTo     RelationshipType = "RELATED_TO"
	RelRequires      RelationshipType = "REQUIRES"
	RelPartOf        RelationshipType = "PART_OF"
)

// MemoryType represents the cognitive type of a memory.
type MemoryType string

const (
	MemoryTypeSemantic    MemoryType = "SEMANTIC"    // Facts, concepts, knowledge
	MemoryTypeEpisodic    MemoryType = "EPISODIC"    // Events, experiences, sessions
	MemoryTypeProcedural  MemoryType = "PROCEDURAL"  // How-to, procedures, patterns
	MemoryTypeAssociative MemoryType = "ASSOCIATIVE" // Links between concepts
	MemoryTypePersonality MemoryType = "PERSONALITY" // Stable user traits, preferences, identity
	MemoryTypePreference  MemoryType = "PREFERENCE"  // User preferences, likes/dislikes
	MemoryTypeThread      MemoryType = "THREAD"      // Conversational thread context, ephemeral
	MemoryTypeTask        MemoryType = "TASK"         // Ongoing tasks, TODOs, action items
	MemoryTypeEmotion     MemoryType = "EMOTION"      // Emotional reactions, sentiments
)

// CorrectionStatus represents the status of a memory correction.
type CorrectionStatus string

const (
	CorrectionPending  CorrectionStatus = "PENDING"
	CorrectionApproved CorrectionStatus = "APPROVED"
	CorrectionRejected CorrectionStatus = "REJECTED"
	CorrectionApplied  CorrectionStatus = "APPLIED"
)

// NoteType represents types of notes.
type NoteType string

const (
	NoteInsight      NoteType = "INSIGHT"
	NoteHindsight    NoteType = "HINDSIGHT"
	NotePattern      NoteType = "PATTERN"
	NoteAntipattern  NoteType = "ANTIPATTERN"
	NoteArchitecture NoteType = "ARCHITECTURE"
	NoteIntegration  NoteType = "INTEGRATION"
)

// NoteCategory represents note scope.
type NoteCategory string

const (
	NoteCategoryProjectSpecific NoteCategory = "PROJECT_SPECIFIC"
	NoteCategoryShared          NoteCategory = "SHARED"
	NoteCategoryGeneric         NoteCategory = "GENERIC"
)

// ObservationType represents typed session observations.
type ObservationType string

const (
	ObservationDecision  ObservationType = "DECISION"
	ObservationBugfix    ObservationType = "BUGFIX"
	ObservationFeature   ObservationType = "FEATURE"
	ObservationRefactor  ObservationType = "REFACTOR"
	ObservationDiscovery ObservationType = "DISCOVERY"
	ObservationChange    ObservationType = "CHANGE"
)

// NoteSeverity represents severity levels for notes.
type NoteSeverity string

const (
	SeverityCritical NoteSeverity = "CRITICAL"
	SeverityHigh     NoteSeverity = "HIGH"
	SeverityMedium   NoteSeverity = "MEDIUM"
	SeverityLow      NoteSeverity = "LOW"
)
