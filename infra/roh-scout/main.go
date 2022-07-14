package go_scout

import (
	secretsmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const zoneAvailURLPrefix = "https://www.roh.org.uk/api/proxy/TXN/Performances/ZoneAvailabilities?performanceIds=%s"
const seatRetrievURLPrefix = "https://www.roh.org.uk/api/proxy/TXN/Performances/%s/Seats?constituentId=0&modeOfSaleId=9&performanceId=%s"
const visitURL = "https://www.roh.org.uk/seatmap?performanceId=%s"
const availableSeatCode = 0

type ZoneAvailability struct {
	Zone struct {
		Id          int
		Description string
	}
	PerformanceId  int
	AvailableCount int
}

type Seat struct {
	Id           int
	ZoneId       int
	SeatRow      string
	SeatNumber   string
	SeatStatusId int
}

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func findZoneId(performanceId string, seats string) (int, error) {
	log.Println("looking for the zoneID...")
	zoneAvailURL := fmt.Sprintf(zoneAvailURLPrefix, performanceId)

	r, err := http.Get(zoneAvailURL)
	if err != nil {
		return 0, err
	}

	var zAvs []ZoneAvailability

	err = json.NewDecoder(r.Body).Decode(&zAvs)
	if err != nil {
		return 0, err
	}

	for _, zAv := range zAvs {
		if zAv.Zone.Description == seats {
			return zAv.Zone.Id, nil
		}
	}

	return 0, errors.New(fmt.Sprintf("seats not found: %s", seats))
}

func findSeatsInZoneId(performanceId string, zoneId int) ([]Seat, error) {
	seatRetrievalUrl := fmt.Sprintf(seatRetrievURLPrefix, performanceId, performanceId)

	r, err := http.Get(seatRetrievalUrl)
	if err != nil {
		return []Seat{}, err
	}

	var seats []Seat

	err = json.NewDecoder(r.Body).Decode(&seats)
	if err != nil {
		return []Seat{}, err
	}

	var availableSeats []Seat
	for _, seat := range seats {
		if seat.ZoneId == zoneId && seat.SeatStatusId == availableSeatCode {
			availableSeats = append(availableSeats, seat)
		}
	}

	return availableSeats, nil
}

func sendTelegramMessage(ctx context.Context, m string) error {
	chatID, err := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
	if err != nil {
		return err
	}

	client, err := secretsmanager.NewClient(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: os.Getenv("TELEGRAM_SECRET_ID") + "/versions/latest",
	}

	secret, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return err
	}

	bot, err := tgbotapi.NewBotAPI(string(secret.GetPayload().Data))
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(int64(chatID), m)

	if _, err := bot.Send(msg); err != nil {
		return err
	}

	log.Printf("the message has been sent to chatID: %d. Message: %s", chatID, m)

	return nil
}

func putSavedNotifications(ctx context.Context, savedNotifications []Seat, performanceId string) error {
	log.Println("putting saved notifications...")

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	obj := client.Bucket(os.Getenv("BUCKET")).Object(performanceId)

	w := obj.NewWriter(ctx)
	if err != nil {
		return err
	}

	err = json.NewEncoder(w).Encode(savedNotifications)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	log.Printf("file %s written", performanceId)
	return nil
}

func getSavedNotifications(ctx context.Context, performanceId string) ([]Seat, error) {
	log.Println("retrieving saved notifications...")

	client, err := storage.NewClient(ctx)
	if err != nil {
		return []Seat{}, err
	}
	defer func() {
		_ = client.Close()
	}()

	obj := client.Bucket(os.Getenv("BUCKET")).Object(performanceId)

	if _, err = obj.Attrs(ctx); err != nil {
		return []Seat{}, nil
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		return []Seat{}, err
	}

	var savedNotifications []Seat
	err = json.NewDecoder(r).Decode(&savedNotifications)
	if err != nil {
		return []Seat{}, err
	}

	err = r.Close()
	if err != nil {
		return []Seat{}, err
	}

	log.Println("saved notification seats retrieved")
	return savedNotifications, nil
}

func findTicket(ctx context.Context, performanceId string) error {
	seats := os.Getenv("SEATS")

	savedNotifications, err := getSavedNotifications(ctx, performanceId)
	if err != nil {
		return err
	}

	ticketsWanted, err := strconv.Atoi(os.Getenv("TICKETS_WANTED"))
	if err != nil {
		return err
	}

	zoneId, err := findZoneId(performanceId, seats)
	if err != nil {
		return err
	}

	log.Printf("zoneId found: %d", zoneId)

	availSeats, err := findSeatsInZoneId(performanceId, zoneId)
	if err != nil {
		return err
	}

	if len(availSeats) >= ticketsWanted {
		msg := "The following seats are available:\n"
		for _, seat := range availSeats {
			msg += fmt.Sprintf("SeatRow: %s, SeatNumber: %s\n", seat.SeatRow, seat.SeatNumber)
		}
		msg += "book: " + fmt.Sprintf(visitURL, performanceId)

		log.Println(msg)

		if reflect.DeepEqual(savedNotifications, availSeats) {
			log.Println("already notified for the tickets, skipping sending telegram...")
		} else {
			if err := sendTelegramMessage(ctx, msg); err != nil {
				return err
			}
			if err = putSavedNotifications(ctx, availSeats, performanceId); err != nil {
				return err
			}
		}
	} else {
		log.Println("no available seats found")
	}

	return nil
}

func Run(ctx context.Context, _ PubSubMessage) error {
	performanceIds := strings.Split(os.Getenv("PERFORMANCE_IDS"), ",")

	for _, performanceId := range performanceIds {
		log.Printf("running performanceId: %s", performanceId)
		if err := findTicket(ctx, performanceId); err != nil {
			return err
		}
	}

	log.Println("execution completed")
	return nil
}
