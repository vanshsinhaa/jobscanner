package database

import (
	"fmt"
	"testing"
	"time"
)

func TestClassifyRole(t *testing.T) {
	cases := []struct {
		title string
		want  string
	}{
		// intern
		{"Software Engineer Intern (Summer 2027)", "intern"},
		{"2027 SDE Internship", "intern"},
		{"Software Engineering Co-op - Fall 2027", "intern"},
		{"Internships, Machine Learning", "intern"},
		// new grad — keyword forms
		{"Software Engineer, New Grad", "new_grad"},
		{"Entry Level Software Developer", "new_grad"},
		{"Graduate Software Engineer", "new_grad"},
		// new grad — cycle-year forms (2027 hiring cycle phrasing)
		{"Class of 2027 - Software Engineer", "new_grad"},
		{"Software Engineer I (2027 Grads)", "new_grad"},
		{"New Grads 2027 (US)", "new_grad"},
		{"Software Engineer - 2027 Graduates", "new_grad"},
		// general — must not false-positive
		{"International Sales Manager", "general"},
		{"Internal Tools Engineer", "general"},
		{"Senior Software Engineer", "general"},
		{"Engineer II", "general"},
	}
	for _, c := range cases {
		if got := ClassifyRole(c.title); got != c.want {
			t.Errorf("ClassifyRole(%q) = %q, want %q", c.title, got, c.want)
		}
	}
}

func TestCycleYear(t *testing.T) {
	cases := []struct {
		title string
		want  int
	}{
		{"Software Engineer Intern (Summer 2027)", 2027},
		{"2026/2027 Rotational Program", 2027}, // ranges resolve to the later cycle
		{"Software Engineer, New Grad", 0},     // no year named
		{"Engineer II", 0},
	}
	for _, c := range cases {
		if got := CycleYear(c.title); got != c.want {
			t.Errorf("CycleYear(%q) = %d, want %d", c.title, got, c.want)
		}
	}
}

func TestIsStaleCycle(t *testing.T) {
	year := time.Now().Year()
	last := fmt.Sprintf("Summer %d Intern", year-1)
	current := fmt.Sprintf("Fall %d Intern", year)
	next := fmt.Sprintf("Summer %d Intern", year+1)

	if !IsStaleCycle(last) {
		t.Errorf("IsStaleCycle(%q) = false, want true (past cycle)", last)
	}
	if IsStaleCycle(current) {
		t.Errorf("IsStaleCycle(%q) = true, want false (current-year cohorts still active)", current)
	}
	if IsStaleCycle(next) {
		t.Errorf("IsStaleCycle(%q) = true, want false (upcoming cycle)", next)
	}
	if IsStaleCycle("Software Engineer, New Grad") {
		t.Error("IsStaleCycle should pass titles with no year")
	}
}
