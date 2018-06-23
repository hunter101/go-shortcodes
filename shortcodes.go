package shortcodes

import (
	"regexp"
	"fmt"
	"strings"
)

type callbackFunc func(args Args) string
type Args map[string]string
type Shortcodes struct {
	registered map[string]callbackFunc
}

func New() Shortcodes {
	return Shortcodes{
		registered: map[string]callbackFunc{},
	}
}

func (s *Shortcodes) Register(name string, callback callbackFunc) error {
	regex := "^[a-z0-9_]+$"
	validName := regexp.MustCompile(regex)
	if !validName.MatchString(name) {
		return fmt.Errorf("shortcode names must match the regex: %s", regex)
	}
	_, exists := s.registered[name];
	if exists {
		return fmt.Errorf("shortcode %s already exists", name)
	}
	s.registered[name] = callback
	return nil
}

func (s *Shortcodes) Parse2(text string) string {
	var names []string
	for name, _ := range s.registered {
		names = append(names, name)
	}
	namesString := strings.Join(names, "|")
	regex := regexp.MustCompile(fmt.Sprintf(
		`\[(%s)(\s+[^\]]+)?\]`, namesString))
	for {
		match := regex.FindStringSubmatchIndex(text)
		if match == nil {
			break
		}
		args := Args{}
		openingTagStart, openingTagClose, tagNameStart, tagNameEnd := match[0], match[1], match[2], match[3]
		fullMatch := text[openingTagStart:openingTagClose]
		tagName := text[tagNameStart:tagNameEnd]

		if _, ok := s.registered[tagName]; !ok {
			continue
		}

		// Parse the arguments
		if match[4] != -1 {
			argsString := text[match[4]:match[5]]
			argsRegex := regexp.MustCompile(`\s*([^=]+)="([^"]+)"`)
			for _, argMatch := range argsRegex.FindAllStringSubmatch(argsString, -1) {
				args[argMatch[1]] = argMatch[2]
			}
		}

		var closingTagEnd = openingTagClose
		var textToReplace = text[openingTagStart:openingTagClose]

		// Is the tag self closing?  If not, get the content.
		if fullMatch[len(fullMatch)-2:len(fullMatch)-1] != "/" {
			closingTagRegex := regexp.MustCompile(fmt.Sprintf(`\[\/(%s)\]`, tagName))
			loc := text[openingTagClose:]
			closingMatch := closingTagRegex.FindStringIndex(loc)
			if closingMatch == nil {
				text = strings.Replace(text, textToReplace, "", 1)
				continue
			}
			closingTagEnd = openingTagClose + closingMatch[1]
			contentStart := openingTagClose
			contentEnd := openingTagClose+closingMatch[0]
			textToReplace = text[openingTagStart:closingTagEnd]
			content := text[contentStart:contentEnd]
			content = s.Parse2(content)
			args["content"] = content
		}

		replaced := s.registered[tagName](args)
		text = strings.Replace(text, textToReplace , replaced, 1)
	}

	return text
}

func (s *Shortcodes) Parse(text string) string {
	var names []string
	for name, _ := range s.registered {
		names = append(names, name)
	}
	namesString := strings.Join(names, "|")
	regex := regexp.MustCompile(fmt.Sprintf(
		`\[(%s)(\s+[^\]]+)?\]`+ // Initial name and arguments
			`([^\[]*)`+ // Content
		//`(.*)` + // Content
			`\[/(%s)\]`, // Closing Tag
		namesString, namesString))
	for _, match := range regex.FindAllStringSubmatch(text, -1) {
		name, content := match[1], match[3]

		fmt.Printf("%#v\n", match)
		// recursivly parse any nested shortcodes in the content
		content = s.Parse(content)
		args := Args{"content": content}

		// get the shortcode args
		if match[2] != "" {
			argsRegex := regexp.MustCompile(`\s*([^=]+)="([^"]+)"`)
			for _, argMatch := range argsRegex.FindAllStringSubmatch(match[2], -1) {
				args[argMatch[1]] = argMatch[2]
			}
		}

		if _, ok := s.registered[name]; !ok {
			continue
		}
		replaced := s.registered[name](args)
		text = strings.Replace(text, match[0], replaced, 1)
	}

	return text
}
