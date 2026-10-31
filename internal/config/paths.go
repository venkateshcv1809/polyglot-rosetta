package config

// Native storage defaults. User cache locations are resolved through config.json;
// temporary workspaces use the OS temporary directory (including TMPDIR).
const (
	AppDirName             = "prosetta"
	ConfigFileName         = "config.json"
	CatalogCacheDirName    = "catalog"
	ToolchainCacheFileName = "toolchains.json"
	CatalogEnvironment     = "PROSETTA_CATALOG"
	CatalogURLEnvironment  = "PROSETTA_CATALOG_URL"
	CatalogURL             = "https://venkateshcv1809.github.io/polyglot-rosetta/catalog/"
)

// Catalog filenames are part of the bundle format, shared by its producer and consumers.
const (
	CatalogDirName        = "problems"
	TemplateDirName       = "templates"
	CatalogInfoFileName   = "info.json"
	CatalogIndexFileName  = "index.json"
	ManifestFileName      = "manifest.json"
	TemplatesArchiveName  = "templates.tar.gz"
	CategoryArchiveSuffix = ".tar.gz"
)

// Private cache layout and temporary filename patterns.
const (
	CatalogAssetsDirName    = "assets"
	CatalogSnapshotsDirName = "versions"
	CatalogCurrentFileName  = "current"
	ExecutionTempPattern    = "polyglot-rosetta-*"
	CatalogStagePattern     = ".staging-*"
	CatalogAssetPattern     = ".asset-*"
	CatalogCurrentPattern   = ".current-*"
	ConfigTempPattern       = ".config-*"
	ToolchainTempPattern    = ".toolchains-*"
	BundleTempPattern       = ".bundle-*"
)
