package actors

import (
	"g7/comprehensive/model_compre/battle/skills"
	"time"
)

type Skills struct {
	skills   map[int32]skills.Skill
	skillCDs map[int32]time.Duration
}

func (s *Skills) DefaultSkills() {
	s.skillCDs = make(map[int32]time.Duration)
	s.skills = make(map[int32]skills.Skill)
}

func (s *Skills) GetSkill(skillid int32) skills.Skill {
	return s.skills[skillid]
}

func (s *Skills) StartSkillCD(skillId int32, startCDTime time.Duration) {
	s.skillCDs[skillId] = startCDTime
}

func (s *Skills) UpdateSkillCD(skillId int32, reduce time.Duration) {
	v, ok := s.skillCDs[skillId]
	if !ok {
		return
	}
	s.skillCDs[skillId] = max(v-reduce, 0)
}

func (s *Skills) UpdateAllCD(reduce time.Duration) {
	for k, v := range s.skillCDs {
		if v <= 0 {
			continue
		}
		s.skillCDs[k] = max(v-reduce, 0)
	}
}

func (s *Skills) IsCDOk(skillId int32) bool {
	v, ok := s.skillCDs[skillId]
	if !ok {
		return true
	}
	return v <= 0
}
