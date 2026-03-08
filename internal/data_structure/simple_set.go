package data_structure

type SimpleSet struct {
	dict map[string]struct{}
}

func NewSimpleSet() *SimpleSet {
	return &SimpleSet{
		dict: make(map[string]struct{}),
	}
}

func (s *SimpleSet) Add(members ...string) int {
	added := 0
	for _, member := range members {
		if _, ok := s.dict[member]; ok {
			continue
		}
		s.dict[member] = struct{}{}
		added += 1
	}
	return added
}

func (s *SimpleSet) Remove(members ...string) int {
	removed := 0
	for _, member := range members {
		if _, ok := s.dict[member]; !ok {
			continue
		}
		delete(s.dict, member)
		removed += 1
	}
	return removed
}

func (s *SimpleSet) Exist(member string) int {
	if _, ok := s.dict[member]; !ok {
		return 0
	}
	return 1
}

func (s *SimpleSet) GetMembers() []string {
	keys := make([]string, 0, len(s.dict))
	for k := range s.dict {
		keys = append(keys, k)
	}
	return keys
}
