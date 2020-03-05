package processor

import (
	"github.com/ledongthuc/pdf"
	"os/exec"
	"strings"
)

// Default way of getting PDF is to call out to pdf2txt which is the better
// option because it actually works
func convertPDFTextPdf2Txt(path string) (string, error) {
	body, err := exec.Command("pdf2txt", path).Output()
	return string(body), err
}

// Fallback to worse but better than nothing attempt
func convertPDFText(path string) (string, error) {
	_, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}

	var str strings.Builder
	for pageIndex := 1; pageIndex <= r.NumPage(); pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		s, err := p.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		str.WriteString(s)
	}

	return str.String(), nil
}
