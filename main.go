package main

import (
	"flag"
	"fmt"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/fix44/heartbeat"
	"github.com/quickfixgo/quickfix"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Pinger struct {
}

func (p Pinger) OnCreate(sessionID quickfix.SessionID) {
	log.Debug().Str("sessionID", fmt.Sprintf("%+v", sessionID)).Msg("OnCreate")

	go func() {
		for {
			hb := heartbeat.New()
			hb.Header.Set(field.NewSenderCompID(sessionID.SenderCompID))
			hb.Header.Set(field.NewTargetCompID(sessionID.TargetCompID))
			err := quickfix.Send(hb)
			if err != nil {
				log.Error().Err(err).Msg("while sending heartbeat")
			}
			time.Sleep(5 * time.Second) // or runtime.Gosched() or similar per @misterbee
		}
	}()
}

func (p Pinger) OnLogon(sessionID quickfix.SessionID) {
	log.Debug().Str("sessionID", fmt.Sprintf("%+v", sessionID)).Msg("OnLogon")
}

func (p Pinger) OnLogout(sessionID quickfix.SessionID) {
	log.Debug().Str("sessionID", fmt.Sprintf("%+v", sessionID)).Msg("OnLogout")
}

func (p Pinger) ToAdmin(message *quickfix.Message, sessionID quickfix.SessionID) {
	log.Debug().Str("sessionID", fmt.Sprintf("%+v", sessionID)).Str("msg", message.String()).Msg("ToAdmin")
}

func (p Pinger) ToApp(message *quickfix.Message, sessionID quickfix.SessionID) error {
	log.Debug().Str("sessionID", fmt.Sprintf("%+v", sessionID)).Str("msg", message.String()).Msg("ToApp")
	return nil
}

func (p Pinger) FromAdmin(message *quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	log.Debug().Str("sessionID", fmt.Sprintf("%+v", sessionID)).Str("msg", message.String()).Msg("FromAdmin")
	return nil
}

func (p Pinger) FromApp(message *quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	log.Debug().Str("sessionID", fmt.Sprintf("%+v", sessionID)).Str("msg", message.String()).Msg("FromApp")
	return nil
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("Pinger!")
	flag.Parse()
	fileName := flag.Arg(0)

	//FooApplication is your type that implements the Application interface
	var app Pinger

	cfg, err := os.Open(fileName)
	if err != nil {
		log.Error().Err(err).Str("configFile", fileName).Msg("while opening file")
		panic(err)
	}
	appSettings, err := quickfix.ParseSettings(cfg)
	if err != nil {
		log.Error().Err(err).Str("configFile", fileName).Msg("while parsing settings")
		panic(err)
	}
	storeFactory := quickfix.NewMemoryStoreFactory()
	logFactory := quickfix.NewScreenLogFactory()

	initiator, err := quickfix.NewInitiator(app, storeFactory, appSettings, logFactory)
	if err != nil {
		log.Error().Err(err).Str("configFile", fileName).Msg("while initiating")
		panic(err)
	}

	err = initiator.Start()
	if err != nil {
		log.Error().Err(err).Str("configFile", fileName).Msg("while starting")
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	initiator.Stop()

}
