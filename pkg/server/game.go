package server

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/richard-to/go-poker/pkg/poker"
)

// General actions
const actionDisconnect string = "disconnect"
const actionError string = "error"
const actionOnJoin string = "on-join"
const actionOnTakeSeat string = "on-take-seat"
const actionJoin string = "join"
const actionMuteVideo string = "mute-video"
const actionNewMessage string = "new-message"
const actionSendMessage string = "send-message"
const actionTakeSeat string = "take-seat"

// WebRTC Signaling actions
const actionOnReceiveSignal string = "on-receive-signal"
const actionSendSignal string = "send-signal"

// Game actions
const actionBet string = "bet"
const actionCall string = "call"
const actionCheck string = "check"
const actionFold string = "fold"
const actionOnHoleCards string = "on-hole-cards"
const actionRaise string = "raise"
const actionUpdateGame string = "update-game"

// Table settings
const roomId int = 1
const defaultChips int = 100
const defaultMinBet int = 2
const minPlayers int = 2
const numPlayers int = 6

const systemUsername string = "System"

// GameStage is an enum for the current round of betting
type GameStage int

// Stages of a game
const (
	Waiting GameStage = iota
	Preflop
	Flop
	Turn
	River
	Showdown
)

func (g GameStage) String() string {
	return [...]string{"Waiting", "Preflop", "Flop", "Turn", "River", "Showdown"}[g]
}

// BroadcastEvent is an event that is broadcasted to multiple clients.
//
// There are cases where we don't want to broadcast to everyone. In this scenario
// the exclude clients map can be used. This will prevent messages from being sent
// to the specified client ID
type BroadcastEvent struct {
	Event          Event
	ExcludeClients map[string]bool
}

// Event is a JSON message in the game loop.
type Event struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
}

// GameState is the current state of the poker game
type GameState struct {
	BettingRound *poker.BettingRound
	CurrentSeat  *poker.Seat
	Deck         poker.Deck
	PlayerMap    map[string]*poker.Player
	Stage        GameStage
	Table        poker.Table
}

// NewBroadcastEvent creates a new broadcast event that will send the message to all clients
func NewBroadcastEvent(e Event) BroadcastEvent {
	return BroadcastEvent{
		Event:          e,
		ExcludeClients: make(map[string]bool, 0),
	}
}

// DisconnectPlayer disconnects player from a client when a client has been disconnected.
//
// - When a client is disconnected, we will set the player to be computer controlled
// - If a hand has not started yet, make the seat available again
// - If the client is disconnected while it's their turn, the player will auto-fold or check
// - Not all clients will be sitting at the table
func DisconnectPlayer(c *Client) {
	player := poker.GetPlayerByID(&c.gameState.Table, c.seatID)
	if player != nil {
		player.IsHuman = false
		if c.gameState.Stage == Waiting {
			player.Status = poker.PlayerVacated
			c.hub.broadcast <- NewBroadcastEvent(createUpdateGameEvent(c, false))
		} else if c.gameState.Stage < Showdown {
			HandleComputerMove(c)
		}
	}

	// If a client does not have a username set, that means they haven't technically
	// joined the table yet. In that case we don't have to post a message.
	if c.username != "" {
		c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(
			systemUsername,
			fmt.Sprintf("%s has left the game.", c.username),
		))
	}
}

// ProcessEvent process event
func ProcessEvent(c *Client, e Event) {
	var err error
	if e.Action == actionJoin {
		err = HandleJoin(c, e.Params["username"].(string))
	} else if e.Action == actionSendMessage {
		err = HandleSendMessage(c, e.Params["username"].(string), e.Params["message"].(string))
	} else if e.Action == actionSendSignal {
		err = HandleSendSignal(
			c,
			e.Params["peerID"].(string),
			e.Params["streamID"].(string),
			e.Params["signalData"],
		)
	} else if e.Action == actionTakeSeat {
		err = HandleTakeSeat(c, e.Params["seatID"].(string))
	} else if e.Action == actionMuteVideo {
		err = HandleMuteVideo(c, e.Params["muted"].(bool))
	} else {
		// The remaining actions are turn dependent. The player can only act if it's their turn.
		if c.gameState.Stage < Preflop || c.gameState.Stage > River {
			err = fmt.Errorf("You cannot move during the %s stage", c.gameState.Stage.String())
		} else if c.gameState.CurrentSeat.Player.ID != c.seatID {
			err = fmt.Errorf("You cannot move out of turn")
		} else if e.Action == actionFold {
			err = HandleFold(c)
		} else if e.Action == actionCheck {
			err = HandleCheck(c)
		} else if e.Action == actionCall {
			err = HandleCall(c)
		} else if e.Action == actionRaise {
			raiseAmount := int(e.Params["value"].(float64))
			err = HandleRaise(c, raiseAmount)
		} else {
			err = fmt.Errorf("Unknown action encountered: %s", e.Action)
		}
	}

	if err != nil {
		HandlePlayerError(c, err)
	}
}

