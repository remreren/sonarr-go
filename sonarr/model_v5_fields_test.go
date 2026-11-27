package sonarr

import (
	"encoding/json"
	"strings"
	"testing"
)

// Sonarr v5 fields: userRejectedExtensions (media management; v5 answers 500 when it's missing)
// and the per-quality sizes on quality profile items (v5 moved them off the quality definitions).

func TestMediaManagementUserRejectedExtensionsRoundTrip(t *testing.T) {
	in := NewMediaManagementConfigResource()
	in.SetUserRejectedExtensions("exe,lnk")

	out, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"userRejectedExtensions":"exe,lnk"`) {
		t.Fatalf("not serialized: %s", out)
	}

	var back MediaManagementConfigResource
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatal(err)
	}
	if !back.HasUserRejectedExtensions() || back.GetUserRejectedExtensions() != "exe,lnk" {
		t.Fatalf("not read back: %+v", back.UserRejectedExtensions)
	}
}

func TestMediaManagementUserRejectedExtensionsEmptyIsSent(t *testing.T) {
	in := NewMediaManagementConfigResource()
	in.SetUserRejectedExtensions("")

	out, _ := json.Marshal(in)
	if !strings.Contains(string(out), `"userRejectedExtensions":""`) {
		t.Fatalf("an empty value must still be sent (v5 needs the field): %s", out)
	}
}

func TestMediaManagementUserRejectedExtensionsAbsentOnV4(t *testing.T) {
	var v4 MediaManagementConfigResource
	if err := json.Unmarshal([]byte(`{"id":1,"extraFileExtensions":"srt"}`), &v4); err != nil {
		t.Fatal(err)
	}
	if v4.HasUserRejectedExtensions() || v4.GetUserRejectedExtensions() != "" {
		t.Fatal("a v4 answer has no userRejectedExtensions")
	}
	out, _ := json.Marshal(v4)
	if strings.Contains(string(out), "userRejectedExtensions") {
		t.Fatalf("unset must not be sent: %s", out)
	}
}

func TestQualityProfileItemSizesRoundTrip(t *testing.T) {
	src := `{"quality":{"id":3,"name":"WEBDL-1080p"},"items":[],"allowed":true,"minSize":4,"maxSize":85.5,"preferredSize":60}`

	var item QualityProfileQualityItemResource
	if err := json.Unmarshal([]byte(src), &item); err != nil {
		t.Fatal(err)
	}
	if item.GetMinSize() != 4 || item.GetMaxSize() != 85.5 || item.GetPreferredSize() != 60 {
		t.Fatalf("sizes not read: %v %v %v", item.GetMinSize(), item.GetMaxSize(), item.GetPreferredSize())
	}

	out, _ := json.Marshal(item)
	for _, want := range []string{`"minSize":4`, `"maxSize":85.5`, `"preferredSize":60`} {
		if !strings.Contains(string(out), want) {
			t.Fatalf("%s not serialized: %s", want, out)
		}
	}
}

func TestQualityProfileItemSizesNullAndAbsent(t *testing.T) {
	var group QualityProfileQualityItemResource // v5 groups: sizes are null
	if err := json.Unmarshal([]byte(`{"name":"WEB 1080p","items":[],"allowed":true,"minSize":null}`), &group); err != nil {
		t.Fatal(err)
	}
	if v, ok := group.GetMinSizeOk(); !ok || v != nil {
		t.Fatal("explicit null: set, but nil")
	}
	if group.HasMaxSize() {
		t.Fatal("absent (v4, or a v5 group): not set")
	}

	plain := NewQualityProfileQualityItemResource()
	out, _ := json.Marshal(plain)
	if strings.Contains(string(out), "Size") {
		t.Fatalf("unset sizes must not be sent: %s", out)
	}
}
