package main

import (
	"fmt"
	"log"
	"net/http"
	"io"
	"os"
	"strings"
	"strconv"
	"crypto/ecdsa"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ethereum/go-ethereum/crypto"
)

// TODO: Refactor main function to reduce its Cognitive Complexity from 16 to the 15 allowed
// TODO: split logic to multiple files

var generalMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Crypto prices"),
			tgbotapi.NewKeyboardButton("Create address"),
			tgbotapi.NewKeyboardButton("Help"),
	),
)

var cryptoPricesMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("BTC"),
			tgbotapi.NewKeyboardButton("ETH"),
			tgbotapi.NewKeyboardButton("ADA"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("BNB"),
		tgbotapi.NewKeyboardButton("SOL"),
		tgbotapi.NewKeyboardButton("<- Back"),
	),
)

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore non-Message updates
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		switch update.Message.Text {
		case "Crypto prices":
			msg.Text = "Such cryptos available"
			msg.ReplyMarkup = cryptoPricesMenu
		case "<- Back":
			msg.Text = "Main menu"
			msg.ReplyMarkup = generalMenu
		case "BTC":
			btcPrice, err := getPrice("BTC")

			if err != nil {
				msg.Text = fmt.Sprint(err)
			} else {
				msg.Text = fmt.Sprint("Last BTC price is ", btcPrice, " USD")
			}
		case "ETH":
			ethPrice, err := getPrice("ETH")

			if err != nil {
				msg.Text = fmt.Sprint(err)
			} else {
				msg.Text = fmt.Sprint("Last ETH price is ", ethPrice, " USD")
			}
		case "Create address":
			msg.Text = generateAddress()
		default:
			msg.Text = "Here is list of bot features"
			msg.ReplyMarkup = generalMenu
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
  }
}

func getPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("https://api-pub.bitfinex.com/v2/ticker/t%sUSD", symbol)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return 0.0, errors.New("Unexpected error happened while fetching price")
	}

	defer resp.Body.Close()
	dataInBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return 0.0, errors.New("Unexpected error happened while reading response body")
	}

	return convertBytesToInt(dataInBytes)[6].(float64), nil
}

func convertBytesToInt(byteArray []byte) []interface{} {
	str := string(byteArray)

  str = strings.Trim(str, "[]")
  numbers := strings.Split(str, ",")

  var result []interface{}

  for _, num := range numbers {
    // Try converting to float first
    if f, err := strconv.ParseFloat(num, 64); err == nil {
      result = append(result, f)
    } else if i, err := strconv.Atoi(num); err == nil {
      // Convert to int if float conversion fails
      result = append(result, i)
    }
  }
	return result
}

func generateAddress() string {
	privateKey, err := crypto.GenerateKey()
  if err != nil {
    log.Fatal(err)
  }

  // privateKeyBytes := crypto.FromECDSA(privateKey)
  // fmt.Println("SAVE BUT DO NOT SHARE THIS (Private Key):", hexutil.Encode(privateKeyBytes))

  publicKey := privateKey.Public()
  publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
  if !ok {
    log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
  }

  // publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
  // fmt.Println("Public Key:", hexutil.Encode(publicKeyBytes)) 

  return crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
}