// HandlePlayerError handles player error (not system error)
func HandlePlayerError(c *Client, err error) error {
	c.send <- createErrorEvent(err)
	return nil
}

// HandleJoin handles join event
func HandleJoin(c *Client, username string) error {
	c.username = username

	c.send <- createOnJoinEvent(c.id, c.username)

	c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(
		systemUsername,
		fmt.Sprintf("%s joined the game.", c.username),
	))

	c.hub.broadcast <- NewBroadcastEvent(createUpdateGameEvent(c, false))
	return nil
}

// HandleSendMessage handles send message event
func HandleSendMessage(c *Client, username string, message string) error {
	c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(username, message))
	return nil
}

// HandleSendSignal handles WebRTC signaling messages
func HandleSendSignal(c *Client, recipientID string, streamID string, signalData interface{}) error {
	recipient, ok := c.hub.clients[recipientID]
	if ok == false {
		// Temporarily make this a noop until I can improve the error handling a bit
		// return fmt.Errorf("Recipient userID (%s) does not exist", recipientID)
		return nil
	}
	recipient.send <- createOnReceiveSignal(c.id, streamID, signalData)
	return nil
}

// HandleMuteVideo unmutes/mutes user
func HandleMuteVideo(c *Client, muted bool) error {
	c.muted = muted
	c.hub.broadcast <- NewBroadcastEvent(createUpdateGameEvent(c, false))
	return nil
}

// HandleTakeSeat takes a seat for the user
func HandleTakeSeat(c *Client, seatID string) error {
	if c.seatID != "" {
		return fmt.Errorf("You can only sit at one seat")
	}

	seat := c.gameState.Table.Seats

	var selectedPlayer *poker.Player
	for i := 0; i < seat.Len(); i++ {
		if seat.Player.ID == seatID {
			// It's possible that two players picked the same seat at the same time
			if seat.Player.Status > poker.PlayerVacated {
				return fmt.Errorf("Seat has already been taken")
			}
			selectedPlayer = seat.Player
			break
		}
		seat = seat.Next()
	}

	if selectedPlayer == nil {
		return fmt.Errorf("Invalid seat chosen")
	}

	dp := getUserBalanceByEmail(c.username)
	// Link user with player seat
	selectedPlayer.Name = c.username
	// selectedPlayer.Chips = defaultChips
	selectedPlayer.Chips = dp.Db_POT
	selectedPlayer.Status = poker.PlayerSittingOut
	selectedPlayer.IsHuman = true
	c.seatID = selectedPlayer.ID

	c.send <- createOnTakeSeatEvent(seatID, createClientSeatMap(c.hub.clients))

	// Try to start a new game if one hasn't started yet.
	if c.gameState.Stage == Waiting {
		StartNewHand(c.gameState)
		sendHoleCardEvents(c.hub.clients)
	}

	c.hub.broadcast <- NewBroadcastEvent(createUpdateGameEvent(c, false))

	return nil
}

// HandleFold folds
func HandleFold(c *Client) error {
	err := c.gameState.CurrentSeat.Player.Fold(c.gameState.BettingRound)
	if err != nil {
		return err
	}
	c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(
		systemUsername,
		fmt.Sprintf("%s folds.", c.gameState.CurrentSeat.Player.Name),
	))
	return GoToNextGameState(c)
}

// HandleCheck checks
func HandleCheck(c *Client) error {
	err := c.gameState.CurrentSeat.Player.Check(c.gameState.BettingRound)
	if err != nil {
		return err
	}
	c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(
		systemUsername,
		fmt.Sprintf("%s checks.", c.gameState.CurrentSeat.Player.Name),
	))
	return GoToNextGameState(c)
}

