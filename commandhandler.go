package main

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Command interface {
	Message(*Context)
	Description() string
	Usage() string
	Detailed() string
}

type CommandHandler struct {
	Commands map[string]Command
}

func (ch *CommandHandler) AddCommand(n string, c Command) {
	ch.Commands[n] = c
}

func (ch *CommandHandler) HandleCommands(ctx *Context) {
	if ctx.Invoked == "help" {
		go ch.HelpFunction(ctx)
	} else {
		called, ok := ch.Commands[ctx.Invoked]
		if ok {
			go called.Message(ctx)
		} else {
			logerror(errors.New(`Command "` + ctx.Invoked + `" not found`))
		}
	}
}

func (ch *CommandHandler) HelpFunction(ctx *Context) {
	ctx.Sess.ChannelMessageDelete(ctx.Mess.ChannelID, ctx.Mess.ID)
	color := ctx.Sess.State.UserColor(ctx.Mess.Author.ID, ctx.Mess.ChannelID)
	var desc string
	if len(ctx.Args) != 0 {
		called, ok := ch.Commands[ctx.Args[0]]
		if ok {
			desc = fmt.Sprintf("`%s%s %s`\n%s", conf.Prefix, ctx.Args[0], called.Usage(), called.Detailed())
		} else {
			desc = "No command called " + ctx.Args[0] + " found!"
		}
	} else {
		desc = "Commands:"
		desc += fmt.Sprintf(" `%shelp [command]` for more info!", conf.Prefix)
		for k, v := range ch.Commands {
			desc += fmt.Sprintf("\n`%s%s` - %s", conf.Prefix, k, v.Description())
		}
	}
	embed := &discordgo.MessageEmbed{Author: &discordgo.MessageEmbedAuthor{Name: ctx.Mess.Author.Username, IconURL: fmt.Sprintf("https://discordapp.com/api/users/%s/avatars/%s.jpg", ctx.Mess.Author.ID, ctx.Mess.Author.Avatar)}, Description: desc, Color: color}
	embed.Description += "\n\n" + ctx.Mess.Author.Username + " is using a version of [FloSelfbot!](https://github.com/Moonlington/FloSelfbot)"
	ctx.Sess.ChannelMessageSendEmbed(ctx.Channel.ID, embed)
}
