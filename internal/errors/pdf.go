package errors

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	pageWidth     = 612.0
	pageHeight    = 792.0
	marginX       = 40.0
	marginTop     = 44.0
	marginBottom  = 44.0
	fontSize      = 10.0
	leading       = 13.0
	charsPerLine  = 92
)

type Section struct {
	Heading string
	Lines   []string
}

func wrapLine(line string, width int) []string {
	if len(line) <= width {
		return []string{line}
	}
	var out []string
	for len(line) > width {
		cut := width
		out = append(out, line[:cut])
		line = line[cut:]
	}
	if len(line) > 0 {
		out = append(out, line)
	}
	return out
}

func escapePDFString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	var b strings.Builder
	for _, r := range s {
		if r > 126 {
			b.WriteRune('?')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func buildAllLines(title string, sections []Section) []string {
	var all []string
	all = append(all, title)
	all = append(all, "")
	for _, sec := range sections {
		if sec.Heading != "" {
			all = append(all, sec.Heading)
			all = append(all, strings.Repeat("-", len(sec.Heading)))
		}
		for _, line := range sec.Lines {
			wrapped := wrapLine(line, charsPerLine)
			all = append(all, wrapped...)
		}
		all = append(all, "")
	}
	return all
}

func paginate(lines []string) [][]string {
	usableHeight := pageHeight - marginTop - marginBottom
	linesPerPage := int(usableHeight/leading) - 1
	if linesPerPage < 1 {
		linesPerPage = 40
	}
	var pages [][]string
	for i := 0; i < len(lines); i += linesPerPage {
		end := i + linesPerPage
		if end > len(lines) {
			end = len(lines)
		}
		pages = append(pages, lines[i:end])
	}
	if len(pages) == 0 {
		pages = [][]string{{""}}
	}
	return pages
}

func contentStreamForPage(lines []string) string {
	var b strings.Builder
	b.WriteString("BT\n")
	b.WriteString(fmt.Sprintf("/F1 %.1f Tf\n", fontSize))
	b.WriteString(fmt.Sprintf("%.1f TL\n", leading))
	b.WriteString(fmt.Sprintf("%.1f %.1f Td\n", marginX, pageHeight-marginTop))
	for i, line := range lines {
		escaped := escapePDFString(line)
		b.WriteString(fmt.Sprintf("(%s) Tj\n", escaped))
		if i != len(lines)-1 {
			b.WriteString("T*\n")
		}
	}
	b.WriteString("ET\n")
	return b.String()
}

func GeneratePDF(title string, sections []Section) []byte {
	allLines := buildAllLines(title, sections)
	pages := paginate(allLines)

	var buf bytes.Buffer
	offsets := map[int]int{}

	buf.WriteString("%PDF-1.4\n")

	writeObj := func(id int, body string) {
		offsets[id] = buf.Len()
		buf.WriteString(fmt.Sprintf("%d 0 obj\n", id))
		buf.WriteString(body)
		buf.WriteString("\nendobj\n")
	}

	numPages := len(pages)
	fontObjID := 3
	firstContentID := 4 + numPages

	kids := make([]string, numPages)
	for i := 0; i < numPages; i++ {
		kids[i] = fmt.Sprintf("%d 0 R", 4+i)
	}

	writeObj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObj(2, fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), numPages))
	writeObj(fontObjID, "<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>")

	for i := 0; i < numPages; i++ {
		pageID := 4 + i
		contentID := firstContentID + i
		writeObj(pageID, fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.0f %.0f] /Resources << /Font << /F1 %d 0 R >> >> /Contents %d 0 R >>",
			pageWidth, pageHeight, fontObjID, contentID,
		))
	}

	for i, pageLines := range pages {
		contentID := firstContentID + i
		stream := contentStreamForPage(pageLines)
		body := fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(stream), stream)
		writeObj(contentID, body)
	}

	totalObjs := 3 + numPages + numPages
	xrefStart := buf.Len()

	buf.WriteString(fmt.Sprintf("xref\n0 %d\n", totalObjs+1))
	buf.WriteString("0000000000 65535 f \n")
	for id := 1; id <= totalObjs; id++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[id]))
	}

	buf.WriteString("trailer\n")
	buf.WriteString(fmt.Sprintf("<< /Size %d /Root 1 0 R >>\n", totalObjs+1))
	buf.WriteString("startxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	buf.WriteString("%%EOF")

	return buf.Bytes()
}