// HandleCall calls
func HandleCall(c *Client) error {
	err := c.gameState.CurrentSeat.Player.Call(&c.gameState.Table, c.gameState.BettingRound)
	if err != nil {
		return err
	}
	c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(
		systemUsername,
		fmt.Sprintf("%s calls.", c.gameState.CurrentSeat.Player.Name),
	))
	return GoToNextGameState(c)
}

// HandleRaise raises/bets
func HandleRaise(c *Client, raiseAmount int) error {
	// Determine whether we're betting or raising. A bet can be determined if call amount is 0,
	// which means no one has bet anything yet.
	actionLabel := "bets"
	if c.gameState.BettingRound.CallAmount > 0 {
		actionLabel = "raises to"
	}

	err := c.gameState.CurrentSeat.Player.Raise(&c.gameState.Table, c.gameState.BettingRound, raiseAmount)
	if err != nil {
		return err
	}

	c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(
		systemUsername,
		fmt.Sprintf("%s %s 🍺%d.", c.gameState.CurrentSeat.Player.Name, actionLabel, raiseAmount),
	))
	return GoToNextGameState(c)
}

// GoToNextGameState moves to the next game state
func GoToNextGameState(c *Client) error {
	// TODO: Check if this loop is needed still
	for {
		err := NextGameState(c)
		if err != nil {
			return err
		}
		if c.gameState.Stage == Waiting {
			break
		}
		if c.gameState.CurrentSeat.Player.Status == poker.PlayerActive && c.gameState.CurrentSeat.Player.Chips > 0 {
			break
		}
	}
	c.hub.broadcast <- NewBroadcastEvent(createUpdateGameEvent(c, false))

	HandleComputerMove(c)

	return nil
}

// HandleComputerMove makes as move (check/folder) for a player who has been disconnected
func HandleComputerMove(c *Client) {
	if c.gameState.Stage == Waiting || c.gameState.Stage == Showdown {
		return
	}
	if c.gameState.CurrentSeat.Player.IsHuman {
		return
	}

	if c.gameState.CurrentSeat.Player.CanFold(c.gameState.BettingRound) {
		HandleFold(c)
	} else if c.gameState.CurrentSeat.Player.CanCheck(c.gameState.BettingRound) {
		HandleCheck(c)
	}
}

// NextGameState gets the next game state
func NextGameState(c *Client) error {
	var err error

	g := c.gameState

	nextSeat := g.CurrentSeat.Next()
	for i := 0; i < g.CurrentSeat.Len(); i++ {
		if nextSeat == g.CurrentSeat {
			return fmt.Errorf("Next active seat not found. All players have folded")
		}
		if nextSeat.Player.Status != poker.PlayerActive {
			continue
		}
		if nextSeat.Player.HasFolded == false || nextSeat.Player.Chips == 0 {
			break
		}
		if nextSeat.Player == g.BettingRound.Raiser {
			break
		}
		nextSeat = nextSeat.Next()
	}

	g.CurrentSeat = nextSeat

	winnerByFold := poker.DetermineWinnerByFold(g.CurrentSeat)

	if winnerByFold != nil {
		c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(
			systemUsername,
			fmt.Sprintf("%s won the hand.", winnerByFold.Name),
		))
		poker.AwardPot(&g.Table, winnerByFold)
		StartNewHand(g)
		sendHoleCardEvents(c.hub.clients)
		c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(systemUsername, "Starting new hand."))
		return nil
	}

	if g.CurrentSeat.Player == g.BettingRound.Raiser {
		g.Stage++

		skipToShowdown := poker.SkipToShowdown(g.CurrentSeat)
		if skipToShowdown {
			if g.Stage < Turn {
				poker.DealFlop(&g.Deck, &g.Table)
				g.Stage = Flop
				c.hub.broadcast <- NewBroadcastEvent(createUpdateGameEvent(c, true))
				time.Sleep(3 * time.Second)
			}
			if g.Stage < River {
				poker.DealTurn(&g.Deck, &g.Table)
				g.Stage = Turn
				c.hub.broadcast <- NewBroadcastEvent(createUpdateGameEvent(c, true))
				time.Sleep(2 * time.Second)
			}
			if g.Stage < Showdown {
				poker.DealRiver(&g.Deck, &g.Table)
				g.Stage = River
				c.hub.broadcast <- NewBroadcastEvent(createUpdateGameEvent(c, true))
			}
			g.Stage = Showdown
		}

		if g.Stage == Flop {
			poker.DealFlop(&g.Deck, &g.Table)
			g.CurrentSeat, err = poker.GetNextActiveSeat(g.Table.Dealer)
			if err != nil {
				return err
			}
			g.BettingRound, err = poker.NewBettingRound(g.CurrentSeat, 0, g.Table.MinBet)
			if err != nil {
				return err
			}
			c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(systemUsername, "Dealing flop."))
		} else if g.Stage == Turn {
			poker.DealTurn(&g.Deck, &g.Table)
			g.CurrentSeat, err = poker.GetNextActiveSeat(g.Table.Dealer)
			if err != nil {
				return err
			}
			g.BettingRound, err = poker.NewBettingRound(g.CurrentSeat, 0, g.Table.MinBet)
			if err != nil {
				return err
			}
			c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(systemUsername, "Dealing turn."))
		} else if g.Stage == River {
			poker.DealRiver(&g.Deck, &g.Table)
			g.CurrentSeat, err = poker.GetNextActiveSeat(g.Table.Dealer)
			if err != nil {
				return err
			}
			g.BettingRound, err = poker.NewBettingRound(g.CurrentSeat, 0, g.Table.MinBet)
			if err != nil {
				return err
			}
			c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(systemUsername, "Dealing river."))
		} else if g.Stage == Showdown {
			c.hub.broadcast <- NewBroadcastEvent(createUpdateGameEvent(c, true))
			time.Sleep(2 * time.Second)
			DetermineWinners(c)
			time.Sleep(1 * time.Second)
			StartNewHand(g)
			sendHoleCardEvents(c.hub.clients)
			c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(systemUsername, "Starting new hand."))
		} else {
			return fmt.Errorf("Invalid game stage encountered: %s", g.Stage.String())
		}
	}

	return nil
}

