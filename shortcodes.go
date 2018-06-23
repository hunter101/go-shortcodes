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

func (s *Shortcodes) Parse(text string) string {
	var names []string
	for name, _ := range s.registered {
		names = append(names, name)
	}
	namesString := strings.Join(names, "|")
	regex := regexp.MustCompile(fmt.Sprintf(
		`\[(%s)(\s+[^\]]+)?\]` + // Initial name and arguments
			  `([^\[]*)` + // Content
			  `\[/(%s)\]`, // Closing Tag
				namesString, namesString))
	for _, match := range regex.FindAllStringSubmatch(text, -1) {
		name, content := match[1], match[3]
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