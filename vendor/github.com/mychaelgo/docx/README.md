## Simple Google Go (golang) library for replace text in microsoft word (.docx) file

The following constitutes the bare minimum required to replace text in DOCX document.
``` go

import (
	"github.com/nguyenthenguyen/docx"
)

func main() {
	r, err := docx.ReadDocxFile("./template.docx")
	if err != nil {
		panic(err)
	}
	docx1 := r.Editable()
	// Replace like https://golang.org/pkg/strings/#Replace
	docx1.Replace("old_1_1", "new_1_1", -1)
	docx1.Replace("old_1_2", "new_1_2", -1)
	docx1.ReplaceLink("http://example.com/", "https://github.com/nguyenthenguyen/docx")
	docx1.ReplaceHeader("out with the old", "in with the new")
	docx1.ReplaceFooter("Change This Footer", "new footer")
	docx1.WriteToFile("./new_result_1.docx")

	docx2 := r.Editable()
	docx2.Replace("old_2_1", "new_2_1", -1)
	docx2.Replace("old_2_2", "new_2_2", -1)
	docx2.WriteToFile("./new_result_2.docx")

	r.Close()
}

```
