import React, { useState, useEffect } from 'react';
import './App.css';

interface PieceState {
  isOccupied: boolean;
  isKing: boolean;
  color: string; // "red" or "black"
}

interface BoardState {
  [key: string]: PieceState;
}

interface GameState {
  board: BoardState;
  turn: string;
  status: string;
  winner: string | null;
  captured?: boolean;
  captured_pos?: string;
  promoted?: boolean;
}

function App() {
  const [gameState, setGameState] = useState<GameState | null>(null);
  const [draggingFrom, setDraggingFrom] = useState<string | null>(null);

  useEffect(() => {
    const fetchGameState = async () => {
      const response = await fetch('http://localhost:6969/game');
      const data: GameState = await response.json();
      setGameState(data);
    };

    fetchGameState();
  }, []);

  const handleDragStart = (position: string) => {
    setDraggingFrom(position);
  };

  const handleDrop = async (toPosition: string) => {
    if (draggingFrom && gameState) {
      const movePayload = {
        source: draggingFrom,
        destination: toPosition,
      };

      const response = await fetch('http://localhost:6969/game/move', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(movePayload),
      });

      if (response.ok) {
        const updatedGameState: GameState = await response.json();
        setGameState(updatedGameState);
      }

      setDraggingFrom(null);
    }
  };

  const startNewGame = async () => {
    const response = await fetch('http://localhost:6969/game/new', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    });

    if (response.ok) {
      const newGameState: GameState = await response.json();
      setGameState(newGameState);
    }
  };

  const renderSquare = (row: number, col: number) => {
    const letters = 'ABCDEFGH';
    const position = `${letters[col]}${row + 1}`;
    const square = gameState?.board[position];

    let content = '';
    let draggable = false;

    if (square && square.isOccupied) {
      content = square.isKing ? (square.color === 'red' ? '♔' : '♚') : (square.color === 'red' ? '🔴' : '⚫');
      draggable = square.color === gameState?.turn;
    }

    return (
      <div
        className={`square ${((row + col) % 2 === 0) ? 'light' : 'dark'}`}
        key={position}
        draggable={draggable}
        onDragStart={() => handleDragStart(position)}
        onDragOver={(e) => e.preventDefault()} 
        onDrop={() => handleDrop(position)}
      >
        {content}
      </div>
    );
  };

  // Render an 8x8 board
  const renderBoard = () => {
    const board = [];
    for (let row = 7; row >= 0; row--) {
      const rowElements = [];
      for (let col = 0; col < 8; col++) {
        rowElements.push(renderSquare(row, col));
      }
      board.push(
        <div className="row" key={row}>
          {rowElements}
        </div>
      );
    }

    return (
      <div className="board">
        {board}
      </div>
    );
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>Checkers Game</h1>
        <button className="new-game-button" onClick={startNewGame}>
          New Game
        </button>
        {renderBoard()}
        {gameState && (
          <div>
            <p>Current Turn: {gameState.turn}</p>
            <p>Game Status: {gameState.status}</p>
            {gameState.captured && <p>Captured a piece at: {gameState.captured_pos}</p>}
            {gameState.promoted && <p>A piece has been promoted to King!</p>}
            {gameState.winner && <p>Winner: {gameState.winner}</p>}
          </div>
        )}
      </header>
    </div>
  );
}

export default App;
