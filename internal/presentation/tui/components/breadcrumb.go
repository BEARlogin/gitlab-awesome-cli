package components

import (
	"strings"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

type Breadcrumb struct {
	Parts []string
}

func NewBreadcrumb() Breadcrumb {
	return Breadcrumb{}
}

func (b Breadcrumb) View() string {
	if len(b.Parts) == 0 {
		return styles.Title.Render("glcli")
	}
	return styles.Title.Render("glcli > " + strings.Join(b.Parts, " > "))
}
