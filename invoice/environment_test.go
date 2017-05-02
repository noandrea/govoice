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

