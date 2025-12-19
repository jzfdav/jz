package scan

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type OSGIBundle struct {
	SymbolicName      string
	Name              string
	Version           string
	ServiceComponents []string
	ManifestPath      string
}

// ScanOSGi recursively walks the rootDir and extracts OSGi bundle metadata.
func ScanOSGi(rootDir string) ([]OSGIBundle, error) {
	var bundles []OSGIBundle

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.ToUpper(filepath.Base(path)) == "MANIFEST.MF" {
			bundle, err := parseManifest(path)
			if err != nil {
				// Ignore parsing errors, just like jaxrs scanner
				return nil
			}
			if bundle.SymbolicName != "" {
				bundles = append(bundles, bundle)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return bundles, nil
}

func parseManifest(path string) (OSGIBundle, error) {
	f, err := os.Open(path)
	if err != nil {
		return OSGIBundle{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var bundle OSGIBundle
	bundle.ManifestPath = path

	var currentHeader string
	var currentValue string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, " ") {
			// Continuation line
			if currentHeader != "" {
				currentValue += strings.TrimPrefix(line, " ")
			}
			continue
		}

		// Process previous header
		if currentHeader != "" {
			assignHeader(&bundle, currentHeader, currentValue)
		}

		// New header
		if idx := strings.Index(line, ":"); idx != -1 {
			currentHeader = strings.TrimSpace(line[:idx])
			currentValue = strings.TrimSpace(line[idx+1:])
		} else {
			// Invalid or empty line resets state
			currentHeader = ""
			currentValue = ""
		}
	}

	// Process last header
	if currentHeader != "" {
		assignHeader(&bundle, currentHeader, currentValue)
	}

	return bundle, nil
}

func assignHeader(b *OSGIBundle, header, value string) {
	switch header {
	case "Bundle-SymbolicName":
		// Can have attributes like ;singleton:=true
		parts := strings.Split(value, ";")
		b.SymbolicName = strings.TrimSpace(parts[0])
	case "Bundle-Name":
		b.Name = value
	case "Bundle-Version":
		b.Version = value
	case "Service-Component":
		// potential wildcards, comma separated
		parts := strings.Split(value, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				b.ServiceComponents = append(b.ServiceComponents, trimmed)
			}
		}
	}
}
