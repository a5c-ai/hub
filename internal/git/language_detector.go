package git

import (
	"path/filepath"
	"strings"
)

// LanguageDetector provides functionality to detect programming languages from file content and extensions
type LanguageDetector struct {
	extensionMap map[string]string
	filenameMap  map[string]string
}

// NewLanguageDetector creates a new language detector with predefined mappings
func NewLanguageDetector() *LanguageDetector {
	return &LanguageDetector{
		extensionMap: getExtensionLanguageMap(),
		filenameMap:  getFilenameLanguageMap(),
	}
}

// DetectLanguage detects the programming language from a file path and optional content
func (ld *LanguageDetector) DetectLanguage(filePath string, content []byte) string {
	// Check by filename first (e.g., Dockerfile, Makefile)
	filename := filepath.Base(filePath)
	if lang, exists := ld.filenameMap[strings.ToLower(filename)]; exists {
		return lang
	}

	// Check by extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if lang, exists := ld.extensionMap[ext]; exists {
		return lang
	}

	// Check for shebang in content if available
	if len(content) > 0 {
		if lang := ld.detectFromShebang(content); lang != "" {
			return lang
		}
	}

	// Default to unknown
	return "Unknown"
}

// detectFromShebang detects language from shebang line
func (ld *LanguageDetector) detectFromShebang(content []byte) string {
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		return ""
	}

	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "#!") {
		return ""
	}

	shebang := strings.ToLower(firstLine)

	if strings.Contains(shebang, "python") {
		return "Python"
	}
	if strings.Contains(shebang, "bash") || strings.Contains(shebang, "sh") {
		return "Shell"
	}
	if strings.Contains(shebang, "node") {
		return "JavaScript"
	}
	if strings.Contains(shebang, "ruby") {
		return "Ruby"
	}
	if strings.Contains(shebang, "perl") {
		return "Perl"
	}
	if strings.Contains(shebang, "php") {
		return "PHP"
	}

	return ""
}

// getExtensionLanguageMap returns a map of file extensions to programming languages
func getExtensionLanguageMap() map[string]string {
	return map[string]string{
		// Go
		".go": "Go",

		// JavaScript/TypeScript
		".js":  "JavaScript",
		".jsx": "JavaScript",
		".ts":  "TypeScript",
		".tsx": "TypeScript",
		".mjs": "JavaScript",
		".cjs": "JavaScript",

		// Python
		".py":  "Python",
		".pyx": "Python",
		".pyi": "Python",
		".pyw": "Python",

		// Java
		".java":  "Java",
		".class": "Java",

		// C/C++
		".c":   "C",
		".h":   "C",
		".cpp": "C++",
		".cxx": "C++",
		".cc":  "C++",
		".hpp": "C++",
		".hxx": "C++",

		// C#
		".cs": "C#",

		// Rust
		".rs": "Rust",

		// PHP
		".php":   "PHP",
		".phtml": "PHP",

		// Ruby
		".rb":  "Ruby",
		".rbw": "Ruby",

		// Shell
		".sh":   "Shell",
		".bash": "Shell",
		".zsh":  "Shell",
		".fish": "Shell",

		// HTML/CSS
		".html": "HTML",
		".htm":  "HTML",
		".css":  "CSS",
		".scss": "SCSS",
		".sass": "Sass",
		".less": "Less",

		// Database
		".sql": "SQL",

		// Configuration
		".json": "JSON",
		".yaml": "YAML",
		".yml":  "YAML",
		".toml": "TOML",
		".xml":  "XML",
		".ini":  "INI",
		".conf": "Configuration",
		".cfg":  "Configuration",

		// Documentation
		".md":  "Markdown",
		".rst": "reStructuredText",
		".txt": "Text",

		// Docker
		".dockerfile": "Dockerfile",

		// Terraform
		".tf":  "HCL",
		".hcl": "HCL",

		// Kotlin
		".kt":  "Kotlin",
		".kts": "Kotlin",

		// Swift
		".swift": "Swift",

		// Objective-C
		".m":  "Objective-C",
		".mm": "Objective-C++",

		// Perl
		".pl": "Perl",
		".pm": "Perl",

		// R
		".r": "R",
		".R": "R",

		// Scala
		".scala": "Scala",
		".sc":    "Scala",

		// Dart
		".dart": "Dart",

		// Lua
		".lua": "Lua",

		// Erlang
		".erl": "Erlang",
		".hrl": "Erlang",

		// Elixir
		".ex":  "Elixir",
		".exs": "Elixir",

		// Haskell
		".hs":  "Haskell",
		".lhs": "Haskell",

		// Clojure
		".clj":  "Clojure",
		".cljs": "ClojureScript",
		".cljc": "Clojure",

		// F#
		".fs":  "F#",
		".fsx": "F#",
		".fsi": "F#",

		// PowerShell
		".ps1":  "PowerShell",
		".psm1": "PowerShell",
		".psd1": "PowerShell",

		// Assembly
		".asm": "Assembly",
		".s":   "Assembly",

		// Visual Basic
		".vb": "Visual Basic",

		// Groovy
		".groovy": "Groovy",
		".gradle": "Gradle",

		// Protocol Buffers
		".proto": "Protocol Buffer",
	}
}

