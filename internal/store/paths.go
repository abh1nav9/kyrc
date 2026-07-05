package store

// ConfigDir exposes kyrc's config directory so other packages (identity)
// write their files alongside results.json. It does not create the dir.
func ConfigDir() (string, error) {
	return configDir()
}
