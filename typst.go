package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path"
)

//go:embed static/*
var staticFiles embed.FS

type Typst struct {
	dossier    *Dossier
	tempDir    string
	outputFile string
	debugMode  bool
}

func NewTypst(dossier *Dossier, outputFile string, debugMode bool) (*Typst, error) {
	tempDir, err := os.MkdirTemp("", "typst-*")
	if err != nil {
		return nil, err
	}
	return &Typst{
		dossier:    dossier,
		tempDir:    tempDir,
		outputFile: outputFile,
		debugMode:  debugMode,
	}, nil
}

func (t *Typst) Build(debugTempDir bool) error {
	err := t.initTempDir()
	if err != nil {
		return err
	}
	if err := t.dossier.ToJSON(path.Join(t.tempDir, "dossier.json")); err != nil {
		return err
	}
	if debugTempDir {
		if err := t.debugTempDir(); err != nil {
			return err
		}
	}
	if err := t.buildTemplate(); err != nil {
		return err
	}
	return nil
}

func (t *Typst) initTempDir() error {
	for _, doc := range t.dossier.JournalEntries {
		// Only link if original file is valid and uuid is present
		if doc.IsValidFile && doc.FileUUID != "" {
			err := doc.CreateSymlinkInFolder(t.tempDir)
			if err != nil {
				return err
			}
		}
	}
	return writeDirToTarget(staticFiles, "static", t.tempDir)
}

// writeDirToTarget recursively copies contents of embeddedDir in fsys to targetDir.
func writeDirToTarget(fsys embed.FS, embeddedDir, targetDir string) error {
	entries, err := fsys.ReadDir(embeddedDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := embeddedDir + "/" + entry.Name()
		dstPath := targetDir + "/" + entry.Name()
		if entry.IsDir() {
			// Create the new directory in the temp root.
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			// Recursively copy files inside the directory.
			if err := writeDirToTarget(fsys, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := fsys.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Typst) buildTemplate() error {
	// Change the working directory to the temp dir so that typst can resolve relative paths
	if err := os.Chdir(t.tempDir); err != nil {
		return fmt.Errorf("failed to change working directory to temp dir: %w", err)
	}

	cmd := exec.Command(
		"typst", "compile",
		"--input", "input=dossier.json",
		"--input", fmt.Sprintf("debug=%t", t.debugMode),
		"template.typ",
		t.outputFile,
	)
	cmd.Dir = t.tempDir

	// Capture stdout & stderr to show on error
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()
	if runErr != nil {
		fmt.Println("Typst build failed.")
		if out := outBuf.String(); out != "" {
			fmt.Println("Stdout:", out)
		}
		if errOut := errBuf.String(); errOut != "" {
			fmt.Println("Stderr:", errOut)
		}
		return runErr
	}

	return nil
}

func (t *Typst) debugTempDir() error {
	fmt.Printf("Opening temp dir in Finder: %s\nPress Enter to continue...", t.tempDir)
	cmd := exec.Command("open", t.tempDir)
	err := cmd.Run()
	if err != nil {
		return err
	}
	// Wait for user to hit Enter before returning
	_, err = fmt.Scanln()
	if err != nil && err.Error() != "unexpected newline" {
		return err
	}
	return nil
}

// Close releases resources held by Typst, such as removing the tempDir.
func (t *Typst) Close() error {
	if t.tempDir != "" {
		return os.RemoveAll(t.tempDir)
	}
	return nil
}
