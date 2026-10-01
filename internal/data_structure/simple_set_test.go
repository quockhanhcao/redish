package data_structure

import (
	"slices"
	"testing"
)

func TestSimpleSetAdd(t *testing.T) {
	t.Run("return number of new members", func(t *testing.T) {
		s := NewSimpleSet()
		added := s.Add("a", "b", "c")
		if added != 3 {
			t.Errorf("Add(a, b, c) on empty set = %d; want 3", added)
		}
	})
	t.Run("existing members are not counted", func(t *testing.T) {
		s := NewSimpleSet()
		s.Add("a", "b")
		added := s.Add("b", "c")
		if added != 1 {
			t.Errorf("Add(b, c) on set with {a, b} = %d; want 1", added)
		}
	})
	t.Run("duplicates within one call count once", func(t *testing.T) {
		s := NewSimpleSet()
		added := s.Add("a", "b", "c", "a")
		if added != 3 {
			t.Errorf("Add(a, b, c, a) on empty set = %d; want 3", added)
		}
		assertMembers(t, s, "a", "b", "c")
	})
	t.Run("Remove returns the number of members actually removed and removed members are gone", func(t *testing.T) {
		s := NewSimpleSet()
		s.Add("a", "b", "c", "d", "e")
		removed := s.Remove("a", "b")
		if removed != 2 {
			t.Errorf("Remove(a, b) on set with {a, b, c, d, e} = %d; want 2", removed)
		}
		members := s.GetMembers()
		assertMembers(t, s, members...)
	})
	t.Run("Exists return 1 for members and 0 for non-members", func(t *testing.T) {
		s := NewSimpleSet()
		s.Add("a")
		exists := s.Exist("a")
		if exists != 1 {
			t.Errorf("Exist(a) on set with {a} = %d; want 1", exists)
		}
		nonExist := s.Exist("b")
		if nonExist != 0 {
			t.Errorf("Exist(b) on set with {a} = %d; want 0", nonExist)
		}
	})
	t.Run("GetMembers returns empty slice with empty list", func (t *testing.T) {
		s := NewSimpleSet()
		members := s.GetMembers()
		if len(members) != 0 {
			t.Errorf("GetMembers on empty set return slice length = %d; want 0", len(members))
		}
	})
}

func assertMembers(t *testing.T, s *SimpleSet, want ...string) {
	t.Helper()
	got := s.GetMembers()
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("members = %q; want %q", got, want)
	}
}
