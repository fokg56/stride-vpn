package client

import "sync"

type GroupInfo struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Configs []string `json:"configs"`
}

type GroupManager struct {
	mu     sync.RWMutex
	groups []*GroupInfo
}

func NewGroupManager() *GroupManager {
	return &GroupManager{
		groups: []*GroupInfo{
			{ID: "default", Name: "По умолчанию", Configs: []string{}},
		},
	}
}

func (gm *GroupManager) List() []*GroupInfo {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	res := make([]*GroupInfo, len(gm.groups))
	for i, g := range gm.groups {
		g2 := *g
		g2.Configs = make([]string, len(g.Configs))
		copy(g2.Configs, g.Configs)
		res[i] = &g2
	}
	return res
}

func (gm *GroupManager) Create(name string) *GroupInfo {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	id := "group-" + name
	for _, g := range gm.groups {
		if g.ID == id {
			g.Name = name
			return g
		}
	}
	g := &GroupInfo{ID: id, Name: name, Configs: []string{}}
	gm.groups = append(gm.groups, g)
	return g
}

func (gm *GroupManager) Delete(id string) bool {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	for i, g := range gm.groups {
		if g.ID == id {
			gm.groups = append(gm.groups[:i], gm.groups[i+1:]...)
			return true
		}
	}
	return false
}

func (gm *GroupManager) AssignConfig(groupID, configID string) bool {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	for _, g := range gm.groups {
		if g.ID == groupID {
			for _, c := range g.Configs {
				if c == configID {
					return true
				}
			}
			g.Configs = append(g.Configs, configID)
			return true
		}
	}
	return false
}

func (gm *GroupManager) RemoveConfig(groupID, configID string) bool {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	for _, g := range gm.groups {
		if g.ID == groupID {
			for i, c := range g.Configs {
				if c == configID {
					g.Configs = append(g.Configs[:i], g.Configs[i+1:]...)
					return true
				}
			}
		}
	}
	return false
}

func (gm *GroupManager) GetGroupForConfig(configID string) string {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	for _, g := range gm.groups {
		for _, c := range g.Configs {
			if c == configID {
				return g.ID
			}
		}
	}
	return "default"
}
