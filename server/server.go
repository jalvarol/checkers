package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

type PieceState struct {
	IsOccupied bool   `json:"isOccupied"`
	IsKing     bool   `json:"isKing"`
	Color      string `json:"color"`
}

type Board map[string]PieceState

type GameState struct {
	Board  Board  `json:"board"`
	Turn   string `json:"turn"`
	Status string `json:"status"`
	Winner string `json:"winner"`
}

var game GameState

func newGame() {
	board := Board{}

	addPiece := func(position string, color string) {
		board[position] = PieceState{
			IsOccupied: true,
			IsKing:     false,
			Color:      color,
		}
	}

	for _, pos := range []string{"A1", "A3", "A5", "A7", "B2", "B4", "B6", "B8", "C1", "C3", "C5", "C7"} {
		addPiece(pos, "red")
	}

	for _, pos := range []string{"F2", "F4", "F6", "F8", "G1", "G3", "G5", "G7", "H2", "H4", "H6", "H8"} {
		addPiece(pos, "black")
	}

	game = GameState{
		Board:  board,
		Turn:   "red",
		Status: "ongoing",
	}
}

func checkWinner() {
	redPieces := 0
	blackPieces := 0

	for _, piece := range game.Board {
		if !piece.IsOccupied {
			continue
		}
		switch piece.Color {
		case "red":
			redPieces++
		case "black":
			blackPieces++
		}
	}

	if redPieces == 0 {
		game.Status = "won"
		game.Winner = "black"
	} else if blackPieces == 0 {
		game.Status = "won"
		game.Winner = "red"
	} else {
		game.Status = "ongoing"
		game.Winner = ""
	}
}

func captureMove(source, destination string) (bool, string) {
	sourceCol := source[0]
	sourceRow, _ := strconv.Atoi(string(source[1]))

	destCol := destination[0]
	destRow, _ := strconv.Atoi(string(destination[1]))

	midCol := byte((int(sourceCol) + int(destCol)) / 2)
	midRow := (sourceRow + destRow) / 2
	midPos := string([]byte{midCol}) + strconv.Itoa(midRow)

	if midPiece, ok := game.Board[midPos]; ok && midPiece.IsOccupied && midPiece.Color != game.Turn {
		return true, midPos
	}

	return false, ""
}

func checkPromotion(destination string, color string) bool {
	destRow, _ := strconv.Atoi(string(destination[1]))
	if (color == "red" && destRow == 8) || (color == "black" && destRow == 1) {
		return true
	}
	return false
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})
	r.Use(c.Handler)

	newGame()

	r.Get("/game", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(game)
	})

	r.Post("/game/new", func(w http.ResponseWriter, r *http.Request) {
		newGame()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "New game started.",
			"board":   game.Board,
			"turn":    game.Turn,
		})
	})

	r.Post("/game/move", func(w http.ResponseWriter, r *http.Request) {
		var move struct {
			Source      string `json:"source"`
			Destination string `json:"destination"`
		}
		if err := json.NewDecoder(r.Body).Decode(&move); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		source, sourceOk := game.Board[move.Source]
		if !sourceOk || !source.IsOccupied || source.Color != game.Turn {
			http.Error(w, "Invalid source or not your turn", http.StatusBadRequest)
			return
		}

		dest, destOk := game.Board[move.Destination]
		if destOk && dest.IsOccupied {
			http.Error(w, "Destination already occupied", http.StatusBadRequest)
			return
		}

		isCapture, capturedPos := captureMove(move.Source, move.Destination)
		if isCapture {
			capturedPiece := game.Board[capturedPos]
			capturedPiece.IsOccupied = false
			game.Board[capturedPos] = capturedPiece
		}

		isPromotion := checkPromotion(move.Destination, source.Color)
		if isPromotion {
			source.IsKing = true
		}

		source.IsOccupied = false
		game.Board[move.Source] = source

		dest = PieceState{IsOccupied: true, IsKing: source.IsKing, Color: source.Color}
		game.Board[move.Destination] = dest

		if game.Turn == "red" {
			game.Turn = "black"
		} else {
			game.Turn = "red"
		}

		checkWinner()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":      "Move successful.",
			"board":        game.Board,
			"turn":         game.Turn,
			"captured":     isCapture,
			"captured_pos": capturedPos,
			"promoted":     isPromotion,
			"status":       game.Status,
			"winner":       game.Winner,
		})
	})

	r.Get("/game/check-winner", func(w http.ResponseWriter, r *http.Request) {
		checkWinner()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": game.Status,
			"winner": game.Winner,
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	http.ListenAndServe(":6969", r)
}
