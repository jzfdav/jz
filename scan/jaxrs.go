package scan

import (
	"bufio"
	"jz/model"
	"os"
	"path/filepath"
	"strings"
)

// Scan recursively walks the rootDir and extracts JAX-RS entry points.
func Scan(rootDir string) ([]model.EntryPoint, error) {
	var entryPoints []model.EntryPoint

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".java") {
			return nil
		}

		eps, err := scanFile(path)
		// We ignore file reading errors to prevent stopping the entire walk
		if err == nil {
			entryPoints = append(entryPoints, eps...)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return entryPoints, nil
}

func scanFile(filePath string) ([]model.EntryPoint, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var entryPoints []model.EntryPoint

	var pendingClassPath string
	var classPath string
	var className string

	// Method scanning state
	var methodPath string
	var methodProto string // HTTP method like GET, POST

	var hasHTTPMethod bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 1. Package (optional, skipped as not used in model)

		// 2. Class @Path
		// Looking for @Path("/...")
		if strings.HasPrefix(line, "@Path") {
			extracted := extractPath(line)
			if className == "" {
				// Assume class-level if we haven't seen class def yet
				pendingClassPath = extracted
			} else {
				// Method-level path
				methodPath = extracted

			}
			continue
		}

		// 3. HTTP Methods
		// @GET, @POST, etc.
		if isHTTPMethod(line) {
			methodProto = extractHTTPMethod(line)
			hasHTTPMethod = true
			continue
		}

		// 4. Class Definition
		if strings.Contains(line, "class ") || strings.Contains(line, "interface ") {
			// Very basic extraction: expected "public class ClassName ..."
			// We take the word after "class"
			parts := strings.Fields(line)
			for i, p := range parts {
				if (p == "class" || p == "interface") && i+1 < len(parts) {
					className = parts[i+1]
					// Strip potnetial "{" or "implements" or "extends"
					// Although usually it's "class Name {" or "class Name implements"
					// We only take the name.
					className = strings.Split(className, "{")[0] // robust against class Name{

					// Assign pending class path if any
					if pendingClassPath != "" {
						classPath = pendingClassPath
						pendingClassPath = ""
					}
					break
				}
			}
			continue
		}

		// 5. Method Definition
		// Requirement: hasHTTPMethod && line contains "(" && (public|protected|private)
		if hasHTTPMethod && strings.Contains(line, "(") {
			isVisibilityPresent := strings.Contains(line, "public ") ||
				strings.Contains(line, "protected ") ||
				strings.Contains(line, "private ")

			if isVisibilityPresent {
				// Extract method name
				// Scan backwards from first '('
				idx := strings.Index(line, "(")
				preParens := line[:idx]
				parts := strings.Fields(preParens)
				if len(parts) > 0 {
					methodName := parts[len(parts)-1]

					fullPath := buildPath(classPath, methodPath)

					ep := model.EntryPoint{
						Method:     methodProto,
						Path:       fullPath,
						Handler:    className + "." + methodName,
						SourceFile: filePath,
					}
					entryPoints = append(entryPoints, ep)
				}

				// Reset method state
				methodPath = ""
				methodProto = ""

				hasHTTPMethod = false
			}
		}
	}

	return entryPoints, nil
}

func extractPath(line string) string {
	// @Path("/foo") or @Path(value = "/foo")
	// Simple string extraction: content between quotes
	start := strings.Index(line, "\"")
	if start == -1 {
		return ""
	}
	end := strings.LastIndex(line, "\"")
	if end <= start {
		return ""
	}
	return line[start+1 : end]
}

func isHTTPMethod(line string) bool {
	return strings.HasPrefix(line, "@GET") ||
		strings.HasPrefix(line, "@POST") ||
		strings.HasPrefix(line, "@PUT") ||
		strings.HasPrefix(line, "@DELETE") ||
		strings.HasPrefix(line, "@PATCH") ||
		strings.HasPrefix(line, "@HEAD") ||
		strings.HasPrefix(line, "@OPTIONS")
}

func extractHTTPMethod(line string) string {
	line = strings.TrimSpace(line)
	// @GET -> GET
	// @GET(...) -> GET
	if idx := strings.Index(line, "("); idx != -1 {
		line = line[:idx]
	}
	return strings.TrimPrefix(line, "@")
}

func buildPath(classPath, methodPath string) string {
	// Normalize
	cp := strings.Trim(classPath, "/")
	mp := strings.Trim(methodPath, "/")

	final := "/" + cp
	if mp != "" {
		if cp != "" {
			final += "/"
		}
		final += mp
	}

	// Ensure single slash start if empty components
	if final == "/" {
		return "/" // Root
	}

	// Handle duplicates just in case
	for strings.Contains(final, "//") {
		final = strings.ReplaceAll(final, "//", "/")
	}

	return final
}
