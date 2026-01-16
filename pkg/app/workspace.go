package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokadev/lazyprisma/pkg/database"
	_ "github.com/dokadev/lazyprisma/pkg/database/drivers" // Register database drivers
	"github.com/dokadev/lazyprisma/pkg/git"
	"github.com/dokadev/lazyprisma/pkg/node"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

type WorkspacePanel struct {
	BasePanel
	nodeVersion    string
	prismaVersion  string
	prismaGlobal   bool
	gitRepoName    string // Git repository name
	gitBranch      string // Git branch name
	isGitRepo      bool   // True if current directory is a git repository
	schemaModified bool   // True if schema.prisma has git changes
	unmaskedURL    string
	maskedURL      string
	showMasked     bool
	dbProvider     string
	dbConnected    bool
	dbError        string
	envVarName     string // Environment variable name (e.g., "DATABASE_URL")
	isHardcoded    bool   // True if URL is hardcoded in schema/config
	originY        int    // Scroll position
}

func NewWorkspacePanel(g *gocui.Gui) *WorkspacePanel {
	wp := &WorkspacePanel{
		BasePanel:  NewBasePanel(ViewWorkspace, g),
		showMasked: true, // Default to masked
	}
	wp.loadVersionInfo()
	return wp
}

func (w *WorkspacePanel) loadVersionInfo() {
	cwd, _ := os.Getwd()

	// Node version
	if nodeVer, err := node.GetVersion(); err == nil {
		w.nodeVersion = nodeVer.Version
	} else {
		w.nodeVersion = "Not found"
	}

	// Prisma version
	if prismaVer, err := prisma.GetVersion(cwd); err == nil {
		w.prismaVersion = prismaVer.Version
		w.prismaGlobal = prismaVer.IsGlobal
	} else {
		w.prismaVersion = "Not found"
		w.prismaGlobal = false
	}

	// Git info
	gitInfo := git.GetGitInfo(cwd)
	w.isGitRepo = gitInfo.IsRepository
	w.gitRepoName = gitInfo.RepositoryName
	w.gitBranch = gitInfo.BranchName

	// Check schema.prisma modification status (only if git repo)
	if w.isGitRepo {
		schemaPath := filepath.Join(cwd, prisma.SchemaDirName, prisma.SchemaFileName)
		w.schemaModified = git.IsFileModified(cwd, schemaPath)
	} else {
		w.schemaModified = false
	}

	// Load database info
	w.loadDatabaseInfo()
}

func (w *WorkspacePanel) buildDatabaseLines() []string {
	var lines []string

	// Display provider with status on the same line
	providerName := database.GetProviderDisplayName(w.dbProvider)
	if providerName == "Unknown" {
		providerName = Stylize(providerName, Style{FgColor: ColorYellow, Bold: true})
	} else {
		providerName = Stylize(providerName, Style{FgColor: ColorYellow, Bold: true})
	}

	// Build provider line with status
	var providerLine string
	if w.dbConnected {
		statusStyled := Stylize("✓ Connected", Style{FgColor: ColorGreen, Bold: true})
		providerLine = fmt.Sprintf("Provider: %s  %s", providerName, statusStyled)
	} else if w.dbError != "" {
		if w.isConfigurationError() {
			statusStyled := Stylize("✗ Not configured", Style{FgColor: ColorRed, Bold: true})
			providerLine = fmt.Sprintf("Provider: %s  %s", providerName, statusStyled)
		} else {
			statusStyled := Stylize("✗ Disconnected", Style{FgColor: ColorRed, Bold: true})
			providerLine = fmt.Sprintf("Provider: %s  %s", providerName, statusStyled)
		}
	} else {
		statusStyled := Stylize("✗ Disconnected", Style{FgColor: ColorRed, Bold: true})
		providerLine = fmt.Sprintf("Provider: %s  %s", providerName, statusStyled)
	}
	lines = append(lines, providerLine)

	// Display URL (always show if available)
	if w.unmaskedURL != "" {
		displayURL := w.maskedURL
		if !w.showMasked {
			displayURL = w.unmaskedURL
		}

		// Add hardcoded warning if applicable
		if w.isHardcoded {
			lines = append(lines, fmt.Sprintf("%s %s", displayURL, Red("(Hard coded)")))
		} else {
			lines = append(lines, displayURL)
		}
	} else if w.dbError != "" && w.isConfigurationError() {
		// Only show error in URL field if it's a configuration issue
		// Apply styling: bold+red env var name, red "not configured"
		if w.envVarName != "" && strings.Contains(w.dbError, " not configured") {
			styledError := Stylize(w.envVarName, Style{FgColor: ColorRed, Bold: true}) + Red(" not configured")
			lines = append(lines, styledError)
		} else {
			lines = append(lines, Red(w.dbError))
		}
	} else {
		lines = append(lines, "Not set")
	}

	// Show detailed error message if disconnected (not configuration error)
	if !w.dbConnected && w.dbError != "" && !w.isConfigurationError() {
		lines = append(lines, Red(fmt.Sprintf("Error: %s", w.dbError)))
	}

	return lines
}

