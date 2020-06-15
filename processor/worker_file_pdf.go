// SPDX-License-Identifier: MIT OR Unlicense

package processor

import (
	"os/exec"
)

// Default way of getting PDF is to call out to pdf2txt which is the better
// option because it actually works
func convertPDFTextPdf2Txt(path string) (string, error) {
	body, err := exec.Command("pdf2txt", path).Output()
	if err != nil {
		body, err = exec.Command("pdftotext", path).Output()
	}
	return string(body), err
}
