package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dokadev/lazyprisma/pkg/database"
	_ "github.com/dokadev/lazyprisma/pkg/database/drivers" // Register database drivers
	"github.com/dokadev/lazyprisma/pkg/git"
	"github.com/dokadev/lazyprisma/pkg/gui/types"
	"github.com/dokadev/lazyprisma/pkg/i18n"
	"github.com/dokadev/lazyprisma/pkg/node"
	"github.com/dokadev/lazyprisma/pkg/prisma"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

// ANSI styling helpers (self-contained to avoid circular import with app)
func stylize(text string, fg string, bold bool) string {
	if text == "" {
		return text
	}
	codes := ""
	if fg != "" {
		codes = fg
	}
	if bold {
		if codes != "" {
			codes += ";1"
		} else {
			codes = "1"
		}
	}
	if codes == "" {
		return text
	}
	return fmt.Sprintf("\x1b[%sm%s\x1b[0m", codes, text)
}

func yellowBold(text string) string {
	return stylize(text, "33", true)
}

func greenBold(text string) string {
	return stylize(text, "32", true)
}

func wsRed(text string) string {
	if text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[31m%s\x1b[0m", text)
}

func wsRedBold(text string) string {
	return stylize(text, "31", true)
}

func orange(text string) string {
	if text == "" {
		return text
	}
	return fmt.Sprintf("\x1b[38;5;208m%s\x1b[0m", text)
}

// Frame and title styling constants (matching app.panel.go values)
var (
	wsDefaultFrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

	wsPrimaryFrameColor = gocui.ColorWhite
	wsFocusedFrameColor = gocui.ColorGreen

	wsPrimaryTitleColor = gocui.ColorWhite | gocui.AttrNone
	wsFocusedTitleColor = gocui.ColorGreen | gocui.AttrBold
)

type WorkspaceContext struct {
	*SimpleContext
	*ScrollableTrait

	g              *gocui.Gui
	tr             *i18n.TranslationSet
	focused        bool
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
}

var _ types.Context = &WorkspaceContext{}
var _ types.IScrollableContext = &WorkspaceContext{}

type WorkspaceContextOpts struct {
	Gui      *gocui.Gui
	Tr       *i18n.TranslationSet
	ViewName string
}

func NewWorkspaceContext(opts WorkspaceContextOpts) *WorkspaceContext {
	baseCtx := NewBaseContext(BaseContextOpts{
		Key:       types.ContextKey(opts.ViewName),
		Kind:      types.SIDE_CONTEXT,
		ViewName:  opts.ViewName,
		Focusable: true,
		Title:     opts.Tr.PanelTitleWorkspace,
	})

	simpleCtx := NewSimpleContext(baseCtx)

	wc := &WorkspaceContext{
		SimpleContext:   simpleCtx,
		ScrollableTrait: &ScrollableTrait{},
		g:               opts.Gui,
		tr:              opts.Tr,
		showMasked:      true, // Default to masked
	}

	wc.loadVersionInfo()

	return wc
}

// ID returns the view identifier (implements Panel interface from app package)
func (w *WorkspaceContext) ID() string {
	return w.GetViewName()
}

// Draw renders the workspace panel (implements Panel interface from app package)
func (w *WorkspaceContext) Draw(dim boxlayout.Dimensions) error {
	v, err := w.g.SetView(w.GetViewName(), dim.X0, dim.Y0, dim.X1, dim.Y1, 0)
	if err != nil && err.Error() != "unknown view" {
		return err
	}

	// Setup view (replicates BasePanel.SetupView)
	w.setupView(v)
	w.SetView(v)                 // BaseContext
	w.ScrollableTrait.SetView(v) // ScrollableTrait

	v.Wrap = true // Enable word wrap

	// Build content from fields
	var lines []string

	// Node and Prisma version on one line
	nodeVersionStyled := yellowBold(w.nodeVersion)
	prismaVersionStyled := yellowBold(w.prismaVersion)
	versionLine := fmt.Sprintf(w.tr.WorkspaceVersionLine, nodeVersionStyled, prismaVersionStyled)
	if w.prismaGlobal {
		versionLine += " " + orange(w.tr.WorkspacePrismaGlobalIndicator)
	}
	lines = append(lines, versionLine)

	// Git info
	lines = append(lines, "")
	if w.isGitRepo {
		// Git line with optional schema modified indicator
		gitLine := fmt.Sprintf(w.tr.WorkspaceGitLine, w.gitRepoName)
		if w.schemaModified {
			gitLine += " " + orange(w.tr.WorkspaceSchemaModifiedIndicator)
		}
		lines = append(lines, gitLine)

		// Branch on separate line
		branchStyled := yellowBold(w.gitBranch)
		lines = append(lines, fmt.Sprintf(w.tr.WorkspaceBranchFormat, branchStyled))
	} else {
		lines = append(lines, w.tr.WorkspaceNotGitRepository)
	}

	lines = append(lines, "")
	lines = append(lines, w.buildDatabaseLines()...)

	content := ""
	for _, line := range lines {
		content += line + "\n"
	}

	fmt.Fprint(v, content)

	// Adjust scroll and apply origin
	w.ScrollableTrait.AdjustScroll()

	return nil
}

// setupView configures the view with common settings (replaces BasePanel.SetupView)
func (w *WorkspaceContext) setupView(v *gocui.View) {
	v.Clear()
	v.Frame = true
	v.Title = w.tr.PanelTitleWorkspace
	v.FrameRunes = wsDefaultFrameRunes

	if w.focused {
		v.FrameColor = wsFocusedFrameColor
		v.TitleColor = wsFocusedTitleColor
	} else {
		v.FrameColor = wsPrimaryFrameColor
		v.TitleColor = wsPrimaryTitleColor
	}
}

// OnFocus handles focus gain (implements Panel interface from app package)
func (w *WorkspaceContext) OnFocus() {
	w.focused = true
	if v := w.GetView(); v != nil {
		v.FrameColor = wsFocusedFrameColor
		v.TitleColor = wsFocusedTitleColor
	}
}

// OnBlur handles focus loss (implements Panel interface from app package)
func (w *WorkspaceContext) OnBlur() {
	w.focused = false
	if v := w.GetView(); v != nil {
		v.FrameColor = wsPrimaryFrameColor
		v.TitleColor = wsPrimaryTitleColor
	}
}

// Refresh reloads all workspace information
func (w *WorkspaceContext) Refresh() {
	// Save current scroll position
	currentOriginY := w.ScrollableTrait.GetOriginY()

	// Reload information
	w.loadVersionInfo()
	w.loadDatabaseInfo()

	// Restore scroll position (will be adjusted by AdjustScroll in Draw if needed)
	w.ScrollableTrait.SetOriginY(currentOriginY)
}

// ScrollUpByWheel scrolls up by wheel increment (delegates to ScrollableTrait)
func (w *WorkspaceContext) ScrollUpByWheel() {
	w.ScrollableTrait.ScrollUpByWheel()
}

// ScrollDownByWheel scrolls down by wheel increment (delegates to ScrollableTrait)
func (w *WorkspaceContext) ScrollDownByWheel() {
	w.ScrollableTrait.ScrollDownByWheel()
}

func (w *WorkspaceContext) loadVersionInfo() {
	cwd, _ := os.Getwd()

	// Node version
	if nodeVer, err := node.GetVersion(); err == nil {
		w.nodeVersion = nodeVer.Version
	} else {
		w.nodeVersion = w.tr.WorkspaceVersionNotFound
	}

	// Prisma version
	if prismaVer, err := prisma.GetVersion(cwd); err == nil {
		w.prismaVersion = prismaVer.Version
		w.prismaGlobal = prismaVer.IsGlobal
	} else {
		w.prismaVersion = w.tr.WorkspaceVersionNotFound
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

func (w *WorkspaceContext) loadDatabaseInfo() {
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
		w.dbError = w.tr.WorkspaceErrorGetWorkingDirectory
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
			w.dbError = w.tr.WorkspaceErrorSchemaNotFound
		} else if strings.Contains(errMsg, "incomplete") {
			// Store plain text, styling will be applied in buildDatabaseLines()
			if w.envVarName != "" {
				w.dbError = w.envVarName + w.tr.WorkspaceNotConfiguredSuffix
			} else {
				w.dbError = w.tr.WorkspaceDatabaseURLNotConfigured
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
			w.dbError = w.envVarName + w.tr.WorkspaceNotConfiguredSuffix
		} else {
			w.dbError = w.tr.WorkspaceNoDatabaseURL
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

func (w *WorkspaceContext) buildDatabaseLines() []string {
	var lines []string

	// Display provider with status on the same line
	providerName := database.GetProviderDisplayName(w.dbProvider)
	providerName = yellowBold(providerName)

	// Build provider line with status
	var providerLine string
	if w.dbConnected {
		statusStyled := greenBold(w.tr.WorkspaceConnected)
		providerLine = fmt.Sprintf(w.tr.WorkspaceProviderLine, providerName, statusStyled)
	} else if w.dbError != "" {
		if w.isConfigurationError() {
			statusStyled := wsRedBold(w.tr.WorkspaceNotConfigured)
			providerLine = fmt.Sprintf(w.tr.WorkspaceProviderLine, providerName, statusStyled)
		} else {
			statusStyled := wsRedBold(w.tr.WorkspaceDisconnected)
			providerLine = fmt.Sprintf(w.tr.WorkspaceProviderLine, providerName, statusStyled)
		}
	} else {
		statusStyled := wsRedBold(w.tr.WorkspaceDisconnected)
		providerLine = fmt.Sprintf(w.tr.WorkspaceProviderLine, providerName, statusStyled)
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
			lines = append(lines, fmt.Sprintf("%s %s", displayURL, wsRed(w.tr.WorkspaceHardcodedIndicator)))
		} else {
			lines = append(lines, displayURL)
		}
	} else if w.dbError != "" && w.isConfigurationError() {
		// Only show error in URL field if it's a configuration issue
		// Apply styling: bold+red env var name, red "not configured"
		if w.envVarName != "" && strings.Contains(w.dbError, w.tr.WorkspaceNotConfiguredSuffix) {
			styledError := wsRedBold(w.envVarName) + wsRed(w.tr.WorkspaceNotConfiguredSuffix)
			lines = append(lines, styledError)
		} else {
			lines = append(lines, wsRed(w.dbError))
		}
	} else {
		lines = append(lines, w.tr.WorkspaceNotSet)
	}

	// Show detailed error message if disconnected (not configuration error)
	if !w.dbConnected && w.dbError != "" && !w.isConfigurationError() {
		lines = append(lines, wsRed(fmt.Sprintf(w.tr.WorkspaceErrorFormat, w.dbError)))
	}

	return lines
}

// isConfigurationError checks if the error is a configuration issue
func (w *WorkspaceContext) isConfigurationError() bool {
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
