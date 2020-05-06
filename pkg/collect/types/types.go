package types

// Collector represents a single collect definition
type Collector struct {
	Name       string   `yaml:"name"`
	Path       string   `yaml:"path"`
	Parser     string   `yaml:"parser"`
	Collectors []string `yaml:"collectors"`
}

type Output struct {
	Owner         string   `yaml:"owner"`
	Repo          string   `yaml:"repo"`
	IsPublic      bool     `yaml:"isPublic"`
	IsArchived    bool     `yaml:"isArchived"`
	DefaultBranch string   `yaml:"defaultBranch"`
	IsFork        bool     `yaml:"isFork"`
	Forks         []string `yaml:"forks,omitempty"`
	Parent        string   `yaml:"parent,omitempty"`
	Head          string   `yaml:"head"`
}

func (c Collector) Equals(other Collector) bool {
	return c.Parser == other.Parser &&
		c.Path == other.Path

}

func (c *Collector) Merge(other Collector) {
	uniqueCollectors := map[string]struct{}{}

	for _, collector := range c.Collectors {
		uniqueCollectors[collector] = struct{}{}
	}
	for _, collector := range other.Collectors {
		uniqueCollectors[collector] = struct{}{}
	}

	c.Collectors = []string{}
	for k := range uniqueCollectors {
		c.Collectors = append(c.Collectors, k)
	}
}
