package git

import (
	"testing"
)

func TestLanguageDetector_DetectLanguage(t *testing.T) {
	detector := NewLanguageDetector()

	tests := []struct {
		name        string
		filePath    string
		content     []byte
		expected    string
	}{
		{
			name:     "Go file by extension",
			filePath: "main.go",
			content:  []byte("package main\n\nfunc main() {}\n"),
			expected: "Go",
		},
		{
			name:     "JavaScript file by extension",
			filePath: "app.js",
			content:  []byte("console.log('Hello, World!');\n"),
			expected: "JavaScript",
		},
		{
			name:     "TypeScript file by extension",
			filePath: "component.tsx",
			content:  []byte("const Component: React.FC = () => <div>Hello</div>;\n"),
			expected: "TypeScript",
		},
		{
			name:     "Python file by extension",
			filePath: "script.py",
			content:  []byte("print('Hello, World!')\n"),
			expected: "Python",
		},
		{
			name:     "Dockerfile by filename",
			filePath: "Dockerfile",
			content:  []byte("FROM ubuntu:20.04\nRUN apt-get update\n"),
			expected: "Dockerfile",
		},
		{
			name:     "Makefile by filename",
			filePath: "Makefile",
			content:  []byte("all:\n\techo 'Building...'\n"),
			expected: "Makefile",
		},
		{
			name:     "JSON package file",
			filePath: "package.json",
			content:  []byte("{\n  \"name\": \"test\"\n}\n"),
			expected: "JSON",
		},
		{
			name:     "Python shebang",
			filePath: "script",
			content:  []byte("#!/usr/bin/env python3\nprint('Hello')\n"),
			expected: "Python",
		},
		{
			name:     "Bash shebang",
			filePath: "script",
			content:  []byte("#!/bin/bash\necho 'Hello'\n"),
			expected: "Shell",
		},
		{
			name:     "Node.js shebang",
			filePath: "script",
			content:  []byte("#!/usr/bin/env node\nconsole.log('Hello');\n"),
			expected: "JavaScript",
		},
		{
			name:     "Unknown file",
			filePath: "unknown.xyz",
			content:  []byte("some unknown content\n"),
			expected: "Unknown",
		},
		{
			name:     "README markdown",
			filePath: "README.md",
			content:  []byte("# Project\n\nThis is a test project.\n"),
			expected: "Markdown",
		},
		{
			name:     "CSS file",
			filePath: "styles.css",
			content:  []byte("body { margin: 0; }\n"),
			expected: "CSS",
		},
		{
			name:     "HTML file",
			filePath: "index.html",
			content:  []byte("<!DOCTYPE html><html><head><title>Test</title></head></html>\n"),
			expected: "HTML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectLanguage(tt.filePath, tt.content)
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q, %q) = %q, expected %q", tt.filePath, string(tt.content), result, tt.expected)
			}
		})
	}
}

func TestLanguageDetector_DetectFromShebang(t *testing.T) {
	detector := NewLanguageDetector()

	tests := []struct {
		name     string
		content  []byte
		expected string
	}{
		{
			name:     "Python shebang",
			content:  []byte("#!/usr/bin/env python3\n"),
			expected: "Python",
		},
		{
			name:     "Bash shebang",
			content:  []byte("#!/bin/bash\n"),
			expected: "Shell",
		},
		{
			name:     "Node shebang",
			content:  []byte("#!/usr/bin/env node\n"),
			expected: "JavaScript",
		},
		{
			name:     "Ruby shebang",
			content:  []byte("#!/usr/bin/env ruby\n"),
			expected: "Ruby",
		},
		{
			name:     "Perl shebang",
			content:  []byte("#!/usr/bin/perl\n"),
			expected: "Perl",
		},
		{
			name:     "PHP shebang",
			content:  []byte("#!/usr/bin/php\n"),
			expected: "PHP",
		},
		{
			name:     "No shebang",
			content:  []byte("just regular content\n"),
			expected: "",
		},
		{
			name:     "Unknown shebang",
			content:  []byte("#!/unknown/interpreter\n"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.detectFromShebang(tt.content)
			if result != tt.expected {
				t.Errorf("detectFromShebang(%q) = %q, expected %q", string(tt.content), result, tt.expected)
			}
		})
	}
}

func TestNewLanguageDetector(t *testing.T) {
	detector := NewLanguageDetector()
	
	if detector == nil {
		t.Error("NewLanguageDetector() returned nil")
	}
	
	if detector.extensionMap == nil {
		t.Error("extensionMap is nil")
	}
	
	if detector.filenameMap == nil {
		t.Error("filenameMap is nil")
	}
	
	// Test that common extensions are mapped
	expectedExtensions := []string{".go", ".js", ".py", ".java", ".cpp", ".rs", ".php", ".rb"}
	for _, ext := range expectedExtensions {
		if _, exists := detector.extensionMap[ext]; !exists {
			t.Errorf("Extension %q not found in extensionMap", ext)
		}
	}
	
	// Test that common filenames are mapped
	expectedFilenames := []string{"dockerfile", "makefile", "package.json", "readme.md"}
	for _, filename := range expectedFilenames {
		if _, exists := detector.filenameMap[filename]; !exists {
			t.Errorf("Filename %q not found in filenameMap", filename)
		}
	}
}