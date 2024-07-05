package main

import (
	"errors"
	"image/color"
	"regexp"
	"strconv"
	"strings"
	"os"
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	BACKSPACE_UNICODE        string = "U+0008"
	BACKSPACE_SYMBOL_UNICODE string = "U+2408"
	ERASE_TO_LEFT_UNICODE    string = "U+232B"

	DELETE_UNICODE         string = "U+007F"
	DELETE_SYMBOL_UNICODE  string = "U+2421"
	ERASE_TO_RIGHT_UNICODE string = "U+2326"
)

var bad_unicode_list []string = []string{BACKSPACE_UNICODE, BACKSPACE_SYMBOL_UNICODE, ERASE_TO_LEFT_UNICODE, DELETE_UNICODE, DELETE_SYMBOL_UNICODE, ERASE_TO_RIGHT_UNICODE}

// Defining custom Fyne theme
type custom_theme struct {
	fyne.Theme
}

func new_custom_theme() fyne.Theme {
	return &custom_theme{Theme: theme.DefaultTheme()}
}

func (t *custom_theme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(name, theme.VariantDark)
}

func (t *custom_theme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return 24 // Defining the point size of all text
	}

	return t.Theme.Size(name)
}

func is_string_in_list_of_strings(target string, str_list []string) bool {
	for _, str := range str_list {
		if str == target {
			return true
		}
	}
	return false
}

func conv_unicode_string_to_point_code(unicode_character_string string) (string, error) {
	if unicode_character_string == "" {
		return "", errors.New("an empty string was passed to the point-code --> unicode conversion")
	} else if is_string_in_list_of_strings(unicode_character_string, bad_unicode_list) {
		return "", errors.New("a backspace character was passed to the point-code --> unicode conversion")
	}

	unicode_character_string = strings.TrimSpace(unicode_character_string)
	var unicode_string_as_runes []rune = []rune(unicode_character_string) // Each rune is max 32-bits
	if len(unicode_string_as_runes) > 1 {
		return "", errors.New("Only one rune should exist in the unicode output box")
	}
	var unicode_character_rune rune = unicode_string_as_runes[0] // Capturing only the first rune

	var unicode_point_code_int int = int(unicode_character_rune)
	var unicode_point_code_string string = strconv.Itoa(unicode_point_code_int)

	var unicode_string string = "U+" + unicode_point_code_string

	return unicode_string, nil
}

func conv_point_code_to_unicode(point_code_string string) (string, error) {
	if point_code_string == "" {
		return "", errors.New("an empty string was passed to the point-code --> unicode conversion")
	} else if point_code_string == BACKSPACE_UNICODE {
		return "", errors.New("a backspace character was passed to the point-code --> unicode conversion")
	}

	point_code_string = strings.TrimSpace(point_code_string)
	var point_code string = strings.Split(point_code_string, "+")[1]

	// Converting string value to representative integer
	point_code_int, conv_err := strconv.Atoi(point_code)
	if conv_err != nil {
		return "", errors.New("failed to conv string to int (32-bits)")
	}

	return string(rune(point_code_int)), nil
}

func create_string_slice_from_config_file(file_loc string) ([]string, error) {
	file_data, file_read_err := os.ReadFile(file_loc)
	if file_read_err != nil {
		return []string{}, errors.New("failed to read in the config file")
	}

	return strings.Split(string(file_data), "\n"), nil // Returns a []string with each string index being one line
}

