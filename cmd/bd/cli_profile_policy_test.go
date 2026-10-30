package main

import "testing"

func TestSelectCLIProfile(t *testing.T) {
	text := func(value string) *string { return &value }
	for _, tc := range []struct {
		name     string
		saved    *string
		argument *string
		profile  cliSurfaceProfile
		source   cliProfileSource
		wantErr  bool
	}{
		{name: "initial default", profile: cliSurfaceCompat, source: cliProfileBuiltin},
		{name: "saved canonical", saved: text("canonical"), profile: cliSurfaceCanonical, source: cliProfileSaved},
		{name: "saved compat", saved: text("compat"), profile: cliSurfaceCompat, source: cliProfileSaved},
		{name: "canonical override", saved: text("compat"), argument: text("canonical"), profile: cliSurfaceCanonical, source: cliProfileArgument},
		{name: "compat override", saved: text("canonical"), argument: text("compat"), profile: cliSurfaceCompat, source: cliProfileArgument},
		{name: "explicit empty override", argument: text(""), wantErr: true},
		{name: "empty saved setting", saved: text(""), wantErr: true},
		{name: "unknown effective profile", saved: text("legacy"), wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectCLIProfile(tc.saved, tc.argument)
			if (err != nil) != tc.wantErr {
				t.Fatalf("selection error = %v; want error %v", err, tc.wantErr)
			}
			if !tc.wantErr && (got.Profile != tc.profile || got.Source != tc.source) {
				t.Fatalf("selection = %#v; want %s from %s", got, tc.profile, tc.source)
			}
		})
	}
}

func TestCLIProfileOverrideDoesNotChangeSavedDefault(t *testing.T) {
	saved := "canonical"
	override := "compat"
	got, err := selectCLIProfile(&saved, &override)
	if err != nil || got.Profile != cliSurfaceCompat {
		t.Fatalf("override selection = %#v, %v", got, err)
	}
	if saved != "canonical" {
		t.Fatalf("saved default changed to %q", saved)
	}
	got, err = selectCLIProfile(&saved, nil)
	if err != nil || got.Profile != cliSurfaceCanonical || got.Source != cliProfileSaved {
		t.Fatalf("following selection = %#v, %v", got, err)
	}
	got, err = selectCLIProfile(nil, nil)
	if err != nil || got.Profile != cliSurfaceCompat || got.Source != cliProfileBuiltin {
		t.Fatalf("cleared selection = %#v, %v", got, err)
	}
}
