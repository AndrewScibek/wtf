package rollbar

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/wtfutil/wtf/wtf"
)

const HelpText = `
 Keyboard commands for Travis CI:

   /: Show/hide this help window
   j: Select the next build in the list
   k: Select the previous build in the list
   r: Refresh the data

   arrow down: Select the next build in the list
   arrow up:   Select the previous build in the list

   return: Open the selected build in a browser
`

type Widget struct {
	wtf.HelpfulWidget
	wtf.TextWidget

	items    *Result
	selected int
}

func NewWidget(app *tview.Application, pages *tview.Pages) *Widget {
	widget := Widget{
		HelpfulWidget: wtf.NewHelpfulWidget(app, pages, HelpText),
		TextWidget:    wtf.NewTextWidget(app, "Rollbar", "rollbar", true),
	}
	widget.HelpfulWidget.SetView(widget.View)
	widget.unselect()

	widget.View.SetInputCapture(widget.keyboardIntercept)

	return &widget
}

/* -------------------- Exported Functions -------------------- */

func (widget *Widget) Refresh() {
	if widget.Disabled() {
		return
	}

	items, err := ItemsFor()

	if err != nil {
		widget.View.SetWrap(true)
		widget.View.SetTitle(widget.Name)
		widget.View.SetText(err.Error())
	} else {
		widget.items = &items.Results
	}

	widget.display()
}

/* -------------------- Unexported Functions -------------------- */

func (widget *Widget) display() {
	if widget.items == nil {
		return
	}

	widget.View.SetWrap(false)

	widget.View.SetTitle(widget.ContextualTitle(fmt.Sprintf("%s - Items", widget.Name)))
	widget.View.SetText(widget.contentFrom(widget.items))
}

func (widget *Widget) contentFrom(result *Result) string {
	var str string
	count := wtf.Config.UInt("wtf.mods.rollbar.count", 10)
	if len(result.Items) > count {
		result.Items = result.Items[:count]
	}
	for idx, item := range result.Items {

		str = str + fmt.Sprintf(
			"[%s] [%s] %s %s [%s]count: %d [%s]%s\n",
			widget.rowColor(idx),
			buildColor(&item),
			item.Title,
			item.Level,
			widget.rowColor(idx),
			item.TotalOccurrences,
			"blue",
			item.Environment,
		)
	}

	return str
}

func (widget *Widget) rowColor(idx int) string {
	if widget.View.HasFocus() && (idx == widget.selected) {
		return wtf.DefaultFocussedRowColor()
	}
	return "White"
}

func buildColor(item *Item) string {
	return "red"
}

func (widget *Widget) next() {
	widget.selected++
	if widget.items != nil && widget.selected >= len(widget.items.Items) {
		widget.selected = 0
	}

	widget.display()
}

func (widget *Widget) prev() {
	widget.selected--
	if widget.selected < 0 && widget.items.Items != nil {
		widget.selected = len(widget.items.Items) - 1
	}

	widget.display()
}

func (widget *Widget) openBuild() {
	sel := widget.selected
	userName := wtf.Config.UString("wtf.mods.rollbar.userName", "")
	projectName := wtf.Config.UString("wtf.mods.rollbar.projectName", "")
	if sel >= 0 && widget.items != nil && sel < len(widget.items.Items) {
		item := &widget.items.Items[widget.selected]
		wtf.OpenFile(fmt.Sprintf("https://rollbar.com/%s/%s/%s/%d", userName, projectName, "items", item.ID))
	}
}

func (widget *Widget) unselect() {
	widget.selected = -1
	widget.display()
}

func (widget *Widget) keyboardIntercept(event *tcell.EventKey) *tcell.EventKey {
	switch string(event.Rune()) {
	case "/":
		widget.ShowHelp()
	case "j":
		widget.next()
		return nil
	case "k":
		widget.prev()
		return nil
	case "r":
		widget.Refresh()
		return nil
	}

	switch event.Key() {
	case tcell.KeyDown:
		widget.next()
		return nil
	case tcell.KeyEnter:
		widget.openBuild()
		return nil
	case tcell.KeyEsc:
		widget.unselect()
		return event
	case tcell.KeyUp:
		widget.prev()
		widget.display()
		return nil
	default:
		return event
	}
}
