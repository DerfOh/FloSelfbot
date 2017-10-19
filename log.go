package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	bufferMin int = 4096 * 2
	bufferMax int = 65536
)

var logbuffers = make(map[string]map[string]*bytes.Buffer)
var logmintime time.Time
var logmaxtime time.Time

var mutex = &sync.Mutex{}

func sendToBuffer(s *discordgo.Session, ChannelID, str string) {
	var channel *discordgo.Channel
	var gn string
	var cn string

	channel, _ = s.State.Channel(ChannelID)
	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		gn = "Direct Message"
		if channel.Name == "" {
			cn = "DM:" + channel.ID
		} else {
			cn = channel.Name
		}
	} else {
		gn = guild.ID
		cn = channel.Name
	}

	now := time.Now()

	mutex.Lock()
	defer mutex.Unlock()

	logbuffer, ok := logbuffers[gn][cn]

	if !ok {
		if _, kok := logbuffers[gn]; !kok {
			logbuffers[gn] = make(map[string]*bytes.Buffer)
		}
		buf := new(bytes.Buffer)
		logbuffers[gn][cn] = buf
		buf.Grow(bufferMin)
		logbuffer = buf
	}

	logbuffer.WriteString(str)

	if now.Before(logmintime) {
	} else {
		if logbuffer.Len() >= logbuffer.Cap()-2200 {
			f, err := getLogFile(s, gn, cn)
			logerror(err)
			if conf.LogModeCompression {
				// var s []byte
				// if fi, _ := f.Stat(); fi.Size() > 0 {
				// 	fz, err := gzip.NewReader(f)
				// 	logerror(err)
				// 	s, err = ioutil.ReadAll(fz)
				// 	logerror(err)
				// 	fz.Close()
				// }
				w := gzip.NewWriter(f)
				// w.Write(append(s, logbuffer.Bytes()...))
				logbuffer.WriteTo(w)
				w.Close()
			} else {
				logbuffer.WriteTo(f)
			}
			f.Close()
		}
	}
}

func bufferLoop(s *discordgo.Session) {
	if conf.LogModeMinBuffer < 2 {
		conf.LogModeMinBuffer = 5
		EditConfigfile(conf)
	}
	if conf.LogModeMaxBuffer < 10 {
		conf.LogModeMaxBuffer = 20
		EditConfigfile(conf)
	}
	logmintime = time.Now().Add(time.Duration(conf.LogModeMinBuffer) * time.Second)
	logmaxtime = time.Now().Add(time.Duration(conf.LogModeMaxBuffer) * time.Second)
	for {
		timer := time.NewTimer(time.Second * time.Duration(conf.LogModeMaxBuffer))
		<-timer.C
		mutex.Lock()
		for k, v := range logbuffers {
			for c, logbuffer := range v {
				if logbuffer.Len() == 0 {
					continue
				}
				f, err := getLogFile(s, k, c)
				logerror(err)
				if conf.LogModeCompression {
					// var s []byte
					// if fi, _ := f.Stat(); fi.Size() > 0 {
					// 	fz, err := gzip.NewReader(f)
					// 	logerror(err)
					// 	s, err = ioutil.ReadAll(fz)
					// 	logerror(err)
					// 	fz.Close()
					// }
					w := gzip.NewWriter(f)
					// w.Write(append(s, logbuffer.Bytes()...))
					logbuffer.WriteTo(w)
					w.Close()
				} else {
					logbuffer.WriteTo(f)
				}
				f.Close()
			}
		}
		mutex.Unlock()

		logmintime = time.Now().Add(time.Duration(conf.LogModeMinBuffer) * time.Second)
		logmaxtime = time.Now().Add(time.Duration(conf.LogModeMaxBuffer) * time.Second)
	}
}

func logMessage(s *discordgo.Session, timestamp time.Time, user *discordgo.User, mID, cID, code, message string) {

	timestampo := timestamp.Format("2006-01-02 15:04:05") + " UTC"
	if user != nil {
		var namestr string
		channel, _ := s.State.Channel(cID)
		member, err := s.State.Member(channel.GuildID, user.ID)
		if err != nil {
			namestr = user.Username + "#" + user.Discriminator
		} else {
			if member.Nick != "" {
				namestr = member.Nick + " " + "(" + user.Username + "#" + user.Discriminator + ")"
			} else {
				namestr = user.Username + "#" + user.Discriminator
			}
		}

		sendToBuffer(s, cID, strings.Replace(fmt.Sprintf("%s %s %s %s ## %s ## %s", mID, timestampo, user.ID, code, namestr, message), "\n", "\t", -1)+"\n")
	} else {
		sendToBuffer(s, cID, strings.Replace(fmt.Sprintf("%s %s %s ## ## %s", mID, timestampo, code, message), "\n", "\t", -1)+"\n")
	}
}

func logMessageNoAuthor(s *discordgo.Session, timestamp time.Time, uID, mID, cID, code, userfield, message string) {
	timestampo := timestamp.Format("2006-01-02 15:04:05") + " UTC"

	sendToBuffer(s, cID, strings.Replace(fmt.Sprintf("%s %s %s %s ## %s ## %s", mID, timestampo, uID, code, userfield, message), "\n", "\t", -1)+"\n")
}

func getLogFile(s *discordgo.Session, g, c string) (*os.File, error) {
	os.MkdirAll(filepath.Join("logs", g), 0750)
	if g != "Direct Message" {
		_, err := os.Stat(filepath.Join("logs", g, "_servername.txt"))
		if os.IsNotExist(err) {
			guild, _ := s.State.Guild(g)
			f, err := os.Create(filepath.Join("logs", g, "_servername.txt"))
			logerror(err)
			defer f.Close()

			f.WriteString(guild.Name)
		}
	}
	re := regexp.MustCompile(`[\\/:\?!\*"<>\|]`)
	c = re.ReplaceAllString(c, "")
	path := filepath.Join("logs", g, c+".txt")
	if conf.LogModeCompression {
		path += ".gz"
	}
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.Create(path)
	}
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
}
