// Package config defines repository-wide defaults shared by application
// services and command entrypoints.
package config

const (
	// CatalogDir is the default root for categories and concepts.
	CatalogDir = "./" + CatalogDirName
	// TemplateDir is the default root for scaffold templates.
	TemplateDir = "./" + TemplateDirName
	// DistributionDir is the default root for generated artifacts.
	DistributionDir = "./dist"
)
