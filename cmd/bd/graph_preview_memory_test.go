package main

import "testing"

func TestGraphPreviewMemoryReplacementPresence(t *testing.T) {
	for _, properties := range []map[string]any{
		{"title": "", "body": ""}, {"title": " Plan ", "body": "雪\n"},
	} {
		got, err := graphPreviewMemoryProperties(properties)
		if err != nil || got.Title != properties["title"] || got.Body != properties["body"] {
			t.Fatalf("changed explicit strings: %+v %v", got, err)
		}
	}
	for _, properties := range []map[string]any{
		nil, {}, {"title": "Plan"}, {"body": "Text"}, {"title": nil, "body": ""},
		{"title": "", "body": false}, {"title": "", "body": "", "metadata": map[string]any{}},
	} {
		if _, err := graphPreviewMemoryProperties(properties); err == nil {
			t.Fatalf("admitted incomplete or invalid replacement: %v", properties)
		}
	}
}
