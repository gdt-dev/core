package api

// Dependency describes a prerequisite binary or package that must be present.
type Dependency struct {
	// Name is the name of the binary or package
	Name string `yaml:"name"`
	// When describes any constraining conditions that apply to this
	// Dependency.
	When *DependencyConstraints `yaml:"when,omitempty"`
}

// DependencyConstraints describes constraining conditions that apply to a
// Dependency, for instance whether the dependency is only required on a
// particular OS or whether a particular version constraint applies to the
// dependency.
type DependencyConstraints struct {
	// OS indicates that the dependency only applies when the tests are run on
	// a particular operating system.
	OS string `yaml:"os,omitempty"`
	// Version indicates a version constraint to apply to the dependency, e.g.
	// >= 1.2.3
	Version string `yaml:"version,omitempty"`
}
