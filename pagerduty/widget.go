package pagerduty

import (
	"fmt"
	"sort"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/rivo/tview"
	"github.com/wtfutil/wtf/wtf"
)

type Widget struct {
	wtf.TextWidget
}

func NewWidget(app *tview.Application) *Widget {
	widget := Widget{
		TextWidget: wtf.NewTextWidget(app, "PagerDuty", "pagerduty", false),
	}

	return &widget
}

/* -------------------- Exported Functions -------------------- */

func (widget *Widget) Refresh() {
	var onCalls []pagerduty.OnCall
	var incidents []pagerduty.Incident

	var err1 error
	var err2 error

	if wtf.Config.UBool("wtf.mods.pagerduty.showSchedules", true) {
		onCalls, err1 = GetOnCalls()
	}

	if wtf.Config.UBool("wtf.mods.pagerduty.showIncidents") {
		incidents, err2 = GetIncidents()
	}

	widget.View.SetTitle(widget.ContextualTitle(fmt.Sprintf("%s", widget.Name)))
	widget.View.Clear()

	var content string
	if err1 != nil || err2 != nil {
		widget.View.SetWrap(true)
		if err1 != nil {
			content = content + err1.Error()
		}
		if err2 != nil {
			content = content + err2.Error()
		}
	} else {
		widget.View.SetWrap(false)
		content = widget.contentFrom(onCalls, incidents)
	}

	widget.View.SetText(content)
}

/* -------------------- Unexported Functions -------------------- */

func (widget *Widget) contentFrom(onCalls []pagerduty.OnCall, incidents []pagerduty.Incident) string {
	var str string

	if len(incidents) > 0 {
		str = str + "[yellow]Incidents[white]\n"
		for _, incident := range incidents {
			str = str + fmt.Sprintf("[red]%s[white]\n", incident.Summary)
			str = str + fmt.Sprintf("Status: %s\n", incident.Status)
			str = str + fmt.Sprintf("Service: %s\n", incident.Service.Summary)
			str = str + fmt.Sprintf("Escalation: %s\n", incident.EscalationPolicy.Summary)
		}
	}

	tree := make(map[string][]pagerduty.OnCall)

	filtering := wtf.Config.UList("wtf.mods.pagerduty.escalationFilter")
	filter := make(map[string]bool)
	for _, item := range filtering {
		filter[item.(string)] = true
	}

	for _, onCall := range onCalls {
		key := onCall.EscalationPolicy.Summary
		if len(filtering) == 0 || filter[key] {
			tree[key] = append(tree[key], onCall)
		}
	}

	// We want to sort our escalation policies for predictability/ease of finding
	keys := make([]string, 0, len(tree))
	for k := range tree {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	if len(keys) > 0 {
		str = str + "[yellow]Schedules[white]\n"
		// Print out policies, and escalation order of users
		for _, key := range keys {
			str = str + fmt.Sprintf("[red]%s\n", key)
			values := tree[key]
			sort.Sort(ByEscalationLevel(values))
			for _, item := range values {
				str = str + fmt.Sprintf("[white]%d - %s\n", item.EscalationLevel, item.User.Summary)
			}
		}
	}

	return str
}
