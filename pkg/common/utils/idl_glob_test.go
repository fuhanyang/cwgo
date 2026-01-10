package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandIDLPaths_SingleFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.proto")
	if err := os.WriteFile(p, []byte("syntax = \"proto3\";"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ExpandIDLPaths(p)
	if err != nil {
		t.Fatalf("ExpandIDLPaths err: %v", err)
	}
	if len(got) != 1 || got[0] != p {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestExpandIDLPaths_Glob(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.proto")
	p2 := filepath.Join(dir, "b.proto")
	if err := os.WriteFile(p1, []byte("syntax = \"proto3\";"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("syntax = \"proto3\";"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ExpandIDLPaths(filepath.Join(dir, "*.proto"))
	if err != nil {
		t.Fatalf("ExpandIDLPaths err: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("unexpected len: %d, %#v", len(got), got)
	}
}

func TestExpandIDLPaths_List(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.proto")
	p2 := filepath.Join(dir, "b.proto")
	if err := os.WriteFile(p1, []byte("syntax = \"proto3\";"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("syntax = \"proto3\";"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ExpandIDLPaths(p2 + ";" + p1 + ";" + p2)
	if err != nil {
		t.Fatalf("ExpandIDLPaths err: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("unexpected len: %d, %#v", len(got), got)
	}
}

func TestSelectRootIDLByService(t *testing.T) {
	dir := t.TempDir()
	noSvc := filepath.Join(dir, "types.proto")
	withSvc := filepath.Join(dir, "svc.proto")
	if err := os.WriteFile(noSvc, []byte("syntax = \"proto3\";\nmessage A {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// include some comments to ensure we don't match "service" in comments.
	content := `syntax = "proto3";
// service Fake {}
/* service AlsoFake {} */
service Foo {
  rpc Ping (A) returns (A);
}
message A {}
`
	if err := os.WriteFile(withSvc, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := SelectRootIDLByService([]string{noSvc, withSvc}, "Foo")
	if err != nil {
		t.Fatalf("SelectRootIDLByService err: %v", err)
	}
	if root != withSvc {
		t.Fatalf("unexpected root: %s", root)
	}
}
