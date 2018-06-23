package shortcodes

import(
	"testing"
	"fmt"
)

var testingFunctions = map[string]callbackFunc{
	"make_bold": func(args Args) string {
		return fmt.Sprintf("<strong>%s</strong>", args["content"])
	},
	"delete_me": func(args Args) string {
		return ""
	},
	"inject_me":func (args Args) string {
		injector, ok := args["injector"]
		if !ok {
			injector = "missing"
		}
		return fmt.Sprintf("%s%s", args["content"], injector)
	},
}

func TestCanRegisterShortcodeFunctions(t *testing.T) {
	shortcodes := New()
	for name, function := range testingFunctions {
		err := shortcodes.Register(name, function)
		if err != nil {
			t.Errorf("registering %s func returned error: %s", name, err)
		}
	}
	if len(testingFunctions) != len(shortcodes.registered) {
		t.Errorf("number of registered functions %d, doesn't match number of function added %d", len(shortcodes.registered), len(testingFunctions))
	}
}

func TestCannotAddMultipleFunctionsWithTheSameName(t *testing.T) {
	shortcodes := New()
	err := shortcodes.Register("make_bold", testingFunctions["make_bold"])
	if err != nil {
		t.Errorf("Error registering function make_bold: %s", err)
	}
	err = shortcodes.Register("make_bold", testingFunctions["make_bold"])
	if err == nil {
		t.Errorf("No error returned when function with same name registered twice")
	}
}

func TestCannotAddShortcodesWithIncorrectNames(t *testing.T) {
	shortcodes := New()
	basicFunc := func(args Args) string {return ""}
	tests := map[string]callbackFunc{
		"with space": basicFunc,
		"Captialised": basicFunc,
		"hy-phens": basicFunc,
		"$#%%%": basicFunc,
		"": basicFunc,
	}

	for name, callback := range tests {
		if err := shortcodes.Register(name, callback); err == nil {
			t.Errorf("Incorrect name %s passed into register and no error was returned", name)
		}
	}
}

func TestSimpleShortcodeRegex(t *testing.T) {
	sc := New()
	text := `This is some text. [make_bold]This is some bold text[/make_bold], here is some [delete_me]badness[/delete_me]`
	sc.Register("make_bold", testingFunctions["make_bold"])
	sc.Register("delete_me", testingFunctions["delete_me"])
	sc.Register("function_name_that_doesnt_exist", testingFunctions["delete_me"])
	formatted := sc.Parse(text)
	expected := "This is some text. <strong>This is some bold text</strong>, here is some "
	if formatted != expected {
		t.Errorf("make_bold function applied to text \n- %s\nwanted to get \n- %s\ninstead got\n- %s\n", text, expected, formatted)
	}
}

func TestRegexWithArgs(t *testing.T) {
	sc := New()
	text := `calculate this sum [inject_me injector="5"][/inject_me] + [inject_me injector="5"][/inject_me] = [inject_me injector="10"][/inject_me]`
	sc.Register("inject_me", testingFunctions["inject_me"])
	formatted := sc.Parse(text)
	expected := `calculate this sum 5 + 5 = 10`
	if formatted != expected {
		t.Errorf("inject_me function applied to text \n- %s\n wanted to get\n- %s\nInstead got\n- %s\n", text, expected, formatted)
	}
}