// DetermineWinners determines who won the hand and awards chips to the winner
func DetermineWinners(c *Client) {
	g := c.gameState

	allWinningHands := poker.DetermineWinners(&g.Table)
	for i, winningHandsByPot := range allWinningHands {
		potText := "main pot"
		if i > 0 {
			if len(allWinningHands) == 2 {
				potText = "side pot"
			} else {
				potText = fmt.Sprintf("side pot %d", i)
			}
		}
		for _, ph := range winningHandsByPot {
			c.hub.broadcast <- NewBroadcastEvent(createNewMessageEvent(
				systemUsername,
				fmt.Sprintf(
					"%s wins 🍺%d %s with %s.",
					ph.Player.Name,
					ph.ChipsWon,
					potText,
					strings.ToLower(ph.Hand.Rank.String()),
				),
			))
		}
	}
}

// NewGameState creates a new game state
func NewGameState() *GameState {
	// Initialize vacated seats
	playerMap := make(map[string]*poker.Player)
	seats := poker.NewSeat(numPlayers)
	for i := 0; i < seats.Len(); i++ {
		seats.Player = &poker.Player{
			ID:      uuid.New().String(),
			Status:  poker.PlayerVacated,
			IsHuman: false,
		}
		playerMap[seats.Player.ID] = seats.Player
		seats = seats.Next()
	}

	return &GameState{
		BettingRound: nil,
		CurrentSeat:  seats,
		Deck:         poker.NewDeck(),
		PlayerMap:    playerMap,
		Stage:        Waiting,
		Table: poker.Table{
			MinBet: defaultMinBet,
			Pot:    poker.NewPot(),
			Seats:  seats,
		},
	}
}