func main() {
	// Default window sizing
	var main_window_width int = 500
	var main_window_height int = 500

	// Getting the directory location of the main executable
	exe_path, exe_err := os.Executable()
	if exe_err != nil {
		log.Fatalln("failed to gather the executables location on the target system")
	}
	var main_directory_loc string = filepath.Dir(exe_path)

	// Creating the application
	var app_obj fyne.App = app.New()
	app_obj.Settings().SetTheme(new_custom_theme())

	// Creating a window from the application
	var main_window fyne.Window = app_obj.NewWindow("Unicode Generator")
	main_window.Resize(fyne.NewSize(float32(main_window_width), float32(main_window_height))) // Changing the main_window default size
	main_window.SetFixedSize(false)                                                           // All the user to resize the main window

	// Capture the Esc key press event --> Quiting if so
	main_window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyEscape {
			app_obj.Quit()
		}
	})

	// Creating widgets within window
	var unicode_output_entry *widget.Entry = widget.NewEntry()
	unicode_output_entry.TextStyle.Monospace = true // Used for better alignment
	unicode_output_entry.TextStyle.Bold = true

	var unicode_code_point_entry *widget.Entry = widget.NewEntry()
	unicode_code_point_entry.TextStyle.Monospace = true // Used for better alignment
	unicode_code_point_entry.TextStyle.Bold = true

	var symbols_config_loc string = filepath.Join(main_directory_loc, "available_symbols.config")
	available_symbols, symbol_err := create_string_slice_from_config_file(symbols_config_loc)
	if symbol_err != nil {
		log.Fatalln("failed to locate the available symbols config file")
	}

	var unicode_list *widget.List = widget.NewList(
		func() int {
			return len(available_symbols)
		},
		func() fyne.CanvasObject {
			list_label := widget.NewLabel("template")
			return list_label
		},
		func(list_id widget.ListItemID, canvas_obj fyne.CanvasObject) {
			var label *widget.Label = canvas_obj.(*widget.Label)
			label.SetText(available_symbols[list_id])
		})

	// Defining what happens when changes occur in the GUI program
	unicode_list.OnSelected = func(id widget.ListItemID) {
		// Gathering the text from the relative label --> Processing the unicode value
		var selected_symbol string = available_symbols[id]
		var in_brackets_regexp *regexp.Regexp = regexp.MustCompile(`\((.*?)\)`)
		var regex_captures []string = in_brackets_regexp.FindStringSubmatch(selected_symbol)

		// Checking that a (*) was found using regex
		if len(regex_captures) > 1 {
			var first_in_brackets_find string = regex_captures[1]
			unicode_output_entry.SetText(first_in_brackets_find) // Setting the text to just the unicode character

			// Displaying the unicode point code in the entry below the unicode character
			point_code, code_err := conv_unicode_string_to_point_code(first_in_brackets_find)
			if code_err != nil {
				// Do nothing
			} else {
				unicode_code_point_entry.SetText(point_code)
			}
		}
	}

	// Acts when a new unicode char is put into the "Output" entry
	unicode_output_entry.OnChanged = func(new_unicode_output string) {
		new_point_code, conv_err := conv_unicode_string_to_point_code(new_unicode_output)

		// Continuing if NO ERROR occurs (otherwise nothing)
		if conv_err == nil {
			unicode_code_point_entry.SetText(new_point_code)
		}
	}

	// Acts when a new unicode point code is put into the "Point Code" entry
	unicode_code_point_entry.OnChanged = func(new_point_code string) {
		new_unicode_output, conv_err := conv_point_code_to_unicode(new_point_code)

		// Continues if NO ERROR occurs (otherwise nothing)
		if conv_err == nil {
			unicode_output_entry.SetText(new_unicode_output)
		}
	}

	// Pushing all widgets into their relavant containers
	var unicode_output_fill_container *fyne.Container = container.NewPadded(unicode_output_entry)
	var unicode_output_label *widget.Label = widget.NewLabel("Output")
	var unicode_output_label_centre *fyne.Container = container.NewCenter(unicode_output_label)

	var unicode_code_point_fill_container *fyne.Container = container.NewPadded(unicode_code_point_entry)
	var unicode_code_point_label *widget.Label = widget.NewLabel("Point Code")
	var unicode_code_point_label_centre *fyne.Container = container.NewCenter(unicode_code_point_label)

	var left_vert_box *fyne.Container = container.New(layout.NewVBoxLayout(),
		layout.NewSpacer(),
		unicode_output_label_centre,
		unicode_output_fill_container,
		layout.NewSpacer(),
		unicode_code_point_label_centre,
		unicode_code_point_fill_container,
		layout.NewSpacer(),
	)
	var horizontal_split *container.Split = container.NewHSplit(left_vert_box, unicode_list)

	// Displaying content on the target machine
	main_window.SetContent(horizontal_split)
	main_window.ShowAndRun()
}
