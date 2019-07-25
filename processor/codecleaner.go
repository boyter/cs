package processor

import "strings"


var firstReplacements = []string{"<", ">", ")", "(", "[", "]", "|", "=", ",", ":"};
var secondReplacements = []string{";", "{", "}", "/"}
var thirdReplacements = []string{"\"", "'"}
var forthReplacements = []string{".", ";", "=", "_", ";", "@", "#", "-", "<", ">"}

func codeCleaner(inputCode string) (string) {
	var indexContents strings.Builder
	modifiedContents := strings.ToLower(inputCode)
	modifiedContents = replaceReplacer(modifiedContents)

	for _, x := range firstReplacements {
		modifiedContents = strings.Replace(modifiedContents, x, " ",-1)
	}
	indexContents.WriteString(modifiedContents)

	modifiedContents = strings.Replace(modifiedContents, ".", " ",-1)
	indexContents.WriteString(modifiedContents)

	for _, x := range secondReplacements {
		modifiedContents = strings.Replace(modifiedContents, x, " ",-1)
	}
	indexContents.WriteString(modifiedContents)

	for _, x := range thirdReplacements {
		modifiedContents = strings.Replace(modifiedContents, x, " ",-1)
	}
	indexContents.WriteString(modifiedContents)

	for _, x := range forthReplacements {
		modifiedContents = strings.Replace(modifiedContents, x, " ",-1)
	}
	indexContents.WriteString(modifiedContents)

	// replace multiple newlines
	return strings.Join(strings.Fields(strings.TrimSpace(indexContents.String())), " ")
}

const replacement = " "
var replacer = strings.NewReplacer(
	"\r\n", replacement,
	"\r", replacement,
	"\n", replacement,
	"\v", replacement,
	"\f", replacement,
	"\u0085", replacement,
	"\u2028", replacement,
	"\u2029", replacement,
)

func replaceReplacer(s string) string {
	return replacer.Replace(s)
}