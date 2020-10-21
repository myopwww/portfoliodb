package main

//TODO: deal with abbreviations & footnotes
//TODO: turn paragraphs into HTML

import (
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	// "github.com/davecgh/go-spew/spew"
	// "github.com/gomarkdown/markdown"
	// "github.com/gomarkdown/markdown/parser"
	// "github.com/gomarkdown/markdown/renderer"
	// "github.com/davecgh/go-spew/spew"
	"github.com/metal3d/go-slugify"
)

const (
	patternImageOrMediaOrLinkDeclaration string = `^([!>]?)\[([^"\]]+)(?: "([^"\]]+)")?\]\(([^\)]+)\)$`
	patternLanguageMarker                string = `^::\s+(.+)$`
	patternFootnoteDeclaration           string = `^\[(\d+)\]:\s+(.+)$`
	patternAbbreviationDefinition        string = `^\*\[([^\]]+)\]:\s+(.+)$`
	patternParagraphID                   string = `^\(([a-z-]+)\)$`
	patternTitle                         string = `^#\s+(.+)$`
)

// ParseYAMLHeader parses the YAML header of a description markdown file and returns
// the rest of the content (all except the YAML header)
func ParseYAMLHeader(descriptionRaw string) (map[string]interface{}, string) {
	var inYAMLHeader bool
	var rawYAMLPart string
	var markdownPart string
	for _, line := range strings.Split(descriptionRaw, "\n") {
		// if strings.TrimSpace(line) == "" && !inYAMLHeader {
		// 	continue
		// }
		if strings.TrimSpace(line) == "---" {
			inYAMLHeader = !inYAMLHeader
			continue
		}
		if inYAMLHeader {
			rawYAMLPart += line + "\n"
		} else {
			markdownPart += line + "\n"
		}
	}
	var parsedYAMLPart map[string]interface{}
	yaml.Unmarshal([]byte(rawYAMLPart), &parsedYAMLPart)
	return parsedYAMLPart, markdownPart
}

// Abbreviation represents an abbreviation declaration in a description.md file
type Abbreviation struct {
	Name       string
	Definition string
}

// Footnote represents a footnote declaration in a description.md file
type Footnote struct {
	Number  uint16 // Oh no, what a bummer, you can't have more than 65 535 footnotes
	Content string
}

type Paragraph struct {
	ID      string
	Content string
}

type Link struct {
	ID    string
	Name  string
	Title string
	URL   string
}

type WorkObject struct {
	Metadata   map[string]interface{}
	Title      map[string]string
	Paragraphs map[string][]Paragraph
	Media      map[string][]Media
	Links      map[string][]Link
	Footnotes  map[string][]Footnote
}

type ParsedDescription struct {
	Metadata               map[string]interface{}
	Title                  map[string]string
	Paragraphs             map[string][]Paragraph
	MediaEmbedDeclarations map[string][]MediaEmbedDeclaration
	ImageEmbedDeclarations map[string][]ImageEmbedDeclaration
	Links                  map[string][]Link
	Footnotes              map[string][]Footnote
}

// Chunk binds some content to its chunk type
// Legal chunk types:
// - abbreviation
// - paragraphWithID
// - image
// - media
// - link
// - footnoteDeclaration
// - paragraph
// - title
type Chunk struct {
	Type    string
	Content string
}

// MediaEmbedDeclaration represents >[media](...) embeds.
// Only stores the info extracted from the syntax, no filesystem interactions.
type MediaEmbedDeclaration struct {
	Alt    string
	Title  string
	Source string
}

// ImageEmbedDeclaration represents ![media](...) embeds.
// Only stores the info extracted from the syntax, no filesystem interactions.
type ImageEmbedDeclaration = MediaEmbedDeclaration

// CollectAbbreviation tries to match the given line and collect an abbreviation.
// Return values:
// 1. Abbreviation struct
// 2. Whether the line defines an abbreviation (bool)
func CollectAbbreviation(line string) (Abbreviation, bool) {
	pattern := regexp.MustCompile(patternAbbreviationDefinition)
	if pattern.MatchString(line) {
		matches := pattern.FindStringSubmatch(line)
		return Abbreviation{Name: matches[0], Definition: matches[1]}, true
	}
	return Abbreviation{}, false
}

// ParseFootnote parses raw markdown into a footnote struct.
func ParseFootnote(markdownRaw string) Footnote {
	groups := RegexpGroups(patternFootnoteDeclaration, markdownRaw)
	footnoteNumber, _ := strconv.ParseInt(groups[0], 10, 16)
	return Footnote{Number: uint16(footnoteNumber), Content: groups[1]}
}

