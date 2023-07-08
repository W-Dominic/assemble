package main

import (
  "flag"
  "fmt"
  "os"
  "strings"
  "assemble/assembler"
  "time"

  tea "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/bubbles/viewport"
  "github.com/charmbracelet/lipgloss"
)
const useHighPerformanceRenderer = false

var (
  titleStyle = func() lipgloss.Style {
    b := lipgloss.RoundedBorder()
    b.Right = "├"
    return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
  }()

  infoStyle = func() lipgloss.Style {
    b := lipgloss.RoundedBorder()
    b.Left = "┤"
    return titleStyle.Copy().BorderStyle(b)
  }()
)

type assemblyMsg struct{
  code string
  err error
}

func createContent(sub chan assemblyMsg, filePath *string) tea.Cmd {
  return func() tea.Msg {
    var oldSize int64 = 0
    for {
      fileInfo, err := os.Stat(*filePath)
      if err != nil {
        sub <- assemblyMsg{ code: fmt.Sprintf("%v\n", err), err: err }
      }   
      newSize := fileInfo.Size()
      if newSize != oldSize {
        oldSize = newSize
        content, err := assembler.Assemble(filePath)
        if err != nil {
          sub <- assemblyMsg{ code: fmt.Sprintf("%v\n", err), err: err }
        }
        if content != "" { // change detected
          sub <- assemblyMsg{ code: content, err: nil }
        }
      }
      time.Sleep(1 * time.Second)
    }
  }
}

func waitForContent(sub chan assemblyMsg) tea.Cmd {
  return func() tea.Msg {
    return assemblyMsg(<-sub)
  }
}

type model struct{
  sub chan assemblyMsg
  content string
  ready bool
  viewport viewport.Model
  filePath *string
}

func (m model) Init() tea.Cmd {
  return tea.Batch(
    waitForContent(m.sub),
    createContent(m.sub, m.filePath),
  )
}

func (m model) Update (msg tea.Msg) (tea.Model, tea.Cmd) {
  var (
    cmd  tea.Cmd
    cmds []tea.Cmd
  )

  switch msg := msg.(type) {
  case tea.KeyMsg:
    if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}

  case assemblyMsg:
    fmt.Printf("Msg found")
      m.content = msg.code
      cmds = append(cmds, waitForContent(m.sub))

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.viewport.SetContent(m.content)
			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
  content, err := assembler.Assemble(m.filePath)
  if err != nil {
    return fmt.Sprintf("\n %v", err)
  }
	if !m.ready {
		return "\n  Initializing..."
	}

  // update model with new content
  if content != "" {
    m.content = content
  } else {
    return "Content blank"
  }

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m model) headerView() string {
	title := titleStyle.Render("x86 Assembly")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
  filePath := flag.String("f", "./", "Path to file")
  flag.Parse() 
  if *filePath == "./" {
    fmt.Printf("Cannot open file \n")
    os.Exit(1)
  }	// Load some text for our viewport

	p := tea.NewProgram(
		model{
      content: "hello",
      sub: make(chan assemblyMsg),
      filePath: filePath,
    },
		tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}  
