package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
)

const (
	// maxNameLength is the maximum length of a name that discord allows.
	maxNameLength int = 32
	// maxDescriptionLength is the maximum length of a description that discord allows.
	maxDescriptionLength int = 100
)

// InfoBuilder is a builder for a slash command info. It is used to create a slash command.
type InfoBuilder struct {
	c discord.SlashCommandCreate
}

// NewInfoBuilder creates a new info builder.
func NewInfoBuilder() InfoBuilder {
	return InfoBuilder{c: discord.SlashCommandCreate{}}
}

// Name sets the name of the slash command and its localizations.
// The name should be unique and should not contain spaces nor be longer than 32 characters.
//
// Provide nil for localizations if the name should not be localized.
func (i InfoBuilder) Name(name string, localizations map[discord.Locale]string) InfoBuilder { //nolint:gocritic // builder pattern
	if len(name) > maxNameLength {
		panic(fmt.Sprintf("name is too long: %d > 32", len(name)))
	}

	for locale, n := range localizations {
		if len(n) > maxNameLength {
			panic(fmt.Sprintf("name for locale %q is too long: %d > 32", locale, len(n)))
		}
	}

	i.c.Name = name
	i.c.NameLocalizations = localizations
	return i
}

// Description sets the description of the slash command and its localizations.
// The description should not be longer than 100 characters.
//
// Provide nil for localizations if the description should not be localized.
func (i InfoBuilder) Description(description string, localizations map[discord.Locale]string) InfoBuilder { //nolint:gocritic // builder pattern
	if len(description) > maxDescriptionLength {
		panic(fmt.Sprintf("description is too long: %d > 100", len(description)))
	}

	for locale, desc := range localizations {
		if len(desc) > maxDescriptionLength {
			panic(fmt.Sprintf("description for locale %q is too long: %d > 100", locale, len(desc)))
		}
	}

	i.c.Description = description
	i.c.DescriptionLocalizations = localizations
	return i
}

// NSFW sets whether the slash command is NSFW.
func (i InfoBuilder) NSFW(nsfw bool) InfoBuilder { //nolint:gocritic // builder pattern
	i.c.NSFW = &nsfw
	return i
}

// Option adds an option to the slash command.
func (i InfoBuilder) Option(option discord.ApplicationCommandOption) InfoBuilder { //nolint:gocritic // builder pattern
	i.c.Options = append(i.c.Options, option)
	return i
}

// Build builds the slash command.
func (i InfoBuilder) Build() discord.SlashCommandCreate { //nolint:gocritic // builder pattern
	return i.c
}

// StringOptionBuilder is a builder for a string option. It is used to create a string option.
type StringOptionBuilder struct {
	o discord.ApplicationCommandOptionString
}

// NewStringOptionBuilder creates a new string option builder.
func NewStringOptionBuilder() StringOptionBuilder {
	return StringOptionBuilder{o: discord.ApplicationCommandOptionString{}}
}

// Name sets the name of the string option and its localizations.
// The name should not be longer than 32 characters.
//
// Provide nil for localizations if the name should not be localized.
func (so StringOptionBuilder) Name(name string, localization map[discord.Locale]string) StringOptionBuilder { //nolint:gocritic // builder pattern
	if len(name) > maxNameLength {
		panic(fmt.Sprintf("name is too long: %d > 32", len(name)))
	}

	for locale, n := range localization {
		if len(n) > maxNameLength {
			panic(fmt.Sprintf("name for locale %q is too long: %d > 32", locale, len(n)))
		}
	}

	so.o.Name = name
	return so
}

// Description sets the description of the string option and its localizations.
// The description should not be longer than 100 characters.
//
// Provide nil for localizations if the description should not be localized.
func (so StringOptionBuilder) Description(description string, localization map[discord.Locale]string) StringOptionBuilder { //nolint:gocritic // builder pattern
	if len(description) > maxDescriptionLength {
		panic(fmt.Sprintf("description is too long: %d > 100", len(description)))
	}

	for locale, desc := range localization {
		if len(desc) > maxDescriptionLength {
			panic(fmt.Sprintf("description for locale %q is too long: %d > 100", locale, len(desc)))
		}
	}

	so.o.Description = description
	so.o.DescriptionLocalizations = localization
	return so
}

// Required sets whether the string option is required.
func (so StringOptionBuilder) Required(required bool) StringOptionBuilder { //nolint:gocritic // builder pattern
	so.o.Required = required
	return so
}

// Choices sets the choices of the string option.
func (so StringOptionBuilder) Choices(choices ...discord.ApplicationCommandOptionChoiceString) StringOptionBuilder { //nolint:gocritic // builder pattern
	so.o.Choices = choices
	return so
}

// Build builds the string option.
func (so StringOptionBuilder) Build() discord.ApplicationCommandOption { //nolint:gocritic // builder pattern
	return so.o
}

// NewStringOptionChoice creates a new string option choice.
// The name and value should not be longer than 32 characters.
//
// Provide nil for localizedNames if the name should not be localized.
func NewStringOptionChoice(name, value string, localizedNames map[discord.Locale]string) discord.ApplicationCommandOptionChoiceString {
	if len(name) > maxNameLength {
		panic(fmt.Sprintf("name is too long: %d > 32", len(name)))
	}

	if len(value) > maxNameLength {
		panic(fmt.Sprintf("value is too long: %d > 32", len(value)))
	}

	for locale, n := range localizedNames {
		if len(n) > maxNameLength {
			panic(fmt.Sprintf("name for locale %q is too long: %d > 32", locale, len(n)))
		}
	}

	return discord.ApplicationCommandOptionChoiceString{
		Name:              name,
		NameLocalizations: localizedNames,
		Value:             value,
	}
}