// CollectAbbreviationsAndFootnotes iterates through the document's lines and
// extracts abbreviations and footnotes declarations from the file
// The first returned value is the markdown document with parsed declarations removed.
func CollectAbbreviationsAndFootnotes(markdownRaw string) (string, []Abbreviation, []Footnote) {
	lines := strings.Split(markdownRaw, "\n")
	markdownRet := ""
	abbreviations := make([]Abbreviation, 8^16)
	footnotes := make([]Footnote, 8^16)
	for _, line := range lines {
		abbreviation, definesAbbreviation := CollectAbbreviation(line)
		footnote := ParseFootnote(line)
		if definesAbbreviation {
			abbreviations = append(abbreviations, abbreviation)
		} else if footnote.Content != "" {
			footnotes = append(footnotes, footnote)
		} else {
			markdownRet += line + "\n"
			continue
		}

	}
	return markdownRet, abbreviations, footnotes
}

// SplitOnLanguageMarkers returns two values:
// 1. the text before any language markers
// 2. a map with language codes as keys and the content as values
func SplitOnLanguageMarkers(markdownRaw string) (string, map[string]string) {
	lines := strings.Split(markdownRaw, "\n")
	pattern := regexp.MustCompile(patternLanguageMarker)
	currentLanguage := ""
	before := ""
	markdownRawPerLanguage := map[string]string{}
	for _, line := range lines {
		if pattern.MatchString(line) {
			currentLanguage = pattern.FindStringSubmatch(line)[1]
			markdownRawPerLanguage[currentLanguage] = ""
		}
		if currentLanguage == "" {
			before += line + "\n"
		} else {
			markdownRawPerLanguage[currentLanguage] += line + "\n"
		}
	}
	return before, markdownRawPerLanguage
}

// ExtractTitle extracts the first <h1> from markdown
func ExtractTitle(line string) string {
	pattern := regexp.MustCompile(`^#\s+(.+)$`)
	if pattern.MatchString(line) {
		return pattern.FindStringSubmatch(line)[0]
	}
	return ""
}

// FindTitle searches through markdownRaw line-by-line until ExtractTitle finds title, or until it reaches the end.
func FindTitle(markdownRaw string) string {
	lines := strings.Split(markdownRaw, "\n")
	foundTitle := ""
	for _, line := range lines {
		foundTitle = line
	}
	return foundTitle
}

// ExtractMedia extracts media declarations (>[alt "title"](source)), images (![alt "title"](source)) or links ([alt "title"](source))
// Return value is a regex match string array: first character (empty for links), alt, title, source.
func extractMediaOrImageOrLink(line string) []string {
	pattern := regexp.MustCompile(patternImageOrMediaOrLinkDeclaration)
	if pattern.MatchString(line) {
		matches := pattern.FindStringSubmatch(line)
		return matches
	}
	return make([]string, 0)
}

func extractLink(regexMatches []string) Link {
	return Link{
		ID:    slugify.Marshal(regexMatches[2]),
		Name:  regexMatches[2],
		Title: regexMatches[3],
		URL:   regexMatches[4],
	}
}

func extractImage(regexMatches []string) ImageEmbedDeclaration {
	return ImageEmbedDeclaration{
		Alt:    regexMatches[2],
		Title:  regexMatches[3],
		Source: regexMatches[4],
	}
}

func extractMedia(regexMatches []string) MediaEmbedDeclaration {
	return MediaEmbedDeclaration{
		Alt:    regexMatches[2],
		Title:  regexMatches[3],
		Source: regexMatches[4],
	}
}

// ParseParagraph takes a chunk of type "paragraph" or "paragraphWithID" and returns a parsed Paragraph with HTML content
func ParseParagraph(chunk Chunk) Paragraph {
	var paragraphID string = ""
	var paragraphContent string = chunk.Content
	if chunk.Type == "paragraphWithID" {
		paragraphID = RegexpGroups(patternParagraphID, strings.Split(chunk.Content, "\n")[0])[1]
		// Every line except the first (the paragraph id marker)
		paragraphContent = strings.Join(strings.Split(chunk.Content, "\n")[1:], "\n")
	}
	return Paragraph{
		Content: paragraphContent,
		ID:      paragraphID,
	}
}

// ParseLanguagedChunks takes in raw markdown without language markers (called on SplitOnLanguageMarker's output)
// and dispatches parsing to the appropriate functions, dependending on each chunk's type (a paragraph, an image, etc.)
func ParseLanguagedChunks(markdownRaw string) []Chunk {
	chunks := strings.Split(markdownRaw, "\n\n")
	typedChunks := make([]Chunk, 0)

	for _, chunk := range chunks {
		// Skip empty chunks
		chunk = strings.TrimSpace(chunk)
		if len(chunk) == 0 {
			continue
		} else if RegexpMatches(patternAbbreviationDefinition, chunk) {
			typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "abbreviation"})
		} else if RegexpMatches(patternParagraphID, strings.Split(chunk, "\n")[0]) {
			typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "paragraphWithID"})
		} else if RegexpMatches(patternImageOrMediaOrLinkDeclaration, chunk) {
			mediaOrImageOrLinkMarker := RegexpGroups(patternImageOrMediaOrLinkDeclaration, chunk)[1]
			if mediaOrImageOrLinkMarker == "" {
				typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "link"})
			} else if mediaOrImageOrLinkMarker == ">" {
				typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "media"})
			} else if mediaOrImageOrLinkMarker == "!" {
				typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "image"})
			}
		} else if RegexpMatches(patternFootnoteDeclaration, chunk) {
			typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "footnoteDeclaration"})
		} else if RegexpMatches(patternLanguageMarker, chunk) {
			continue
		} else if RegexpMatches(patternTitle, chunk) {
			typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "title"})
		} else {
			typedChunks = append(typedChunks, Chunk{Content: chunk, Type: "paragraph"})
		}
	}

	return typedChunks
}