// StartNewHand starts a new hand
func StartNewHand(g *GameState) {
	deck := poker.NewDeck()
	seats := g.Table.Seats

	// Reset player hands
	for i := 0; i < seats.Len(); i++ {
		seats.Player.HoleCards = [2]*poker.Card{}
		seats.Player.HasFolded = false
		seats = seats.Next()
	}

	// Get active players for the next game
	for i := 0; i < seats.Len(); i++ {
		// If a player is computer controlled, then vacate the seat
		if seats.Player.IsHuman == false {
			seats.Player.Name = ""
			seats.Player.Chips = 0
			seats.Player.Status = poker.PlayerVacated
		}

		if seats.Player.Status > poker.PlayerVacated {
			if seats.Player.Chips < defaultMinBet {
				seats.Player.Status = poker.PlayerSittingOut
			} else {
				seats.Player.Status = poker.PlayerActive
			}
		}
		seats = seats.Next()
	}

	activePlayerCount := poker.CountSeatsByPlayerStatus(seats, poker.PlayerActive)

	if activePlayerCount < minPlayers {
		// Change active player status to sitting out if we don't have enough players
		for i := 0; i < seats.Len(); i++ {
			if seats.Player.Status == poker.PlayerActive {
				seats.Player.Status = poker.PlayerSittingOut
			}
			seats = seats.Next()
		}
		g.Stage = Waiting
		g.Table = poker.Table{
			MinBet: defaultMinBet,
			Pot:    poker.NewPot(),
			Seats:  seats,
		}
		return
	}

	if g.Table.Dealer == nil {
		g.Table.Dealer = g.Table.Seats
	}

	dealer, err := poker.GetNextActiveSeat(g.Table.Dealer)
	if err != nil {
		panic(err)
	}

	smallBlind, err := poker.GetNextActiveSeat(dealer)
	if err != nil {
		panic(err)
	}

	// In a head to head match, the dealer is the small blind
	if activePlayerCount == 2 {
		smallBlind = dealer
	}

	bigBlind, err := poker.GetNextActiveSeat(smallBlind)
	if err != nil {
		panic(err)
	}

	table := poker.Table{
		BigBlind:   bigBlind,
		Dealer:     dealer,
		MinBet:     defaultMinBet,
		Pot:        poker.NewPot(),
		Seats:      seats,
		SmallBlind: smallBlind,
	}

	poker.DealHands(&deck, &table)

	currentSeat, err := poker.GetNextActiveSeat(bigBlind)
	if err != nil {
		panic(err)
	}

	preflopRound, err := poker.NewBettingRound(currentSeat, table.MinBet, table.MinBet)
	if err != nil {
		panic(err)
	}

	poker.TakeSmallBlind(&table, preflopRound)
	poker.TakeBigBlind(&table, preflopRound)

	g.BettingRound = preflopRound
	g.CurrentSeat = currentSeat
	g.Deck = deck
	g.Stage = Preflop
	g.Table = table
}

// GetActions gets the actions available to active player
func GetActions(g *GameState) []string {
	var actions []string

	if g.CurrentSeat.Player.CanFold(g.BettingRound) {
		actions = append(actions, actionFold)
	}

	if g.CurrentSeat.Player.CanCheck(g.BettingRound) {
		actions = append(actions, actionCheck)
	}

	if g.CurrentSeat.Player.CanCall(g.BettingRound) {
		actions = append(actions, actionCall)
	}

	if g.CurrentSeat.Player.CanRaise(g.BettingRound) {
		actions = append(actions, actionRaise)
	}
	return actions
}

func sendHoleCardEvents(clients map[string]*Client) {
	for _, c := range clients {
		if p, ok := c.gameState.PlayerMap[c.seatID]; ok {
			c.send <- createPlayerHoleCardsEvent(c.seatID, p.HoleCards)
		}
	}
}

func createOnJoinEvent(userID string, username string) Event {
	return Event{
		Action: actionOnJoin,
		Params: map[string]interface{}{
			"userID":   userID,
			"username": username,
		},
	}
}

// //roomID add
// func createOnJoinEvent2(roomID string,userID string, username string) Event {
// 	return Event{
// 		Action: actionOnJoin,
// 		Params: map[string]interface{}{
// 			"roomID":   roomID,
// 			"userID":   userID,
// 			"username": username,
// 		},
// 	}
// }

func createOnTakeSeatEvent(seatID string, clientSeatMap map[string]string) Event {
	return Event{
		Action: actionOnTakeSeat,
		Params: map[string]interface{}{
			"seatID":        seatID,
			"clientSeatMap": clientSeatMap,
		},
	}
}

func createOnReceiveSignal(peerID string, streamID string, signalData interface{}) Event {
	return Event{
		Action: actionOnReceiveSignal,
		Params: map[string]interface{}{
			"peerID":     peerID,
			"streamID":   streamID,
			"signalData": signalData,
		},
	}
}

