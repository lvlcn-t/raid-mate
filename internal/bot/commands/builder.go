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

type InfoBuilder struct {
	c discord.SlashCommandCreate
}

func NewInfoBuilder() InfoBuilder {
	return InfoBuilder{c: discord.SlashCommandCreate{}}
}

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

func (i InfoBuilder) NSFW(nsfw bool) InfoBuilder { //nolint:gocritic // builder pattern
	i.c.NSFW = &nsfw
	return i
}

func (i InfoBuilder) Option(option discord.ApplicationCommandOption) InfoBuilder { //nolint:gocritic // builder pattern
	i.c.Options = append(i.c.Options, option)
	return i
}

func (i InfoBuilder) Build() discord.SlashCommandCreate { //nolint:gocritic // builder pattern
	return i.c
}

type StringOptionBuilder struct {
	o discord.ApplicationCommandOptionString
}

func NewStringOptionBuilder() StringOptionBuilder {
	return StringOptionBuilder{o: discord.ApplicationCommandOptionString{}}
}

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

func (so StringOptionBuilder) Required(required bool) StringOptionBuilder { //nolint:gocritic // builder pattern
	so.o.Required = required
	return so
}

func (so StringOptionBuilder) Choices(choices ...discord.ApplicationCommandOptionChoiceString) StringOptionBuilder { //nolint:gocritic // builder pattern
	so.o.Choices = choices
	return so
}

func (so StringOptionBuilder) Build() discord.ApplicationCommandOption { //nolint:gocritic // builder pattern
	return so.o
}

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