func ParseImageChunk(chunk Chunk) ImageEmbedDeclaration {
	return extractImage(RegexpGroups(patternImageOrMediaOrLinkDeclaration, chunk.Content))
}
func ParseMediaChunk(chunk Chunk) MediaEmbedDeclaration {
	return extractMedia(RegexpGroups(patternImageOrMediaOrLinkDeclaration, chunk.Content))
}
func ParseLinkChunk(chunk Chunk) Link {
	return extractLink(RegexpGroups(patternImageOrMediaOrLinkDeclaration, chunk.Content))
}

// GetAllLanguages returns all language codes used in the document
func GetAllLanguages(markdownRaw string) []string {
	lines := strings.Split(markdownRaw, "\n")
	languages := make([]string, 0)
	for _, line := range lines {
		if RegexpMatches(patternLanguageMarker, line) {
			languages = append(languages, RegexpGroups(patternLanguageMarker, line)[1])
		}
	}
	return languages
}

func ParseDescription(markdownRaw string) ParsedDescription {
	metadata, markdownRaw := ParseYAMLHeader(markdownRaw)
	// notLocalizedRaw: raw markdown before the first language marker
	notLocalizedRaw, localizedRawBlocks := SplitOnLanguageMarkers(markdownRaw)
	paragraphs := make(map[string][]Paragraph, 0)
	mediaEmbedDeclarations := make(map[string][]MediaEmbedDeclaration, 0)
	imageEmbedDeclarations := make(map[string][]ImageEmbedDeclaration, 0)
	links := make(map[string][]Link, 0)
	title := make(map[string]string, 0)
	footnotes := make(map[string][]Footnote, 0)
	for _, language := range GetAllLanguages(markdownRaw) {
		// unlocalized stuff appears the same in every language.
		chunks := ParseLanguagedChunks(notLocalizedRaw)
		chunks = append(chunks, ParseLanguagedChunks(localizedRawBlocks[language])...)
		currentLanguageParagraphs := make([]Paragraph, 0)
		currentLanguageMediaEmbedDeclarations := make([]MediaEmbedDeclaration, 0)
		currentLanguageImageEmbedDeclarations := make([]ImageEmbedDeclaration, 0)
		currentLanguageLinks := make([]Link, 0)
		currentLanguageFootnotes := make([]Footnote, 0)
		var currentLanguageTitle string
		for _, chunk := range chunks {
			if chunk.Type == "title" {
				currentLanguageTitle = RegexpGroups(patternTitle, chunk.Content)[1]
			} else if chunk.Type == "footnote" {
				footnote := ParseFootnote(chunk.Content)
				currentLanguageFootnotes = append(currentLanguageFootnotes, footnote)
			} else if chunk.Type == "paragraph" || chunk.Type == "paragraphWithID" {
				currentLanguageParagraphs = append(currentLanguageParagraphs, ParseParagraph(chunk))
			} else if chunk.Type == "media" {
				currentLanguageMediaEmbedDeclarations = append(currentLanguageMediaEmbedDeclarations, ParseMediaChunk(chunk))
			} else if chunk.Type == "image" {
				currentLanguageImageEmbedDeclarations = append(currentLanguageImageEmbedDeclarations, ParseImageChunk(chunk))
			} else if chunk.Type == "links" {
				currentLanguageLinks = append(currentLanguageLinks, ParseLinkChunk(chunk))
			}
		}
		paragraphs[language] = currentLanguageParagraphs
		links[language] = currentLanguageLinks
		title[language] = currentLanguageTitle
		mediaEmbedDeclarations[language] = currentLanguageMediaEmbedDeclarations
		imageEmbedDeclarations[language] = currentLanguageImageEmbedDeclarations
		footnotes[language] = currentLanguageFootnotes
	}
	return ParsedDescription{
		Metadata:               metadata,
		Paragraphs:             paragraphs,
		Links:                  links,
		Title:                  title,
		MediaEmbedDeclarations: mediaEmbedDeclarations,
		ImageEmbedDeclarations: imageEmbedDeclarations,
		Footnotes:              footnotes,
	}
}
