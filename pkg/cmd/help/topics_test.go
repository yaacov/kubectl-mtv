package help

import "testing"

func TestGetTopic_TSL(t *testing.T) {
	topic := GetTopic("tsl")
	if topic == nil {
		t.Fatal("expected TSL topic, got nil")
	}
	if topic.Name != "tsl" {
		t.Errorf("expected name 'tsl', got %q", topic.Name)
	}
	if topic.Short == "" {
		t.Error("expected non-empty short description")
	}
	if topic.Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestGetTopic_KARL(t *testing.T) {
	topic := GetTopic("karl")
	if topic == nil {
		t.Fatal("expected KARL topic, got nil")
	}
	if topic.Name != "karl" {
		t.Errorf("expected name 'karl', got %q", topic.Name)
	}
	if topic.Short == "" {
		t.Error("expected non-empty short description")
	}
	if topic.Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestGetTopic_CaseInsensitive(t *testing.T) {
	tests := []string{"TSL", "Tsl", "tSl", "tsl"}
	for _, name := range tests {
		topic := GetTopic(name)
		if topic == nil {
			t.Errorf("GetTopic(%q) returned nil", name)
		}
	}
}

func TestGetTopic_NotFound(t *testing.T) {
	topic := GetTopic("nonexistent")
	if topic != nil {
		t.Errorf("expected nil for nonexistent topic, got %+v", topic)
	}
}

func TestGetTopic_ReturnsCopy(t *testing.T) {
	topic1 := GetTopic("tsl")
	topic2 := GetTopic("tsl")

	if topic1 == topic2 {
		t.Error("GetTopic should return a copy, not the same pointer")
	}

	// Modify one and verify the other is unchanged
	topic1.Short = "modified"
	topic2Again := GetTopic("tsl")
	if topic2Again.Short == "modified" {
		t.Error("modifying returned topic should not affect the registry")
	}
}

func TestListTopics(t *testing.T) {
	topics := ListTopics()

	if len(topics) < 2 {
		t.Fatalf("expected at least 2 topics, got %d", len(topics))
	}

	names := map[string]bool{}
	for _, topic := range topics {
		names[topic.Name] = true
		if topic.Short == "" {
			t.Errorf("topic %q has empty short description", topic.Name)
		}
		if topic.Content == "" {
			t.Errorf("topic %q has empty content", topic.Name)
		}
	}

	if !names["tsl"] {
		t.Error("expected 'tsl' topic in list")
	}
	if !names["karl"] {
		t.Error("expected 'karl' topic in list")
	}
}

func TestListTopics_ReturnsCopy(t *testing.T) {
	topics1 := ListTopics()
	topics2 := ListTopics()

	// Modify the first copy and verify the second is unaffected
	topics1[0].Name = "modified"
	if topics2[0].Name == "modified" {
		t.Error("ListTopics should return a copy of the registry")
	}
}

func TestTopicContent_TSL_HasExpectedSections(t *testing.T) {
	topic := GetTopic("tsl")
	if topic == nil {
		t.Fatal("expected TSL topic")
	}

	sections := []string{
		"Query Structure:",
		"Operators:",
		"VM Fields by Provider",
		"vSphere:",
		"oVirt / RHV:",
		"OpenStack:",
		"EC2 (PascalCase):",
		"Examples",
	}
	for _, section := range sections {
		if !containsString(topic.Content, section) {
			t.Errorf("TSL content missing section: %q", section)
		}
	}
}

func TestTopicContent_KARL_HasExpectedSections(t *testing.T) {
	topic := GetTopic("karl")
	if topic == nil {
		t.Fatal("expected KARL topic")
	}

	sections := []string{
		"Rule Types:",
		"Topology Keys:",
		"Label Selectors:",
		"Examples",
	}
	for _, section := range sections {
		if !containsString(topic.Content, section) {
			t.Errorf("KARL content missing section: %q", section)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && contains(s, substr)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
