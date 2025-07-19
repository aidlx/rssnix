package main

import "testing"

func TestFileNameTruncation(t *testing.T) {
	names := []string{
		"我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我",  // 255 x Chinese wo3 (我)
		"我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我", // 256 x Chinese wo3 (我)
		"我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我我", // 32 x Chinese wo3 (我)
		"short"} // should not get truncated

	for _, name := range names {
		shortened := truncateString(name, maxFileNameLength)
		if len(name) < maxFileNameLength {
			if name != shortened {
				t.Errorf("Filename should not be altered, but it was. Original was %s", name)
			}
		} else {
			if len(shortened) > maxFileNameLength {
				t.Errorf("Filename was too long - should have been truncated. Length was %d", len(name))
			}
		}
	}
}