// isConfigurationError checks if the error is a configuration issue
func (w *WorkspacePanel) isConfigurationError() bool {
	if w.dbError == "" {
		return false
	}

	configErrors := []string{
		"not found",
		"not configured",
		"not set",
		"incomplete",
		"no database_url",
	}

	errLower := strings.ToLower(w.dbError)
	for _, substr := range configErrors {
		if strings.Contains(errLower, substr) {
			return true
		}
	}
	return false
}

func (w *WorkspacePanel) loadDatabaseInfo() {
	// Reset fields
	w.dbProvider = ""
	w.unmaskedURL = ""
	w.maskedURL = ""
	w.dbConnected = false
	w.dbError = ""
	w.envVarName = ""
	w.isHardcoded = false

	cwd, err := os.Getwd()
	if err != nil {
		w.dbError = "Error getting working directory"
		return
	}

	// Get datasource from schema
	ds, err := prisma.GetDatasource(cwd)
	if err != nil {
		// Try to extract provider only (even if URL resolution fails)
		if provider, err2 := prisma.GetProvider(cwd); err2 == nil {
			w.dbProvider = provider
		}

		// Try to extract env var name (even if resolution fails)
		if envVar, err2 := prisma.GetEnvVarName(cwd); err2 == nil {
			w.envVarName = envVar
		}

		// Categorize error message for better user understanding
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			w.dbError = "Schema file not found"
		} else if strings.Contains(errMsg, "incomplete") {
			// Store plain text, styling will be applied in buildDatabaseLines()
			if w.envVarName != "" {
				w.dbError = w.envVarName + " not configured"
			} else {
				w.dbError = "DATABASE_URL not configured"
			}
		} else {
			w.dbError = errMsg
		}
		return
	}

	// Store provider, URLs, and metadata
	w.dbProvider = ds.Provider
	w.unmaskedURL = ds.URL
	w.maskedURL = prisma.MaskPassword(ds.URL)
	w.envVarName = ds.EnvVarName
	w.isHardcoded = ds.IsHardcoded

	// Try to connect to database
	if ds.URL == "" {
		if w.envVarName != "" {
			w.dbError = w.envVarName + " not configured"
		} else {
			w.dbError = "No DATABASE_URL"
		}
		return
	}

	// Attempt connection
	client, err := database.NewClientFromDSN(ds.Provider, ds.URL)
	if err != nil {
		w.dbError = err.Error()
		return
	}
	defer client.Close()

	// Test connection with ping
	if err := client.Ping(); err != nil {
		w.dbError = err.Error()
		return
	}

	// Connection successful
	w.dbConnected = true
}

