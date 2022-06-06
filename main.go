package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	//	"os"
	//	"io/ioutil"
)

type AutoGenerated struct { //output for validatros status in struct
	Validators []struct {
		//			OperatorAddress string `json:"operator_address"`
		Status      string `json:"status"`
		Description struct {
			Moniker string `json:"moniker"`
		} `json:"description"`
	} `json:"validators"`
}

func main() {
	bot, err := tgbotapi.NewBotAPI("BOT_TOKEN")
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			
			address := update.Message.Text
			prefix := strings.HasPrefix(address, "stars") //bool, if word starts with `stars`,  below code should run
			if prefix == true {

				log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text) //got log message in cli

				var (
					out []byte
					err error
				)
				//struct for balances
				type AutoGenerated struct {
					Balances []struct {
						Amount string `json:"amount"`
					} `json:"balances"`
				}
				//got balance in CLI
				out, err = exec.Command("starsd", "q", "bank", "balances", address, "--node", "https://stargaze.c29r3.xyz:443/rpc", "--output", "json").Output()
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please, input correct address")) //if address was with mistake
					continue
				}

				lookFor := "ustars"
				contain := strings.Contains(string(out), lookFor) //find key word(ustar) in output from out (starsd)
				if contain == true {                              //if it successful
					data := AutoGenerated{}
					jsonErr := json.Unmarshal(out, &data)
					if jsonErr != nil {
						log.Fatal(jsonErr)
					}

					amount, _ := strconv.ParseFloat(data.Balances[0].Amount, 64) //convert to float
					amountMath := amount / 1000000
					str := fmt.Sprintf("%.2f", amountMath)
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Balance is "+str+" stars"))
					continue

				} else if contain == false { //if it didn't successful
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Balance is 0 stars"))
					continue
				}
				
			}
			
			msgArr := strings.Split(update.Message.Text, " ")
			log.Printf("[%s] %s ", update.Message.Chat.UserName, update.Message.Text)
			switch msgArr[0] {
			case "/balance":
				
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please, input your address, it should starts with `stars1...`"))
				
			case "/status":

				var (
					out []byte
					err error
				)
				//recived status of all validatords with command
				out, err = exec.Command("starsd", "q", "staking", "validators", "--node", "https://stargaze.c29r3.xyz:443/rpc", "--output", "json").Output()

				if err != nil {
					log.Fatal(err)
				}
				data := AutoGenerated{}
				jsonErr := json.Unmarshal(out, &data) 
				if jsonErr != nil {
					log.Fatal(jsonErr)
				}
				//range all data in struct from command above
				for _, s := range data.Validators {
					str := fmt.Sprintf("%s", s)
					//find and replace words
					var re = regexp.MustCompile(`{BOND_STATUS_BONDED {`)
					var re1 = regexp.MustCompile(`{BOND_STATUS_UNBONDING {`)
					var re2 = regexp.MustCompile(`{BOND_STATUS_UNBONDED {`)
					var re3 = regexp.MustCompile(`}}`)

					strRe := re.ReplaceAllString(str, "Status: Active ✅ Moniker: ")
					strRe1 := re1.ReplaceAllString(strRe, "Status: Jailed ⚠️ Moniker: ")
					strRe2 := re2.ReplaceAllString(strRe1, "Status: Inactive ❌ Moniker: ")
					strRe3 := re3.ReplaceAllString(strRe2, "")
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, strRe3)) //print all of validators
				}
			default:
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, now I know only /balance and /status commands."))
			}

		}
	}
}
