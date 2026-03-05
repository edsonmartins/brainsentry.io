package service

import (
	"testing"

	"github.com/integraltech/brainsentry/internal/domain"
)

func TestClassifyMemoryType_Personality(t *testing.T) {
	mtype, conf := ClassifyMemoryType("I am a senior Go developer working at a startup", nil, "")
	if mtype != domain.MemoryTypePersonality {
		t.Errorf("expected PERSONALITY, got %s (conf=%.2f)", mtype, conf)
	}
}

func TestClassifyMemoryType_Preference(t *testing.T) {
	mtype, conf := ClassifyMemoryType("I prefer dark mode and always use bun instead of npm", nil, "")
	if mtype != domain.MemoryTypePreference {
		t.Errorf("expected PREFERENCE, got %s (conf=%.2f)", mtype, conf)
	}
}

func TestClassifyMemoryType_Procedural(t *testing.T) {
	mtype, conf := ClassifyMemoryType("How to deploy: Step 1, run the command. Then configure the environment. Finally, restart the service.", nil, "")
	if mtype != domain.MemoryTypeProcedural {
		t.Errorf("expected PROCEDURAL, got %s (conf=%.2f)", mtype, conf)
	}
}

func TestClassifyMemoryType_Task(t *testing.T) {
	mtype, conf := ClassifyMemoryType("TODO: need to fix the authentication bug before the deadline", nil, "")
	if mtype != domain.MemoryTypeTask {
		t.Errorf("expected TASK, got %s (conf=%.2f)", mtype, conf)
	}
}

func TestClassifyMemoryType_Emotion(t *testing.T) {
	mtype, conf := ClassifyMemoryType("I feel frustrated with this deployment process, it's stressful", nil, "")
	if mtype != domain.MemoryTypeEmotion {
		t.Errorf("expected EMOTION, got %s (conf=%.2f)", mtype, conf)
	}
}

func TestClassifyMemoryType_Episodic(t *testing.T) {
	mtype, conf := ClassifyMemoryType("Yesterday I encountered a weird bug during the deployment", nil, "")
	if mtype != domain.MemoryTypeEpisodic {
		t.Errorf("expected EPISODIC, got %s (conf=%.2f)", mtype, conf)
	}
}

func TestClassifyMemoryType_Thread(t *testing.T) {
	mtype, conf := ClassifyMemoryType("Continuing from our conversation earlier, you asked about the API design", nil, "")
	if mtype != domain.MemoryTypeThread {
		t.Errorf("expected THREAD, got %s (conf=%.2f)", mtype, conf)
	}
}

func TestClassifyMemoryType_Semantic(t *testing.T) {
	mtype, conf := ClassifyMemoryType("PostgreSQL is defined as a relational database management system", nil, "")
	if mtype != domain.MemoryTypeSemantic {
		t.Errorf("expected SEMANTIC, got %s (conf=%.2f)", mtype, conf)
	}
}

func TestClassifyMemoryType_EmptyContent(t *testing.T) {
	mtype, conf := ClassifyMemoryType("", nil, "")
	if mtype != domain.MemoryTypeSemantic {
		t.Errorf("expected SEMANTIC for empty, got %s", mtype)
	}
	if conf != 0 {
		t.Errorf("expected 0 confidence for empty, got %.2f", conf)
	}
}

func TestClassifyMemoryType_CategoryBoost(t *testing.T) {
	// ACTION category should boost TASK
	mtype, _ := ClassifyMemoryType("generic text content", nil, domain.CategoryAction)
	if mtype != domain.MemoryTypeTask {
		t.Errorf("expected TASK with ACTION category, got %s", mtype)
	}
}

func TestClassifyMemoryType_TagBoost(t *testing.T) {
	mtype, _ := ClassifyMemoryType("some content about settings", []string{"preference", "config"}, "")
	if mtype != domain.MemoryTypePreference {
		t.Errorf("expected PREFERENCE with preference tags, got %s", mtype)
	}
}

func TestClassifyMemoryType_ReturnsConfidence(t *testing.T) {
	_, conf := ClassifyMemoryType("I prefer using Go for backend development and I like typed languages", nil, "")
	if conf <= 0 || conf > 1 {
		t.Errorf("expected confidence between 0 and 1, got %.2f", conf)
	}
}
