package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

var (
	database       = &pgx.Conn{}
	fullRollLength = float64(0)
)

type RugOutput struct {
	RollID     int       `json:"roll_id"`
	RollLength float64   `json:"length"`
	Plan       RugBucket `json:"plan"`
}

type NextRequest struct {
	RollLength  float64 `json:"roll_length"`
	IncludeRush bool    `json:"include_rush"`
}

type RugBlocks struct {
	ThreeFeet int
	SevenFeet int
}

func PrintJSON(obj interface{}, w http.ResponseWriter) {
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.SetIndent("", " ")
	err := jsonEncoder.Encode(obj)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func UseRugBlocks(length float64) (blocks RugBlocks, remainder float64) {
	// we're going to take 3ft sections until the remainder length is divisible by 7,
	// or if we're taking over half of the rug, area,
	remainder = length
	for remainder >= 3 {
		if remainder/7 == math.Floor(remainder/7) {
			blocks.SevenFeet = int(remainder / 7)
			remainder = 0
		}
		remainder = remainder - 3
		blocks.ThreeFeet++
	}
	return blocks, remainder
}

type RugBucket = []*RugItem

type RugItem struct {
	ComponentID   int       `json:"component_id"`
	ComponentSize string    `json:"component_size"`
	OrderDate     time.Time `json:"order_date"`
	Position      int       `json:"position"`
	Sku           string    `json:"sku"`
	Rush          bool      `json:"rush"`
}

// GetRugBucketOfSize todo
func GetRugBucketOfSize(tx pgx.Tx, ctx context.Context, size string, max int, includeRush bool) (RugBucket, error) {
	q := `select component.id, component.size, "order".order_date, line_item.sku, line_item.rush
		FROM component, line_item, "order"
		WHERE component.line_item_id = line_item.id
		 AND line_item.order_id = "order".id
		 AND component.status = 'Pending'
		 AND  "order".cancelled = 'false'
		 AND component.size = $1
		 AND (line_item.rush = 'false' OR line_item.rush = $2)
		ORDER BY line_item.rush DESC, "order".order_date ASC
		LIMIT $3;`
	rb := RugBucket{}

	rows, err := tx.Query(ctx, q, size, includeRush, max)
	if err != nil {
		return rb, err
	}
	for rows.Next() {
		ri := RugItem{}
		err := rows.Scan(&ri.ComponentID, &ri.ComponentSize, &ri.OrderDate, &ri.Sku, &ri.Rush)
		if err != nil {
			return rb, err
		}
		rb = append(rb, &ri)
	}
	return rb, err
}

// HighestPriority is a function that will tell you if the priority of the rugs matches the
// order in which they were passed, or if they need to be flipped (false)
func HighestPriority(l, r *RugItem) bool {

	// Compare
	// if time.Now().Add(-time.Hour*24).After(r.OrderDate) || time.Now().Add(-time.Hour*24).After(l.OrderDate) {
	// 	return r.OrderDate.After(l.OrderDate)
	// }

	// If both are the same type, then compare dates to find the newest item
	if l.Rush == r.Rush {
		return r.OrderDate.After(l.OrderDate)
	}
	if l.Rush {
		return true
	}
	return false
}

// NextHandler will give a printer the next print job.
func NextHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	includeRush := true
	var err error
	newDecorder := json.NewDecoder(r.Body)
	nextReq := NextRequest{}
	newDecorder.DisallowUnknownFields()
	err = newDecorder.Decode(&nextReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		PrintJSON(map[string]string{
			"error": "request not valid json",
		}, w)
		return
	}
	// Since we cannot print on rug fragments, that is wasted material, so lets take the floor
	// of that value to make the numbers easier.
	nextReq.RollLength = math.Floor(nextReq.RollLength)

	// The smallest length rug we can print is 3ft, so we cannot accept rug fragments less then 3ft
	if nextReq.RollLength <= 3 {
		w.WriteHeader(http.StatusNotAcceptable)
		PrintJSON(map[string]string{
			"error": "roll length is required and must be greater then 3ft",
		}, w)
		return
	}
	// lets see what kind of volume we have in here, this will allow us to pick the rugs to order and choose area or runner rugs
	// this is a quick inventory heuristic to determine what's available.
	tx, err := database.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		PrintJSON(map[string]string{
			"error": "could not connect to server",
		}, w)
		return
	}
	largePlots, err := GetRugBucketOfSize(tx, ctx, "5x7", int(math.Floor(nextReq.RollLength/7))+1, includeRush)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		PrintJSON(map[string]string{
			"error": "database error " + err.Error(),
		}, w)
		return
	}
	runnerPlots, err := GetRugBucketOfSize(tx, ctx, "2.5x7", int(math.Ceil(nextReq.RollLength/7))*2, includeRush)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		PrintJSON(map[string]string{
			"error": "database error " + err.Error(),
		}, w)
		return
	}
	smallPlots, err := GetRugBucketOfSize(tx, ctx, "3x5", int(math.Ceil(nextReq.RollLength/3)), includeRush)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		PrintJSON(map[string]string{
			"error": "database error " + err.Error(),
		}, w)
		return
	}

	// if piece supplied is a fragment, let's see if we can calculate the ramaining space. Ideality this would be a heuristic value
	printPlots := RugBucket{}
	// we don't have a perfect fit of rug blocks, so let's do the algorithm presented in the project.
	// take high priority items first, and then order by date. We can check from the 3 queues for this.
	largeIterator := 0
	runnerIterator := 0
	smallIterator := 0
	queuedLength := 0
	position := 1 // Starts at 1

	for {
		var top *RugItem

		if largeIterator < len(largePlots) && int(nextReq.RollLength)-queuedLength >= 7 {
			top = largePlots[largeIterator]
		}

		if smallIterator < len(smallPlots) && int(nextReq.RollLength)-queuedLength >= 3 {
			if top != nil {
				// if top is not null, let's compare with the other rug item and see if we can beat it.
				if HighestPriority(smallPlots[smallIterator], top) {
					top = smallPlots[smallIterator]
				}
			} else {
				top = smallPlots[smallIterator]
			}
		}
		// let's see if there are two
		if runnerIterator < len(runnerPlots) && int(nextReq.RollLength)-queuedLength >= 7 {
			if top != nil {
				// if top is not null, let's compare with the other rug item and see if we can beat it.
				if HighestPriority(runnerPlots[runnerIterator], top) {
					// Lets check if there is a second runner in the queue to add to the print Queue
					if runnerIterator+1 < len(runnerPlots) {
						//there is another runner, lets add both to the print queue with the same position
						//this is our special case:
						runnerPlots[runnerIterator].Position = position
						runnerPlots[runnerIterator+1].Position = position
						printPlots = append(printPlots, runnerPlots[runnerIterator], runnerPlots[runnerIterator+1])
						runnerIterator = runnerIterator + 2
						position = position + 1
						queuedLength = queuedLength + 7
						continue
					}
					top = runnerPlots[runnerIterator]
				}
			} else {
				// Lets check if there is a second runner in the queue to add to the print Queue
				if runnerIterator+1 < len(runnerPlots) {
					//there is another runner, lets add both to the print queue with the same position
					//this is our special case:
					runnerPlots[runnerIterator].Position = position
					runnerPlots[runnerIterator+1].Position = position
					printPlots = append(printPlots, runnerPlots[runnerIterator], runnerPlots[runnerIterator+1])
					runnerIterator = runnerIterator + 2
					position = position + 1
					queuedLength = queuedLength + 7
					continue
				}
				top = runnerPlots[runnerIterator]
			}
		}
		if top == nil {
			// there was no new rugs to grab, we have reached the end of the roll or the queue
			break
		}

		if top.ComponentSize == "5x7" {
			queuedLength = queuedLength + 7
			largeIterator = largeIterator + 1
		}
		if top.ComponentSize == "2.5x7" {
			queuedLength = queuedLength + 7
			runnerIterator = runnerIterator + 1
		}
		if top.ComponentSize == "3x5" {
			queuedLength = queuedLength + 3
			smallIterator = smallIterator + 1
		}
		top.Position = position
		position = position + 1
		printPlots = append(printPlots, top)
	}
	sql := `UPDATE component
	SET status = 'Printing'
	WHERE id in ($1`
	componentIds := []interface{}{}
	for i, p := range printPlots {
		componentIds = append(componentIds, p.ComponentID)
		if i != 0 {
			sql = sql + ",$" + strconv.Itoa(i+1)
		}
	}
	// We have a result here, so let's mark these components as scheduled to print.
	tx.Exec(ctx, sql+")", componentIds...)
	_ = tx.Commit(ctx)

	/* print the roll to the user*/
	result := RugOutput{
		Plan:       printPlots,
		RollLength: float64(queuedLength),
	}
	PrintJSON(result, w)
	return
}

func main() {
	var err error
	database, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close(context.Background())

	r := mux.NewRouter()
	r.HandleFunc("/next", NextHandler)
	http.Handle("/", r)

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*1, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	// Add your routes as needed

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