func createUpdateGameEvent(c *Client, showCards bool) Event {
	var actionBar map[string]interface{}

	g := c.gameState

	players := make([]map[string]interface{}, 0)
	seats := g.Table.Seats

	mutedSeatMap := createMutedSeatMap(c.hub.clients)

	if g.Stage == Waiting {
		// Players data
		for i := 0; i < seats.Len(); i++ {
			players = append(players, map[string]interface{}{
				"chips":      seats.Player.Chips,
				"chipsInPot": nil,
				"hasFolded":  seats.Player.HasFolded,
				"holeCards":  [2]*poker.Card{},
				"id":         seats.Player.ID,
				"isActive":   false,
				"isDealer":   false,
				"muted":      mutedSeatMap[seats.Player.ID],
				"name":       seats.Player.Name,
				"status":     seats.Player.Status.String(),
			})
			seats = seats.Next()
		}

		// Actions data
		actionBar = map[string]interface{}{
			"actions":        []string{},
			"callAmount":     0,
			"chipsInPot":     0,
			"maxRaiseAmount": 0,
			"minBetAmount":   0,
			"minRaiseAmount": 0,
			"totalChips":     0,
		}
	} else {
		// Players data
		activePlayer := g.CurrentSeat.Player
		for i := 0; i < seats.Len(); i++ {
			holeCards := [2]*poker.Card{}
			if showCards && seats.Player.HasFolded == false {
				holeCards = seats.Player.HoleCards

				if holeCards[0] == nil || holeCards[1] == nil {
				} else {
					setgameLogByEmail(seats.Player.Chips, g.BettingRound.Bets[seats.Player.ID], strconv.FormatBool(seats.Player.HasFolded),
						string(holeCards[0].Symbol()), string(holeCards[1].Symbol()),
						seats.Player.ID, activePlayer.ID, g.Table.Dealer.Player.ID, seats.Player.Name, seats.Player.Status.String())
				}
			}
			players = append(players, map[string]interface{}{
				"chips":      seats.Player.Chips,
				"chipsInPot": g.BettingRound.Bets[seats.Player.ID],
				"hasFolded":  seats.Player.HasFolded,
				"holeCards":  holeCards,
				"id":         seats.Player.ID,
				"isActive":   seats.Player.ID == activePlayer.ID,
				"isDealer":   seats.Player.ID == g.Table.Dealer.Player.ID,
				"muted":      mutedSeatMap[seats.Player.ID],
				"name":       seats.Player.Name,
				"status":     seats.Player.Status.String(),
			})
			// setUserBalanceByEmail(username string, chgPOT string) (cnt int64) {
			setUserBalanceByEmail(seats.Player.Name, seats.Player.Chips)

			seats = seats.Next()
		}

		// Actions data

		// If the player does not have enough chips to meet the call amount, then set the max raise
		// to the player's remaining chips/
		callRemainingAmount := g.BettingRound.CallAmount - g.BettingRound.Bets[activePlayer.ID]
		maxRaiseAmount := activePlayer.Chips - callRemainingAmount
		if maxRaiseAmount < 0 {
			maxRaiseAmount = activePlayer.Chips
		}

		// If the min raise amount is less than the max raise amount, then use the max raise
		// as the min raise. Essentially this means that the play must go all in if they were
		// to raise.
		minRaiseAmount := g.BettingRound.RaiseByAmount
		if minRaiseAmount > maxRaiseAmount {
			minRaiseAmount = maxRaiseAmount
		}

		actionBar = map[string]interface{}{
			"actions":        GetActions(g),
			"callAmount":     g.BettingRound.CallAmount,
			"chipsInPot":     g.BettingRound.Bets[activePlayer.ID],
			"maxRaiseAmount": maxRaiseAmount,
			"minBetAmount":   g.Table.MinBet,
			"minRaiseAmount": minRaiseAmount,
			"seatID":         activePlayer.ID,
			"totalChips":     activePlayer.Chips,
		}
	}

	// Table data
	table := map[string]interface{}{
		"flop":  g.Table.Flop,
		"pot":   g.Table.Pot.GetTotal(),
		"river": g.Table.River,
		"turn":  g.Table.Turn,
	}

	return Event{
		Action: actionUpdateGame,
		Params: map[string]interface{}{
			"actionBar":     actionBar,
			"clientSeatMap": createClientSeatMap(c.hub.clients),
			"players":       players,
			"stage":         g.Stage.String(),
			"table":         table,
		},
	}
}