// getFilenameLanguageMap returns a map of specific filenames to programming languages
func getFilenameLanguageMap() map[string]string {
	return map[string]string{
		"dockerfile":         "Dockerfile",
		"dockerfile.dev":     "Dockerfile",
		"dockerfile.prod":    "Dockerfile",
		"dockerfile.test":    "Dockerfile",
		"makefile":           "Makefile",
		"makefile.am":        "Makefile",
		"makefile.in":        "Makefile",
		"rakefile":           "Ruby",
		"gemfile":            "Ruby",
		"gemfile.lock":       "Ruby",
		"package.json":       "JSON",
		"package-lock.json":  "JSON",
		"composer.json":      "JSON",
		"composer.lock":      "JSON",
		"yarn.lock":          "YAML",
		"requirements.txt":   "Text",
		"pipfile":            "TOML",
		"pipfile.lock":       "JSON",
		"cargo.toml":         "TOML",
		"cargo.lock":         "TOML",
		"go.mod":             "Go Module",
		"go.sum":             "Go Module",
		"pom.xml":            "XML",
		"build.gradle":       "Gradle",
		"settings.gradle":    "Gradle",
		"cmake.txt":          "CMake",
		"cmakelists.txt":     "CMake",
		"vcxproj":            "MSBuild",
		"pubspec.yaml":       "YAML",
		"pubspec.lock":       "YAML",
		"podfile":            "Ruby",
		"podfile.lock":       "YAML",
		"readme":             "Text",
		"readme.md":          "Markdown",
		"readme.txt":         "Text",
		"license":            "Text",
		"license.md":         "Markdown",
		"license.txt":        "Text",
		"changelog":          "Text",
		"changelog.md":       "Markdown",
		"changelog.txt":      "Text",
		"contributing.md":    "Markdown",
		"code_of_conduct.md": "Markdown",
		"security.md":        "Markdown",
		".gitignore":         "Gitignore",
		".gitattributes":     "Gitattributes",
		".editorconfig":      "EditorConfig",
		".eslintrc":          "JSON",
		".eslintrc.js":       "JavaScript",
		".eslintrc.json":     "JSON",
		".eslintrc.yaml":     "YAML",
		".eslintrc.yml":      "YAML",
		".prettierrc":        "JSON",
		".prettierrc.js":     "JavaScript",
		".prettierrc.json":   "JSON",
		".prettierrc.yaml":   "YAML",
		".prettierrc.yml":    "YAML",
		".babelrc":           "JSON",
		".babelrc.js":        "JavaScript",
		".babelrc.json":      "JSON",
		"jest.config.js":     "JavaScript",
		"webpack.config.js":  "JavaScript",
		"rollup.config.js":   "JavaScript",
		"vite.config.js":     "JavaScript",
		"vite.config.ts":     "TypeScript",
		"tsconfig.json":      "JSON",
		"tslint.json":        "JSON",
		"angular.json":       "JSON",
		"nx.json":            "JSON",
		"next.config.js":     "JavaScript",
		"next.config.ts":     "TypeScript",
		"nuxt.config.js":     "JavaScript",
		"nuxt.config.ts":     "TypeScript",
		"vue.config.js":      "JavaScript",
		"svelte.config.js":   "JavaScript",
		"tailwind.config.js": "JavaScript",
		"postcss.config.js":  "JavaScript",
		".env":               "Environment",
		".env.local":         "Environment",
		".env.example":       "Environment",
		".env.development":   "Environment",
		".env.production":    "Environment",
		".env.test":          "Environment",
	}
}
