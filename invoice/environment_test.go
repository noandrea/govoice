package invoice

import (
	"os"
	"path"
	"testing"
)

func TestWorkspacePaths(t *testing.T) {

	workspaceBase := os.TempDir()
	workspaceName := "govoice_workspace"

	workspace := path.Join(workspaceBase, workspaceName)
	c := Config{
		Workspace:      workspace,
		MasterTemplate: "_master",
		Layout: Layout{
			Style:    Style{Margins{0, 20, 20, 10}, "helvetica", 8, 14, 16, 6, 3.7, 6, 4, 3, 60, 13, 13, 13, 8, 6},
			Items:    Block{Coords{-1, 100}},
			From:     Block{Coords{-1, 28}},
			To:       Block{Coords{-1, 60}},
			Invoice:  Block{Coords{140, 28}},
			Payments: Block{Coords{-1, 210}},
			Notes:    Block{Coords{-1, 240}},
		},
	}
	// pdf
	filePath, exists := c.GetInvoicePdfPath("0000")

	if exists == true {
		t.Error("path ", filePath, "should not exists") // dumb test
	}

	if fp := path.Join(workspaceBase, workspaceName, "0000.pdf"); filePath != fp {
		t.Error("Expected\n", fp, "\nfound\n", filePath)
	}
	// json
	filePath, exists = c.GetInvoiceJsonPath("0000")
	if exists == true {
		t.Error("path ", filePath, "should not exists") // dumb test
	}

	if fp := path.Join(workspaceBase, workspaceName, "0000.json.cfb"); filePath != fp {
		t.Error("Expected\n", fp, "\nfound\n", filePath)
	}
}

func TestSetup(t *testing.T) {
	tmpDir := os.TempDir()
	tmpHome := path.Join(tmpDir, "govoice")
	tmpWorkspace := path.Join(tmpDir, "workspace")

	os.Setenv("HOME", tmpHome)
	// check if config path and master path are set
	configPath, masterPath, err := Setup(tmpWorkspace)

	if err != nil {
		t.Error("error Setup", err)
	}

	if p := path.Join(tmpHome, ".govoice", "config.toml"); configPath != p {
		t.Error("Expected", p, "found", configPath)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Error("file", p, "does not exists")
		}
	}

	if p := path.Join(tmpWorkspace, "_master.json"); masterPath != p {
		t.Error("Expected", p, "found", masterPath)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Error("file", p, "does not exists")
		}
	}

	os.RemoveAll(tmpHome)
	os.RemoveAll(tmpWorkspace)
}
