package app

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazycore/pkg/boxlayout"
)

func (a *App) layoutManager(g *gocui.Gui) error {
	width, height := g.Size()

	root := &boxlayout.Box{
		Direction: boxlayout.ROW,
		Children: []*boxlayout.Box{
			{
				Direction: boxlayout.COLUMN,
				Weight:    1,
				Children: []*boxlayout.Box{
					{
						Direction: boxlayout.ROW,
						Weight:    1,
						Children: []*boxlayout.Box{
							{
								Window: ViewWorkspace,
								Size:   10, // 실제 컨텐츠 길이 확인 후 재조정 필요
							},
							{
								Window: ViewMigrations,
								Weight: 1,
							},
						},
					},
					{
						Direction: boxlayout.ROW,
						Weight:    2,
						Children: []*boxlayout.Box{
							{
								Window: ViewDetails,
								Weight: 3,
							},
							{
								Window: ViewOutputs,
								Weight: 1,
							},
						},
					},
				},
			},
			{
				Window: ViewStatusbar,
				Size:   1,
			},
		},
	}

	// boxlayout으로 차원 계산
	dimensionMap := boxlayout.ArrangeWindows(root, 0, 0, width, height)

	// 각 패널 렌더링
	for id, dim := range dimensionMap {
		if panel, ok := a.panels[id]; ok {
			if err := panel.Draw(dim); err != nil {
				return err
			}
		}
	}

	// Render modal if active (modal is rendered on top of panels)
	if a.activeModal != nil {
		// Modal uses full screen dimensions for positioning
		if err := a.activeModal.Draw(boxlayout.Dimensions{
			X0: 0,
			Y0: 0,
			X1: width,
			Y1: height,
		}); err != nil {
			return err
		}

		// Set focus to modal
		_, err := g.SetCurrentView(a.activeModal.ID())
		if err != nil && err.Error() != "unknown view" {
			// Ignore "unknown view" error
		}
	} else {
		// Set current view for keyboard input (after views are created)
		// This ensures gocui can route keyboard events properly
		if len(a.focusOrder) > 0 && a.currentFocus < len(a.focusOrder) {
			currentViewID := a.focusOrder[a.currentFocus]
			_, err := g.SetCurrentView(currentViewID)
			if err != nil && err.Error() != "unknown view" {
				// Ignore "unknown view" error (happens during initialization)
			}
		}
	}

	return nil
}
