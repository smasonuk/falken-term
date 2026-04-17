package chat

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	submit      key.Binding
	interrupt   key.Binding
	quitSlash   key.Binding
	quitEsc     key.Binding
	toolDetails key.Binding
	commands    key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		interrupt: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "interrupt"),
		),
		quitSlash: key.NewBinding(
			key.WithKeys("/exit"),
			key.WithHelp("/exit", "quit"),
		),
		quitEsc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "quit"),
		),
		toolDetails: key.NewBinding(
			key.WithKeys("ctrl+i"),
			key.WithHelp("ctrl+i", "tool details"),
		),
		commands: key.NewBinding(
			key.WithKeys("/help"),
			key.WithHelp("/help", "commands"),
		),
	}
}

func (k keyMap) forPrompt() keyMap {
	return keyMap{
		submit:    k.submit,
		commands:  k.commands,
		quitSlash: k.quitSlash,
		quitEsc:   k.quitEsc,
	}
}

func (k keyMap) forState(state State) keyMap {
	switch state {
	case StateRunning:
		return keyMap{
			interrupt:   k.interrupt,
			toolDetails: k.toolDetails,
			commands:    k.commands,
			quitSlash:   k.quitSlash,
		}
	case StateDone:
		return keyMap{
			submit:      k.submit,
			toolDetails: k.toolDetails,
			commands:    k.commands,
			quitSlash:   k.quitSlash,
		}
	default:
		return k.forPrompt()
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	bindings := make([]key.Binding, 0, 4)
	for _, binding := range []key.Binding{k.submit, k.interrupt, k.toolDetails, k.commands, k.quitSlash, k.quitEsc} {
		if len(binding.Keys()) == 0 || !binding.Enabled() {
			continue
		}
		bindings = append(bindings, binding)
	}
	return bindings
}

func (k keyMap) FullHelp() [][]key.Binding {
	short := k.ShortHelp()
	if len(short) == 0 {
		return nil
	}
	return [][]key.Binding{short}
}