func (w *WorkspacePanel) Draw(dim boxlayout.Dimensions) error {
	v, err := w.g.SetView(w.id, dim.X0, dim.Y0, dim.X1, dim.Y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	w.SetupView(v, "Workspace")
	w.v = v
	v.Wrap = true // Enable word wrap

	// Build content from fields
	var lines []string

	// Node and Prisma version on one line
	nodeVersionStyled := Stylize(w.nodeVersion, Style{FgColor: ColorYellow, Bold: true})
	prismaVersionStyled := Stylize(w.prismaVersion, Style{FgColor: ColorYellow, Bold: true})
	versionLine := fmt.Sprintf("Node: %s | Prisma: %s", nodeVersionStyled, prismaVersionStyled)
	if w.prismaGlobal {
		versionLine += " " + Orange("(Global)")
	}
	lines = append(lines, versionLine)

	// Git info
	lines = append(lines, "")
	if w.isGitRepo {
		// Git line with optional schema modified indicator
		gitLine := fmt.Sprintf("Git: %s", w.gitRepoName)
		if w.schemaModified {
			gitLine += " " + Orange("(schema modified)")
		}
		lines = append(lines, gitLine)

		// Branch on separate line
		branchStyled := Stylize(w.gitBranch, Style{FgColor: ColorYellow, Bold: true})
		lines = append(lines, fmt.Sprintf("(%s)", branchStyled))
	} else {
		lines = append(lines, "Git: Not a git repository")
	}

	lines = append(lines, "")
	lines = append(lines, w.buildDatabaseLines()...)

	content := ""
	for _, line := range lines {
		content += line + "\n"
	}

	fmt.Fprint(v, content)

	// Adjust origin to ensure it's within valid bounds
	AdjustOrigin(v, &w.originY)
	v.SetOrigin(0, w.originY)

	return nil
}

// ScrollUp scrolls the workspace panel up
func (w *WorkspacePanel) ScrollUp() {
	if w.originY > 0 {
		w.originY--
	}
}

// ScrollDown scrolls the workspace panel down
func (w *WorkspacePanel) ScrollDown() {
	w.originY++
	// AdjustOrigin will be called in Draw() to ensure bounds
}

// ScrollUpByWheel scrolls the workspace panel up by 2 lines (mouse wheel)
func (w *WorkspacePanel) ScrollUpByWheel() {
	if w.originY > 0 {
		w.originY -= 2
		if w.originY < 0 {
			w.originY = 0
		}
	}
}

// ScrollDownByWheel scrolls the workspace panel down by 2 lines (mouse wheel)
func (w *WorkspacePanel) ScrollDownByWheel() {
	if w.v == nil {
		return
	}

	// Get actual content lines from the rendered view buffer
	contentLines := len(w.v.ViewBufferLines())
	_, viewHeight := w.v.Size()
	innerHeight := viewHeight - 2 // Exclude frame (top + bottom)

	// Calculate maxOrigin
	maxOrigin := contentLines - innerHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}

	// Only scroll if we haven't reached the bottom
	if w.originY < maxOrigin {
		w.originY += 2
		if w.originY > maxOrigin {
			w.originY = maxOrigin
		}
	}
}

// ScrollToTop scrolls to the top of the workspace panel
func (w *WorkspacePanel) ScrollToTop() {
	w.originY = 0
}

// ScrollToBottom scrolls to the bottom of the workspace panel
func (w *WorkspacePanel) ScrollToBottom() {
	if w.v == nil {
		return
	}

	// Get actual content lines from the rendered view buffer
	contentLines := len(w.v.ViewBufferLines())
	_, viewHeight := w.v.Size()
	innerHeight := viewHeight - 2 // Exclude frame (top + bottom)

	// Calculate maxOrigin
	maxOrigin := contentLines - innerHeight
	if maxOrigin < 0 {
		maxOrigin = 0
	}

	w.originY = maxOrigin
}

// Refresh reloads all workspace information
func (w *WorkspacePanel) Refresh() {
	// Save current scroll position
	currentOriginY := w.originY

	// Reload information
	w.loadVersionInfo()
	w.loadDatabaseInfo()

	// Restore scroll position (will be adjusted by AdjustOrigin in Draw if needed)
	w.originY = currentOriginY
}
