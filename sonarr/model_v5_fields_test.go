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

// roundTrip marshals v, checks every want substring, and unmarshals it into back.
func roundTrip(t *testing.T, v interface{}, back interface{}, want ...string) {
	t.Helper()

	out, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range want {
		if !strings.Contains(string(out), w) {
			t.Fatalf("%s not serialized: %s", w, out)
		}
	}
	if err := json.Unmarshal(out, back); err != nil {
		t.Fatal(err)
	}
}

func TestHostConfigOidcFieldsRoundTrip(t *testing.T) {
	in := NewHostConfigResource()
	in.SetOidcAuthority("https://id.example")
	in.SetOidcClientId("sonarr")
	in.SetOidcClientSecret("s3cret")
	in.SetOidcScopes("openid profile email")
	in.SetOidcUserIdentifier("")
	in.SetSslKeyPath("")
	in.SetAuthenticationMethod(AUTHENTICATIONTYPE_OIDC)
	in.SetAuthenticationRequired(AUTHENTICATIONREQUIREDTYPE_DISABLED_FOR_LOCALHOST)

	var back HostConfigResource
	roundTrip(t, in, &back, `"oidcAuthority":"https://id.example"`, `"oidcUserIdentifier":""`, `"sslKeyPath":""`,
		`"authenticationMethod":"oidc"`, `"authenticationRequired":"disabledForLocalhost"`)
	if back.GetOidcClientSecret() != "s3cret" || !back.HasSslKeyPath() || back.GetAuthenticationMethod() != AUTHENTICATIONTYPE_OIDC {
		t.Fatalf("not read back: %+v", back)
	}
}

func TestMediaManagementSeasonPackFieldsRoundTrip(t *testing.T) {
	in := NewMediaManagementConfigResource()
	in.SetSeasonPackUpgrade(SEASONPACKUPGRADETYPE_THRESHOLD)
	in.SetSeasonPackUpgradeThreshold(75)
	in.SetSkipFreeSpaceCheckWhenGrabbing(true)

	var back MediaManagementConfigResource
	roundTrip(t, in, &back, `"seasonPackUpgrade":"threshold"`, `"seasonPackUpgradeThreshold":75`, `"skipFreeSpaceCheckWhenGrabbing":true`)
	if back.GetSeasonPackUpgrade() != SEASONPACKUPGRADETYPE_THRESHOLD || back.GetSeasonPackUpgradeThreshold() != 75 || !back.GetSkipFreeSpaceCheckWhenGrabbing() {
		t.Fatalf("not read back: %+v", back)
	}

	var bad MediaManagementConfigResource
	if err := json.Unmarshal([]byte(`{"seasonPackUpgrade":"sometimes"}`), &bad); err == nil {
		t.Fatal("an unknown seasonPackUpgrade must be rejected")
	}
}

func TestReleaseProfileAirDateFieldsRoundTrip(t *testing.T) {
	in := NewReleaseProfileResource()
	in.SetAirDateRestriction(true)
	in.SetAirDateGracePeriod(3)
	in.SetAllowSeasonPackWithoutAllEpisodesAired(true)
	in.SetExcludedTags([]int32{4, 7})

	var back ReleaseProfileResource
	roundTrip(t, in, &back, `"airDateRestriction":true`, `"airDateGracePeriod":3`, `"allowSeasonPackWithoutAllEpisodesAired":true`, `"excludedTags":[4,7]`)
	if !back.GetAirDateRestriction() || back.GetAirDateGracePeriod() != 3 || len(back.GetExcludedTags()) != 2 {
		t.Fatalf("not read back: %+v", back)
	}
}

func TestImportListAndSeriesLanguageRoundTrip(t *testing.T) {
	lang := NewLanguage()
	lang.SetId(1)
	lang.SetName("English")

	list := NewImportListResource()
	list.SetLanguage(*lang)
	list.SetTagExisting(TAGEXISTINGTYPE_SYNC)

	var backList ImportListResource
	roundTrip(t, list, &backList, `"language":{"id":1,"name":"English"}`, `"tagExisting":"sync"`)
	listLang := backList.GetLanguage()
	if listLang.GetName() != "English" || backList.GetTagExisting() != TAGEXISTINGTYPE_SYNC {
		t.Fatalf("not read back: %+v", backList)
	}

	series := NewSeriesResource()
	series.SetLanguage(*lang)

	var backSeries SeriesResource
	roundTrip(t, series, &backSeries, `"language":{"id":1,"name":"English"}`)
	seriesLang := backSeries.GetLanguage()
	if seriesLang.GetId() != 1 {
		t.Fatalf("not read back: %+v", backSeries)
	}
}

func TestV4AnswersLeaveNewFieldsUnset(t *testing.T) {
	var mm MediaManagementConfigResource
	var rp ReleaseProfileResource
	var host HostConfigResource
	_ = json.Unmarshal([]byte(`{"id":1}`), &mm)
	_ = json.Unmarshal([]byte(`{"id":1,"name":"x"}`), &rp)
	_ = json.Unmarshal([]byte(`{"id":1}`), &host)

	if mm.HasSeasonPackUpgrade() || mm.HasSeasonPackUpgradeThreshold() || rp.HasAirDateRestriction() || rp.HasExcludedTags() || host.HasOidcScopes() {
		t.Fatal("fields Sonarr 4 doesn't send must stay unset")
	}
	for _, v := range []interface{}{mm, rp, host} {
		out, _ := json.Marshal(v)
		for _, f := range []string{"seasonPack", "airDate", "excludedTags", "oidc", "sslKeyPath"} {
			if strings.Contains(string(out), f) {
				t.Fatalf("unset %s must not be sent: %s", f, out)
			}
		}
	}
}

func TestNewEnumValues(t *testing.T) {
	for _, c := range []struct {
		v  string
		ok bool
	}{
		{string(RELEASETYPE_MULTI_SEASON_PACK), RELEASETYPE_MULTI_SEASON_PACK.IsValid()},
		{string(COMMANDRESULT_INDETERMINATE), COMMANDRESULT_INDETERMINATE.IsValid()},
		{string(IMPORTLISTTYPE_TMDB), IMPORTLISTTYPE_TMDB.IsValid()},
		{string(TAGEXISTINGTYPE_ADD), TAGEXISTINGTYPE_ADD.IsValid()},
	} {
		if !c.ok {
			t.Fatalf("%s not valid", c.v)
		}
	}
}
