package commands

import (
	"context"
	"errors"
	"reflect"

	"github.com/bwmarrin/discordgo"
)

// Event is an constaint interface for all Discord events.
// This should only be used for type constraints.
type Event interface {
	*discordgo.MessageCreate | *discordgo.MessageUpdate | *discordgo.MessageDelete |
		*discordgo.MessageReactionAdd | *discordgo.MessageReactionRemove | *discordgo.MessageReactionRemoveAll |
		*discordgo.ChannelCreate | *discordgo.ChannelUpdate | *discordgo.ChannelDelete |
		*discordgo.ChannelPinsUpdate | *discordgo.GuildCreate | *discordgo.GuildUpdate | *discordgo.GuildDelete |
		*discordgo.GuildBanAdd | *discordgo.GuildBanRemove | *discordgo.GuildEmojisUpdate |
		*discordgo.GuildIntegrationsUpdate | *discordgo.GuildMemberAdd | *discordgo.GuildMemberRemove |
		*discordgo.GuildMemberUpdate | *discordgo.GuildMembersChunk | *discordgo.GuildRoleCreate |
		*discordgo.GuildRoleUpdate | *discordgo.GuildRoleDelete | *discordgo.PresenceUpdate | *discordgo.TypingStart |
		*discordgo.UserUpdate | *discordgo.VoiceStateUpdate | *discordgo.VoiceServerUpdate |
		*discordgo.WebhooksUpdate | *discordgo.InteractionCreate
}

// Handler is a function that handles a Discord event.
type Handler[T Event] func(s *discordgo.Session, e T) error

// Command is an interface for a command.
type Command[T Event] interface {
	// Name returns the name of the command.
	Name() string
	// Execute is the handler for the command that is called when the event is triggered.
	Execute(s *discordgo.Session, e T)
}

// Base is a common base for all commands.
// It should not be executed but rather embedded in other commands.
// The embedder may only implement command specific methods.
//
// Example:
//
//	type MyCommand struct {
//		*Base[*discordgo.InteractionCreate]
//	}
//
//	func NewMyCommand() *MyCommand {
//		return &MyCommand{Base: NewBase[*discordgo.InteractionCreate]("mycommand")}
//	}
//
//	func (c *MyCommand) Execute(s *discordgo.Session, e *discordgo.InteractionCreate) error {
//		// do something
//	}
type Base[T Event] struct {
	// name is the name of the command.
	name string
}

// Name returns the name of the command.
func (c *Base[T]) Name() string {
	return c.name
}

// Execute is the handler for the command that is called when the event is triggered.
func (c *Base[T]) Execute(_ *discordgo.Session, _ T) {
	panic("base command should not be executed")
}

// NewBase creates the common base for all commands.
// The name is the name of the command.
// The name should be unique and should not contain spaces.
func NewBase[T Event](name string) *Base[T] {
	if name == "" {
		panic("name cannot be empty")
	}

	return &Base[T]{name: name}
}

func (c *Base[T]) ReplyToInteraction(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.InteractionResponseData) error {
	var basetype T
	if reflect.TypeOf(basetype) != reflect.TypeOf((*discordgo.Interaction)(nil)) {
		return errors.New("invalid event type")
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	}, discordgo.WithContext(ctx))
}