func createNewMessageEvent(username string, message string) Event {
	return Event{
		Action: actionNewMessage,
		Params: map[string]interface{}{
			"id":       uuid.New().String(),
			"message":  message,
			"username": username,
		},
	}
}

func createErrorEvent(err error) Event {
	return Event{
		Action: actionError,
		Params: map[string]interface{}{
			"error": err.Error(),
		},
	}
}

func createPlayerHoleCardsEvent(seatID string, c [2]*poker.Card) Event {
	return Event{
		Action: actionOnHoleCards,
		Params: map[string]interface{}{
			"holeCards": c,
			"seatID":    seatID,
		},
	}
}

func createClientSeatMap(clients map[string]*Client) map[string]string {
	clientSeatMap := make(map[string]string)
	for _, c := range clients {
		clientSeatMap[c.id] = c.seatID
	}
	return clientSeatMap
}

func createMutedSeatMap(clients map[string]*Client) map[string]bool {
	mutedSeatMap := make(map[string]bool)
	for _, c := range clients {
		if c.seatID != "" {
			mutedSeatMap[c.seatID] = c.muted
		}
	}
	return mutedSeatMap
}

//#############################
//#############################
//#############################
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

type dbPlayer struct {
	Db_Id           int
	Db_email        string
	Db_c4ei_addr    string
	Db_c4ei_balance string
	Db_POT          int
}

func getDBconnStr() (constr string) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// fmt.Println("env [HOST]:", os.Getenv("HOST"))
	// fmt.Println("username:", username)
	return os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASS") + "@tcp(" + os.Getenv("HOST") + ":3306)/" + os.Getenv("DB_DATABASE")
}

func setgameLogByEmail(chips int, chipsInPot int, hasFolded string, holeCards1 string, holeCards2 string,
	id string, isActive string, isDealer string, name string, status string) (cnt int64) {
	var constr = getDBconnStr()
	conn, err := sql.Open("mysql", constr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var strsql = "insert into gamelog (chips, chipsInPot, hasFolded, holeCards1, holeCards2, id, isActive, isDealer, name, status) values ('" + strconv.Itoa(int(chips)) + "','" + strconv.Itoa(int(chipsInPot)) + "','" + hasFolded + "','" + holeCards1 + "','" + holeCards2 + "','" + id + "','" + isActive + "','" + isDealer + "','" + name + "','" + status + "')"
	// log.Printf("strsql: %v", strsql)
	result, err := conn.Exec(strsql)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	nRow, err := result.RowsAffected() //RowsAffected() 함수를 통해 update 한 갯수를 확인한다.
	// fmt.Println("update count : ", nRow +" / username : " )
	// fmt.Sprintf("username : %s update count : %d.", username, nRow)
	conn.Close()
	return nRow
}
func setUserBalanceByEmail(username string, chgPOT int) (cnt int64) {
	var constr = getDBconnStr()
	conn, err := sql.Open("mysql", constr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	result, err := conn.Exec("update game_user set pot=" + strconv.Itoa(int(chgPOT)) + ", last_reg=now() where c4ei_addr='" + username + "' ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	nRow, err := result.RowsAffected() //RowsAffected() 함수를 통해 update 한 갯수를 확인한다.
	// fmt.Println("update count : ", nRow +" / username : " )
	fmt.Sprintf("username : %s update count : %d.", username, nRow)
	conn.Close()
	return nRow
}

func getUserBalanceByEmail(username string) (dp dbPlayer) {

	var constr = getDBconnStr()
	db, err := sql.Open("mysql", constr)
	checkError(err)
	defer db.Close()

	// fmt.Println("connect success")
	var _id, _pot int
	var _c4ei_addr, _c4ei_balance string //_email,
	rows, err := db.Query("SELECT id, c4ei_addr, c4ei_balance, pot FROM game_user a WHERE a.c4ei_addr='" + username + "'")
	checkError(err)
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&_id, &_c4ei_addr, &_c4ei_balance, &_pot)
		checkError(err)
		dp.Db_Id = _id
		// dp.Db_email = _email
		dp.Db_c4ei_addr = _c4ei_addr
		dp.Db_c4ei_balance = _c4ei_balance
		dp.Db_POT = _pot

		log.Println(dp.Db_POT)
		// fmt.Println("rows", _pot)
	}
	return dp
}

//#############################
//#############################
//#############################
